package models

import (
	"fmt"
	"sync"
	"time"

	"filosofos/internal/logger"
	"filosofos/internal/utils"
)

type Philosopher struct {
	ID        int
	LeftFork  *Fork
	RightFork *Fork
	Stats     *PhilosopherStats
}

func NewPhilosopher(
	id int,
	left *Fork,
	right *Fork,
) *Philosopher {
	return &Philosopher{
		ID:        id,
		LeftFork:  left,
		RightFork: right,
		Stats:     &PhilosopherStats{},
	}
}

func (p *Philosopher) Think() {
	logger.Log(
		fmt.Sprintf(
			"Philosopher %d is thinking",
			p.ID,
		),
	)

	utils.RandomDelay(500)
}

func (p *Philosopher) Eat() {
	logger.Log(
		fmt.Sprintf(
			"Philosopher %d started eating",
			p.ID,
		),
	)

	utils.RandomDelay(500)

	p.Stats.IncrementMeals()

	logger.Log(
		fmt.Sprintf(
			"Philosopher %d finished eating",
			p.ID,
		),
	)
}

//
// BASE STRATEGY (DEADLOCK PRONE)
//
// All philosophers pick LEFT fork first, then RIGHT fork.
// If every philosopher manages to acquire the LEFT fork
// before any of them tries the RIGHT fork, a circular
// wait is established (Coffman condition 4) and the system
// deadlocks: no philosopher can ever release a fork.
//

func (p *Philosopher) Dine(
	wg *sync.WaitGroup,
	iterations int,
	startBarrier *sync.WaitGroup,
) {
	defer wg.Done()

	for i := 0; i < iterations; i++ {

		p.Think()

		// On the first iteration, synchronise all philosophers
		// right before the LEFT-fork acquisition. This forces
		// the worst case scenario: everyone reaches for LEFT
		// at the same moment, every fork gets grabbed, and the
		// circular wait is established immediately.
		// Without this, random Think delays let some philosophers
		// finish a full meal before others even start, masking
		// the deadlock.
		if i == 0 && startBarrier != nil {
			startBarrier.Done()
			startBarrier.Wait()
		}

		// Start waiting timer
		waitStart := time.Now()

		// Pick LEFT fork
		logger.Log(
			fmt.Sprintf(
				"Philosopher %d trying LEFT fork %d",
				p.ID,
				p.LeftFork.ID,
			),
		)

		<-p.LeftFork.Token

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d picked LEFT fork %d",
				p.ID,
				p.LeftFork.ID,
			),
		)

		// Forced delay to maximize the chance of deadlock:
		// gives every philosopher time to grab its LEFT fork
		// before anyone tries to grab a RIGHT fork.
		time.Sleep(100 * time.Millisecond)

		// Pick RIGHT fork
		logger.Log(
			fmt.Sprintf(
				"Philosopher %d trying RIGHT fork %d",
				p.ID,
				p.RightFork.ID,
			),
		)

		<-p.RightFork.Token

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d picked RIGHT fork %d",
				p.ID,
				p.RightFork.ID,
			),
		)

		// Stop waiting timer
		waitDuration := time.Since(waitStart)

		p.Stats.AddWaitTime(waitDuration)

		p.Eat()

		// Release forks
		p.LeftFork.Token <- struct{}{}
		p.RightFork.Token <- struct{}{}

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d released forks",
				p.ID,
			),
		)
	}
}

//
// WAITER STRATEGY
//
// A waiter (semaphore channel with capacity N-1) limits
// the number of philosophers competing for forks at the
// same time. With at most N-1 contenders, at least one
// philosopher always has access to both forks.
// Breaks Coffman condition 2 (hold and wait at scale).
//

func (p *Philosopher) DineWithWaiter(
	wg *sync.WaitGroup,
	waiter chan struct{},
	iterations int,
) {
	defer wg.Done()

	for i := 0; i < iterations; i++ {

		p.Think()

		// Ask waiter permission
		logger.Log(
			fmt.Sprintf(
				"Philosopher %d asking waiter permission",
				p.ID,
			),
		)

		waiter <- struct{}{}

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d received waiter permission",
				p.ID,
			),
		)

		// Start waiting timer
		waitStart := time.Now()

		// Pick LEFT fork
		logger.Log(
			fmt.Sprintf(
				"Philosopher %d trying LEFT fork",
				p.ID,
			),
		)

		<-p.LeftFork.Token

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d picked LEFT fork",
				p.ID,
			),
		)

		time.Sleep(100 * time.Millisecond)

		// Pick RIGHT fork
		logger.Log(
			fmt.Sprintf(
				"Philosopher %d trying RIGHT fork",
				p.ID,
			),
		)

		<-p.RightFork.Token

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d picked RIGHT fork",
				p.ID,
			),
		)

		// Stop waiting timer
		waitDuration := time.Since(waitStart)

		p.Stats.AddWaitTime(waitDuration)

		p.Eat()

		// Release forks
		p.LeftFork.Token <- struct{}{}
		p.RightFork.Token <- struct{}{}

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d released forks",
				p.ID,
			),
		)

		// Release waiter permission
		<-waiter

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d released waiter permission",
				p.ID,
			),
		)
	}
}

//
// HIERARCHY STRATEGY
//
// Global ordering: every philosopher acquires the fork
// with the lower ID first and the higher ID second.
// This breaks Coffman condition 4 (circular wait):
// no cycle can form because acquisition follows a
// total order on fork IDs.
//

func (p *Philosopher) DineWithHierarchy(
	wg *sync.WaitGroup,
	iterations int,
) {
	defer wg.Done()

	for i := 0; i < iterations; i++ {

		p.Think()

		// Determine fork order
		firstFork := p.LeftFork
		secondFork := p.RightFork

		if p.RightFork.ID < p.LeftFork.ID {
			firstFork = p.RightFork
			secondFork = p.LeftFork
		}

		// Start waiting timer
		waitStart := time.Now()

		// Pick first fork
		logger.Log(
			fmt.Sprintf(
				"Philosopher %d trying fork %d first",
				p.ID,
				firstFork.ID,
			),
		)

		<-firstFork.Token

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d picked fork %d",
				p.ID,
				firstFork.ID,
			),
		)

		time.Sleep(100 * time.Millisecond)

		// Pick second fork
		logger.Log(
			fmt.Sprintf(
				"Philosopher %d trying fork %d second",
				p.ID,
				secondFork.ID,
			),
		)

		<-secondFork.Token

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d picked fork %d",
				p.ID,
				secondFork.ID,
			),
		)

		// Stop waiting timer
		waitDuration := time.Since(waitStart)

		p.Stats.AddWaitTime(waitDuration)

		p.Eat()

		// Release forks
		firstFork.Token <- struct{}{}
		secondFork.Token <- struct{}{}

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d released forks %d and %d",
				p.ID,
				firstFork.ID,
				secondFork.ID,
			),
		)
	}
}
