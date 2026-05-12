package metrics

import (
	"fmt"
	"sort"

	"dorminhoco/internal/models"
)

// PrintGameHeader prints a banner for a game run.
func PrintGameHeader(run int, numPlayers int) {
	fmt.Printf(
		"\n=== DORMINHOCO — Run %d | %d players ===\n",
		run,
		numPlayers,
	)
}

// PrintRoundReport sorts the round results by reaction order
// and prints a comparative table. The slowest reactor is
// announced as the loser of the round.
func PrintRoundReport(results []models.RoundResult) {

	if len(results) == 0 {
		fmt.Println("No results to report")
		return
	}

	// Sort by reaction time ascending.
	sorted := make([]models.RoundResult, len(results))
	copy(sorted, results)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ReactedAt < sorted[j].ReactedAt
	})

	fmt.Println("\n=== ROUND REPORT ===")

	knocker := -1
	for _, r := range sorted {
		if r.Knocked {
			knocker = r.PlayerID
			break
		}
	}

	if knocker >= 0 {
		fmt.Printf(
			"Knocked first: Player %d\n",
			knocker,
		)
	}

	fmt.Println("\nReaction order (fastest to slowest):")

	for pos, r := range sorted {

		marker := ""

		switch {
		case r.Knocked:
			marker = "  (KNOCKED)"
		case pos == len(sorted)-1:
			marker = "  <-- LOST THE ROUND"
		}

		fmt.Printf(
			"  #%d  Player %d  +%v%s\n",
			pos+1,
			r.PlayerID,
			r.ReactionTime,
			marker,
		)
	}

	loser := sorted[len(sorted)-1]

	fmt.Printf(
		"\nRound loser: Player %d (last to react, +%v)\n",
		loser.PlayerID,
		loser.ReactionTime,
	)
}
