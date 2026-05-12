package main

import (
	"dorminhoco/internal/logger"
	"dorminhoco/internal/metrics"
	"dorminhoco/internal/models"
	"fmt"
)

const (
	NumPlayers = 5
	NumRounds  = 3
	MaxTurns   = 200 // safety bound per round
)

func main() {

	fmt.Println("=== DORMINHOCO ===")
	fmt.Printf("Players: %d | Rounds: %d\n",
		NumPlayers, NumRounds)

	for round := 1; round <= NumRounds; round++ {

		metrics.PrintGameHeader(round, NumPlayers)

		game := models.NewGame(NumPlayers)
		game.Deal()

		logger.Log("Round starting")

		results := game.PlayRound(MaxTurns)

		metrics.PrintRoundReport(results)
	}

	fmt.Println("\n=== Game session finished ===")
}
