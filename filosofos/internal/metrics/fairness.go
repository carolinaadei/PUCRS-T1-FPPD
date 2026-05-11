package metrics

import (
	"filosofos/internal/models"
	"fmt"
)

func PrintFairnessReport(
	philosophers []*models.Philosopher,
) {
	fmt.Println("\n=== FAIRNESS REPORT ===")

	minMeals := philosophers[0].Stats.GetMeals()
	maxMeals := philosophers[0].Stats.GetMeals()

	for _, philosopher := range philosophers {

		meals := philosopher.Stats.GetMeals()

		fmt.Printf(
			"Philosopher %d ate %d times\n",
			philosopher.ID,
			meals,
		)

		if meals < minMeals {
			minMeals = meals
		}

		if meals > maxMeals {
			maxMeals = meals
		}
	}

	fmt.Printf(
		"\nFairness deviation: %d\n",
		maxMeals-minMeals,
	)
}