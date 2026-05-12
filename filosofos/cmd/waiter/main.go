package main

import (
	"filosofos/internal/metrics"
	"filosofos/internal/models"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	NumPhilosophers = 5
	Iterations      = 1000
	Experiments     = 5
)

func main() {
	rand.Seed(time.Now().UnixNano())

	for experiment := 1; experiment <= Experiments; experiment++ {

		metrics.PrintExperimentHeader(
			"WAITER",
			experiment,
		)

		waiter := make(
			chan struct{},
			NumPhilosophers-1,
		)

		forks := make(
			[]*models.Fork,
			NumPhilosophers,
		)

		for i := 0; i < NumPhilosophers; i++ {
			forks[i] = models.NewFork(i)
		}

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

		for _, philosopher := range philosophers {

			wg.Add(1)

			go philosopher.DineWithWaiter(
				&wg,
				waiter,
				Iterations,
			)
		}

		wg.Wait()

		metrics.PrintDetailedReport(
			philosophers,
		)
	}

	fmt.Println("\nAll experiments finished")
}