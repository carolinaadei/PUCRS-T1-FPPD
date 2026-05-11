package models

type Fork struct {
	ID    int
	Token chan struct{}
}

func NewFork(id int) *Fork {
	fork := &Fork{
		ID:    id,
		Token: make(chan struct{}, 1),
	}

	// Fork starts available
	fork.Token <- struct{}{}

	return fork
}