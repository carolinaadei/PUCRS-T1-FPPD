package models

import "fmt"

// A Card carries a single Rank value (Ace, King, Queen, ...).
// Suits are not modelled: the game only needs equality on rank
// to detect a three-of-a-kind ("trinca").
type Card struct {
	Rank string
}

func NewCard(rank string) Card {
	return Card{Rank: rank}
}

func (c Card) String() string {
	return fmt.Sprintf("[%s]", c.Rank)
}
