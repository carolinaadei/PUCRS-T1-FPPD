package metrics

import (
	"filosofos/internal/models"
	"fmt"
	"time"
)

func PrintDetailedReport(
	philosophers []*models.Philosopher,
) {
	fmt.Println("\n=== METRICS REPORT ===")

	totalMeals := 0

	var totalWait time.Duration

	minMeals := philosophers[0].
		Stats.
		GetMeals()

	maxMeals := philosophers[0].
		Stats.
		GetMeals()

	for _, philosopher := range philosophers {

		meals := philosopher.
			Stats.
			GetMeals()

		totalBlockedTime := philosopher.
			Stats.
			GetTotalWaitTime()

		averageWait := philosopher.
			Stats.
			AverageWaitTime()

		totalMeals += meals

		totalWait += totalBlockedTime

		if meals < minMeals {
			minMeals = meals
		}

		if meals > maxMeals {
			maxMeals = meals
		}

		fmt.Printf(
			"\nPhilosopher %d\n",
			philosopher.ID,
		)

		fmt.Printf(
			"Meals eaten: %d\n",
			meals,
		)

		fmt.Printf(
			"Total blocked time: %v\n",
			totalBlockedTime,
		)

		fmt.Printf(
			"Average waiting time: %v\n",
			averageWait,
		)
	}

	fmt.Println("\n=== GLOBAL ANALYSIS ===")

	fmt.Printf(
		"Total meals: %d\n",
		totalMeals,
	)

	fmt.Printf(
		"Fairness deviation: %d\n",
		maxMeals-minMeals,
	)

	fmt.Printf(
		"Total blocked time: %v\n",
		totalWait,
	)
}