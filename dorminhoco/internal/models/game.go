package models

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"dorminhoco/internal/logger"
)

// Game models a single round of Dorminhoco with N players
// arranged in a ring.
type Game struct {
	NumPlayers int
	Players    []*Player

	// One outbound channel per player. cardCh[i] is the
	// channel that player i SENDS to (i.e. it is the
	// inbound channel of player (i-1) mod N).
	cardCh []chan Card
}

// NewGame builds a ring of NumPlayers, allocates the inter-
// player card channels (buffered, capacity 1) and deals
// HandSize cards to each player.
//
// Buffered-1 card channels prevent the classic ring deadlock
// where every player blocks on send waiting for its left
// neighbour to receive. With one slot of buffer per edge,
// a player can always pass its discard along without first
// receiving from the right.
func NewGame(numPlayers int) *Game {

	if numPlayers < 2 {
		numPlayers = 2
	}

	g := &Game{
		NumPlayers: numPlayers,
		Players:    make([]*Player, numPlayers),
		cardCh:     make([]chan Card, numPlayers),
	}

	// Allocate channels: cardCh[i] is what player i sends
	// to (and what player (i-1) mod N receives from).
	for i := 0; i < numPlayers; i++ {
		g.cardCh[i] = make(chan Card, 1)
	}

	// Wire the players in a ring.
	//   - OutCh of player i = cardCh[i]   (player i sends to its left)
	//   - InCh  of player i = cardCh[(i+1) % N]
	//                         (player i receives from player (i+1),
	//                          its right neighbour)
	for i := 0; i < numPlayers; i++ {
		out := g.cardCh[i]
		in := g.cardCh[(i+1)%numPlayers]

		g.Players[i] = NewPlayer(i, in, out)
	}

	return g
}

// Deal shuffles the deck and gives HandSize cards to each
// player. Any remaining cards are seeded into the inbound
// channels so that play can start immediately (every player
// will have a card waiting on InCh).
//
// Accounting for numPlayers = N:
//   - Deck size      = 4 * N
//   - Dealt to hands = 3 * N  (HandSize per player)
//   - Remaining      = N      (1 per inbound channel slot)
//
// The remaining cards are placed in each cardCh so that
// every player's very first <-p.InCh succeeds without any
// other player having to send first. This eliminates a
// "cold-start" deadlock at the beginning of the round.
func (g *Game) Deal() {
	deck := BuildDeck(g.NumPlayers)
	ShuffleDeck(deck)

	cursor := 0

	// Hand out HandSize cards per player.
	for c := 0; c < HandSize; c++ {

		for i := 0; i < g.NumPlayers; i++ {
			g.Players[i].AddCard(deck[cursor])
			cursor++
		}
	}

	// Seed every inbound channel with one card.
	for i := 0; i < g.NumPlayers; i++ {

		if cursor >= len(deck) {
			break
		}

		// cardCh[i] is the channel that player (i-1) RECEIVES
		// from — but for the cold-start seed we want each
		// player i to have one card waiting on their InCh.
		// Player i's InCh is cardCh[(i+1) % N], so we send
		// to that channel.
		seedTarget := g.cardCh[(i+1)%g.NumPlayers]
		seedTarget <- deck[cursor]
		cursor++
	}

	for i := 0; i < g.NumPlayers; i++ {
		logger.Log(
			fmt.Sprintf(
				"Player %d initial hand: %s",
				i,
				g.Players[i].HandString(),
			),
		)
	}
}

// PlayRound starts all player goroutines and waits for the
// round to terminate (either by a trinca or by the maxTurns
// safety bound). It returns the ordered results of the round.
func (g *Game) PlayRound(maxTurns int) []RoundResult {

	slapCh := make(chan struct{})
	var slapOnce sync.Once
	var slapper atomic.Int32
	slapper.Store(-1)

	var reactionCounter atomic.Int32

	resultsCh := make(chan RoundResult, g.NumPlayers)

	var wg sync.WaitGroup

	roundStart := time.Now()

	for _, p := range g.Players {
		wg.Add(1)
		go p.Play(
			&wg,
			slapCh,
			&slapOnce,
			&slapper,
			&reactionCounter,
			resultsCh,
			maxTurns,
		)
	}

	wg.Wait()
	close(resultsCh)

	// Collect raw results from the goroutines.
	results := make([]RoundResult, 0, g.NumPlayers)

	for r := range resultsCh {
		results = append(results, r)
	}

	// Normalise reaction times against the earliest reaction
	// so the report is human-readable.
	if len(results) > 0 {

		earliest := results[0].ReactedAt
		for _, r := range results {
			if r.ReactedAt < earliest {
				earliest = r.ReactedAt
			}
		}

		for i := range results {
			results[i].ReactionTime = time.Duration(
				results[i].ReactedAt - earliest,
			)
		}
	}

	logger.Log(
		fmt.Sprintf(
			"Round finished in %v",
			time.Since(roundStart),
		),
	)

	return results
}
