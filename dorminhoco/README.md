# Problema 2 — Jogo do Dorminhoco

Implementação do jogo de cartas Dorminhoco com N jogadores dispostos em
**anel**, cada jogador como uma **goroutine** comunicando-se com
vizinhos via **channels**.

## Regras

- Cada jogador recebe uma mão de 3 cartas; o objetivo é formar uma
  **trinca** (3 cartas com o mesmo rank).
- A cada turno, o jogador descarta uma carta para o vizinho da
  **esquerda** e recebe uma carta do vizinho da **direita**.
- Quando alguém forma uma trinca, ele **bate**; os demais devem reagir
  o mais rápido possível — o **último a reagir perde a rodada**.

## Estrutura

```
dorminhoco/
├── cmd/
│   └── game/        ponto de entrada
├── internal/
│   ├── models/      Card, Deck, Player, Game
│   ├── metrics/     relatório por rodada
│   ├── logger/      logging com timestamp
│   └── utils/       utilitários (delays aleatórios)
├── results/         saídas das execuções
└── go.mod
```

A arquitetura segue exatamente o mesmo padrão da pasta `filosofos/`:
pacotes em `internal/`, comando em `cmd/<algo>/main.go`.

## Topologia de canais

```
   ... -> Player(i-1) -> Player(i) -> Player(i+1) -> ...
                  cardCh[i-1]   cardCh[i]   cardCh[i+1]
```

- Cada `cardCh[i]` é um `chan Card` com **buffer 1** (uma carta em
  trânsito por aresta).
- Player(i) **envia** em `cardCh[i]` (descartando para a esquerda) e
  **recebe** de `cardCh[(i+1) mod N]` (do vizinho da direita).

Para sinalizar a batida, usamos um único canal compartilhado
`slapCh chan struct{}`. Quando alguém forma trinca, ele **fecha** o
canal (`close(slapCh)`), o que faz **todos os receivers** do canal
desbloquearem imediatamente — broadcast atômico nativo de Go. Um
`sync.Once` garante que o canal seja fechado uma única vez, mesmo se
dois jogadores formarem trinca simultaneamente.

## Loop do jogador (visão geral)

Todo ponto de bloqueio é protegido por um `select` contra `slapCh`:

```go
select {
case card := <-p.InCh:    // recebeu carta da direita
case <-slapCh:            // alguém bateu — sai do loop
}

// ... atualiza mão, checa trinca, escolhe descarte ...

select {
case p.OutCh <- discard:  // passou a carta para a esquerda
case <-slapCh:            // alguém bateu antes — sai
}
```

Essa estrutura é o que **previne deadlock** (ver análise abaixo).

## Análise de deadlock

O Dorminhoco tem três pontos onde deadlock poderia ocorrer:

### 1. Troca síncrona simétrica no anel

Se todos os jogadores tentassem `send` em canal sem buffer e só depois
`receive`, todos travariam simultaneamente em `send` esperando que
alguém recebesse — clássico deadlock de produtor-consumidor circular.

**Prevenção**: canais de carta com **buffer 1**. Há sempre espaço
para que uma carta seja despachada sem que o vizinho da esquerda
tenha que estar pronto. O `Deal` semeia cada `cardCh` com uma carta
inicial, eliminando também o deadlock de "ninguém tem nada para mandar
primeiro".

### 2. Travamento após batida

Quando um jogador forma trinca, ele para de circular cartas. Os demais
ficariam bloqueados em `<-InCh` esperando uma carta que nunca chegará.

**Prevenção**: todas as operações de canal estão dentro de um `select`
que também escuta `slapCh`. Como o sinal de batida é dado via
`close(slapCh)`, **todos os receivers desbloqueiam atomicamente** —
nenhum jogador pode ficar preso esperando carta após a batida.

### 3. Múltiplas trincas simultâneas

Mais de um jogador pode formar trinca no mesmo turno (cartas
diferentes chegando por canais bufferizados). Se cada um tentasse
fechar `slapCh` independentemente, teríamos `panic: close of closed
channel`.

**Prevenção**: o fechamento de `slapCh` é envolto por `sync.Once.Do`.
Apenas o primeiro a chamar `Do` efetivamente fecha o canal e registra
seu ID em uma variável `atomic.Int32` (`slapper`). Os demais
permanecem como "reactors" comuns.

### 4. Rodada que nunca termina (nenhuma trinca formada)

Em uma distribuição patológica de cartas, é teoricamente possível que
nenhuma trinca seja formada por muitos turnos.

**Prevenção**: cada jogador executa no máximo `MaxTurns` iterações.
Esgotado o limite, ele dispara `knock` por conta própria (que aciona
o mesmo `sync.Once`) e a rodada é encerrada com diagnóstico no log.

## Coleta de métricas

A cada rodada, o sistema:

- Identifica quem **bateu** primeiro (o que formou a trinca).
- Mede o tempo de **reação** de cada jogador, relativo ao momento da
  batida (instante em que `slapCh` foi fechado).
- Ordena os jogadores do mais rápido ao mais lento.
- Anuncia o **perdedor** da rodada: o último a reagir.

## Execução

```bash
go run -race ./cmd/game
```

Saída típica de uma rodada:

```
=== ROUND REPORT ===
Knocked first: Player 1

Reaction order (fastest to slowest):
  #1  Player 0  +0s
  #2  Player 3  +9.3ms
  #3  Player 2  +14.2ms
  #4  Player 4  +26.7ms
  #5  Player 1  +64.2ms  (KNOCKED)

Round loser: Player 1 (last to react, +64.2ms)
```

Observe que o jogador que bateu pode até ser o último a reagir (ele
já está dentro de `Play` quando fecha o canal, e a reação dos demais
acontece em paralelo): isso é coerente com o jogo real, onde quem bate
geralmente reage por último por estar focado em compor a trinca.
