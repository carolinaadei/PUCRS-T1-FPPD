# Problema 1 — Jantar dos Filósofos

Implementação do problema clássico do Jantar dos Filósofos em Go,
utilizando **goroutines** e **channels** para representar os garfos.

## Estrutura

```
filosofos/
├── cmd/
│   ├── deadlock/     versão base (demonstra deadlock)
│   ├── hierarchy/    estratégia 1: ordem global de aquisição
│   └── waiter/       estratégia 2: garçom (semáforo contador)
├── internal/
│   ├── models/       Fork, Philosopher, Stats
│   ├── metrics/      relatórios de execução e fairness
│   ├── logger/       logging com timestamp
│   └── utils/        utilitários (delays aleatórios)
├── results/          saídas das execuções (gerar via redirect)
└── go.mod
```

## Modelagem dos garfos

Cada garfo é um `chan struct{}` com buffer 1 contendo inicialmente um
token. Pegar o garfo = receber do canal; devolver o garfo = enviar de
volta. Esse padrão (recomendado pela aula 10) torna o garfo um recurso
de exclusão mútua sem `sync.Mutex`.

## Versões

### `deadlock` (versão base)

Todos os filósofos pegam o garfo da esquerda primeiro, depois o da
direita. Para tornar o deadlock **determinístico** (independente do
escalonador), os filósofos sincronizam em uma `sync.WaitGroup`
imediatamente antes da primeira tentativa de pegar o garfo da esquerda
— garantindo que todos cheguem juntos à fase de aquisição.

Resultado: **deadlock garantido na primeira rodada**. Um watchdog de
5 segundos no `main` detecta a ausência de progresso e imprime
um diagnóstico, evitando que o programa fique travado indefinidamente.

```bash
go run -race ./cmd/deadlock
```

Saída esperada (final):
```
*** WATCHDOG TRIGGERED: no philosopher finished in 5s ***
*** Probable DEADLOCK — circular wait reached ***
Partial meal counts at the moment of detection:
  Philosopher 0: 0 meals
  Philosopher 1: 0 meals
  ...
```

### `hierarchy` (estratégia 1: quebra de simetria por ordem global)

Cada filósofo adquire primeiro o garfo de **menor ID** e depois o de
maior ID, independentemente de qual seja o da esquerda ou direita. Isso
estabelece uma ordem total sobre os recursos e elimina a espera
circular (condição 4 de Coffman).

```bash
go run -race ./cmd/hierarchy
```

### `waiter` (estratégia 2: limitar participantes)

Um garçom modelado como `chan struct{}` com capacidade `N-1` limita
o número de filósofos competindo por garfos ao mesmo tempo. Com no
máximo `N-1` competidores, sempre há um filósofo com acesso aos dois
garfos — elimina a condição 2 de Coffman (hold-and-wait).

```bash
go run -race ./cmd/waiter
```

## Coleta de métricas

Tanto `hierarchy` quanto `waiter` executam `Experiments = 5` rodadas
de `Iterations = 100` refeições por filósofo, imprimindo no final:

- Número de refeições por filósofo
- Tempo total bloqueado (esperando pelos garfos)
- Tempo médio de espera por refeição
- "Fairness deviation": diferença entre o filósofo que mais comeu e o
  que menos comeu — quanto menor, mais justa a estratégia

Para salvar resultados em arquivo:

```bash
go run ./cmd/hierarchy > results/hierarchy_results.txt
go run ./cmd/waiter    > results/waiter_results.txt
```

## Análise resumida

| Versão     | Deadlock?    | Starvation?            | Fairness  |
|------------|--------------|------------------------|-----------|
| `deadlock` | **Sim**      | N/A (ninguém progride) | N/A       |
| `hierarchy`| Não          | Possível em teoria     | Alta      |
| `waiter`   | Não          | Não (com fila do canal)| Alta      |

A discussão detalhada (Coffman, safety/liveness, fairness fraca/forte)
está no relatório PDF.
