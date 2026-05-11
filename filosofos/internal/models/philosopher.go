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
	}
}

func (p *Philosopher) Think() {
	logger.Log(
		fmt.Sprintf(
			"Philosopher %d is thinking",
			p.ID,
		),
	)

	utils.RandomDelay(1000)
}

func (p *Philosopher) Eat() {
	logger.Log(
		fmt.Sprintf(
			"Philosopher %d started eating",
			p.ID,
		),
	)

	utils.RandomDelay(1000)

	logger.Log(
		fmt.Sprintf(
			"Philosopher %d finished eating",
			p.ID,
		),
	)
}

func (p *Philosopher) Dine(
	wg *sync.WaitGroup,
	iterations int,
) {
	defer wg.Done()

	for i := 0; i < iterations; i++ {

		// Thinking
		p.Think()

		// Pick left fork
		logger.Log(
			fmt.Sprintf(
				"Philosopher %d trying to pick LEFT fork",
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

		// Artificial pause to increase deadlock probability
		time.Sleep(200 * time.Millisecond)

		// Pick right fork
		logger.Log(
			fmt.Sprintf(
				"Philosopher %d waiting for RIGHT fork",
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

		// Eating
		p.Eat()

		// Release forks
		p.LeftFork.Token <- struct{}{}
		p.RightFork.Token <- struct{}{}

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d released both forks",
				p.ID,
			),
		)
	}
}