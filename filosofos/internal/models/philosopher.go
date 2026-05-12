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
// WAITER STRATEGY
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