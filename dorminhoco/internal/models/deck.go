package models

import (
	"math/rand"
)

// BuildDeck creates a deck containing 4 copies of each
// rank, for a number of distinct ranks equal to numPlayers.
// This guarantees that:
//   - every player can be dealt 3 cards (hand size = 3),
//   - one extra "circulating" card is always in flight,
//   - a three-of-a-kind ("trinca") is always reachable.
//
// Total cards = 4 * numPlayers; dealt cards = 3 * numPlayers;
// circulating cards = numPlayers (one per player slot in the
// inbound channel of each player at game start, minus the
// initial hand). See Game.Deal for the exact accounting.
func BuildDeck(numPlayers int) []Card {
	ranks := []string{
		"A", "K", "Q", "J", "10",
		"9", "8", "7", "6", "5",
		"4", "3", "2",
	}

	if numPlayers > len(ranks) {
		numPlayers = len(ranks)
	}

	deck := make([]Card, 0, 4*numPlayers)

	for r := 0; r < numPlayers; r++ {
		for j := 0; j < 4; j++ {
			deck = append(deck, NewCard(ranks[r]))
		}
	}

	return deck
}

// ShuffleDeck performs an in-place Fisher-Yates shuffle.
func ShuffleDeck(deck []Card) {
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
}
