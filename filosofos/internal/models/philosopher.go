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
	Stats     *Stats
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
		Stats:     &Stats{},
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

func (p *Philosopher) DineWithWaiter(
	wg *sync.WaitGroup,
	waiter chan struct{},
	iterations int,
) {
	defer wg.Done()

	for i := 0; i < iterations; i++ {

		p.Think()

		// Request permission from waiter
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

		// Pick left fork
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

		// Pick right fork
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