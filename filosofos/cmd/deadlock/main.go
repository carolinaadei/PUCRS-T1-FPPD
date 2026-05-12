package main

import (
	"filosofos/internal/models"
	"fmt"
	"sync"
	"time"
)

const (
	NumPhilosophers = 5
	Iterations      = 100
	WatchdogSeconds = 5
)

func main() {
	fmt.Println("=== DINING PHILOSOPHERS — BASE VERSION (DEADLOCK PRONE) ===")
	fmt.Println("All philosophers pick LEFT fork first, then RIGHT fork.")
	fmt.Println("Expect a circular wait (Coffman condition 4).")
	fmt.Printf("Watchdog: %ds without progress => declares deadlock.\n\n",
		WatchdogSeconds)

	// Create forks
	forks := make([]*models.Fork, NumPhilosophers)

	for i := 0; i < NumPhilosophers; i++ {
		forks[i] = models.NewFork(i)
	}

	// Create philosophers
	philosophers := make([]*models.Philosopher, NumPhilosophers)

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

	// Start barrier: forces all philosophers to leave the
	// initial thinking phase together, guaranteeing that
	// the circular wait pattern is triggered on the very
	// first round (instead of relying on random timing).
	var startBarrier sync.WaitGroup
	startBarrier.Add(NumPhilosophers)

	// Start philosophers
	for _, philosopher := range philosophers {
		wg.Add(1)

		go philosopher.Dine(
			&wg,
			Iterations,
			&startBarrier,
		)
	}

	// Watchdog goroutine: completion vs timeout.
	// The Go runtime itself crashes with
	// "fatal error: all goroutines are asleep - deadlock!"
	// when *every* goroutine is blocked, but the WaitGroup
	// waiter in the main goroutine would prevent the runtime
	// from detecting it cleanly. We use a watchdog so the
	// demo always terminates with a readable diagnosis.
	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("\nDinner finished (no deadlock observed this run)")
	case <-time.After(WatchdogSeconds * time.Second):
		fmt.Printf("\n*** WATCHDOG TRIGGERED: no philosopher finished in %ds ***\n",
			WatchdogSeconds)
		fmt.Println("*** Probable DEADLOCK — circular wait reached ***")

		// Print partial progress
		fmt.Println("\nPartial meal counts at the moment of detection:")

		for _, p := range philosophers {
			fmt.Printf("  Philosopher %d: %d meals\n",
				p.ID,
				p.Stats.GetMeals(),
			)
		}
	}
}
