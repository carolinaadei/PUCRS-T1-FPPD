package main

import (
	"filosofos/internal/metrics"
	"filosofos/internal/models"
	"fmt"
	"sync"
)

const (
	NumPhilosophers = 5
	Iterations      = 100
	Experiments     = 5
)

func main() {
	for experiment := 1; experiment <= Experiments; experiment++ {

		metrics.PrintExperimentHeader(
			"HIERARCHY",
			experiment,
		)

		// Create forks
		forks := make(
			[]*models.Fork,
			NumPhilosophers,
		)

		for i := 0; i < NumPhilosophers; i++ {
			forks[i] = models.NewFork(i)
		}

		// Create philosophers
		philosophers := make(
			[]*models.Philosopher,
			NumPhilosophers,
		)

		for i := 0; i < NumPhilosophers; i++ {

			leftFork := forks[i]
			rightFork := forks[(i+1)%NumPhilosophers]

			philosophers[i] = models.NewPhilosopher(
				i,
				leftFork,
				rightFork,
			)
		}

		var wg sync.WaitGroup

		// Start simulation
		for _, philosopher := range philosophers {

			wg.Add(1)

			go philosopher.DineWithHierarchy(
				&wg,
				Iterations,
			)
		}

		wg.Wait()

		fmt.Println(
			"\nDinner finished successfully",
		)

		metrics.PrintDetailedReport(
			philosophers,
		)
	}

	fmt.Println("\nAll experiments finished")
}