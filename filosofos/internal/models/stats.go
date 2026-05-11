package models

import "sync"

type Stats struct {
	Meals int
	Mutex sync.Mutex
}

func (s *Stats) IncrementMeals() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.Meals++
}

func (s *Stats) GetMeals() int {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	return s.Meals
}