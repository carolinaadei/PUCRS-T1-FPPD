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
	Iterations      = 20
)

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("=== DINING PHILOSOPHERS — WAITER STRATEGY ===")

	// Create waiter semaphore
	waiter := make(
		chan struct{},
		NumPhilosophers-1,
	)

	// Create forks
	forks := make([]*models.Fork, NumPhilosophers)

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

	// Start dining
	for _, philosopher := range philosophers {

		wg.Add(1)

		go philosopher.DineWithWaiter(
			&wg,
			waiter,
			Iterations,
		)
	}

	wg.Wait()

	fmt.Println("\nDinner finished successfully")

	metrics.PrintFairnessReport(
		philosophers,
	)
}