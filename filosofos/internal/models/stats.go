package models

import (
	"sync"
	"time"
)

type PhilosopherStats struct {
	Meals int

	TotalWaitTime time.Duration

	Mutex sync.Mutex
}

func (s *PhilosopherStats) IncrementMeals() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.Meals++
}

func (s *PhilosopherStats) AddWaitTime(
	duration time.Duration,
) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.TotalWaitTime += duration
}

func (s *PhilosopherStats) AverageWaitTime() time.Duration {

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.Meals == 0 {
		return 0
	}

	return s.TotalWaitTime / time.Duration(s.Meals)
}