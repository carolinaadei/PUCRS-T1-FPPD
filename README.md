# PUCRS — T1 FPPD

Trabalho 1 da disciplina **FPPD — Fundamentos de Processamento Paralelo e
Distribuído** (98713-04), Opção A: Problemas Clássicos de Concorrência.

O repositório contém os dois problemas exigidos, cada um em sua própria
pasta com seu próprio `go.mod`:

```
PUCRS-T1-FPPD/
├── filosofos/    Problema 1 — Jantar dos Filósofos
└── dorminhoco/   Problema 2 — Jogo do Dorminhoco
```

Cada subprojeto possui seu próprio `README.md` com instruções de execução
e descrição da arquitetura.

## Requisitos

- Go 1.22 ou superior
- Plataforma: qualquer (Linux, macOS, Windows com WSL ou nativo)

## Execução rápida

```bash
# Problema 1 — Jantar dos Filósofos
cd filosofos
go run -race ./cmd/deadlock    # versão base (demonstra deadlock)
go run -race ./cmd/hierarchy   # estratégia 1: hierarquia de garfos
go run -race ./cmd/waiter      # estratégia 2: garçom (limita participantes)

# Problema 2 — Jogo do Dorminhoco
cd ../dorminhoco
go run -race ./cmd/game
```

Todos os binários executam sem data race com `go run -race`.