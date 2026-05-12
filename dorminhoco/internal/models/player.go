package models

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"dorminhoco/internal/logger"
	"dorminhoco/internal/utils"
)

// HandSize is fixed at 3: the goal is to assemble a
// three-of-a-kind ("trinca"), so each player keeps
// exactly 3 cards in hand at all times.
const HandSize = 3

// Player models one participant of the game as a goroutine.
//
// Channel topology (ring):
//
//	    ... -> Player(i-1) -> Player(i) -> Player(i+1) -> ...
//
//	where the arrow means "discards to". Each player therefore
//	holds:
//	  - InCh:  inbound channel from the RIGHT neighbour
//	  - OutCh: outbound channel to the LEFT neighbour
//
// Both card channels are buffered (capacity 1): at any given
// moment a single "in-flight" card may sit between two
// neighbours without forcing a synchronous rendezvous.
//
// The slap channel is shared by all players in the round and
// is signalled via channel-close (broadcast): the player who
// forms a trinca closes it; every other player observes the
// closure through their select statement and reacts.
type Player struct {
	ID    int
	Hand  []Card
	InCh  chan Card
	OutCh chan Card
}

func NewPlayer(id int, inCh, outCh chan Card) *Player {
	return &Player{
		ID:    id,
		Hand:  make([]Card, 0, HandSize),
		InCh:  inCh,
		OutCh: outCh,
	}
}

// AddCard appends a card to the hand.
func (p *Player) AddCard(c Card) {
	p.Hand = append(p.Hand, c)
}

// HasTrinca returns true if the player's hand contains
// three cards of the same rank.
func (p *Player) HasTrinca() bool {
	if len(p.Hand) < HandSize {
		return false
	}

	counts := make(map[string]int)

	for _, c := range p.Hand {
		counts[c.Rank]++

		if counts[c.Rank] >= HandSize {
			return true
		}
	}

	return false
}

// HandString returns a deterministic textual representation
// of the hand, useful for logging.
func (p *Player) HandString() string {
	ranks := make([]string, len(p.Hand))

	for i, c := range p.Hand {
		ranks[i] = c.Rank
	}

	sort.Strings(ranks)

	out := ""

	for _, r := range ranks {
		out += "[" + r + "]"
	}

	return out
}

// chooseDiscard picks the card to pass to the left neighbour.
// Strategy: discard the card whose rank is least represented
// in the hand. This is a reasonable heuristic and keeps the
// game progressing toward trincas without making one player
// dominant.
func (p *Player) chooseDiscard() (Card, int) {
	counts := make(map[string]int)

	for _, c := range p.Hand {
		counts[c.Rank]++
	}

	worstIdx := 0
	worstCount := counts[p.Hand[0].Rank]

	for i := 1; i < len(p.Hand); i++ {
		c := counts[p.Hand[i].Rank]

		if c < worstCount {
			worstCount = c
			worstIdx = i
		}
	}

	return p.Hand[worstIdx], worstIdx
}

// removeAt removes the card at index idx from the hand.
func (p *Player) removeAt(idx int) {
	p.Hand = append(p.Hand[:idx], p.Hand[idx+1:]...)
}

// RoundResult is the outcome of a single round for one player.
type RoundResult struct {
	PlayerID     int
	Knocked      bool // this player formed the trinca and slapped
	ReactionTime time.Duration
	ReactedAt    int64 // absolute timestamp (UnixNano) at react time
}

// Play runs the player goroutine for a single round.
//
// The goroutine alternates between three responsibilities,
// all multiplexed through select statements so that no path
// can deadlock:
//
//  1. Receive a card from the RIGHT neighbour (InCh) OR
//     observe a slap (close of slapCh).
//  2. Discard a card to the LEFT neighbour (OutCh) OR
//     observe a slap (close of slapCh).
//  3. If the hand forms a trinca, close slapCh to signal
//     every other player; if a slap is observed, react and
//     record reaction time.
//
// slapCh is closed exactly once per round (protected by
// the supplied sync.Once). The reactionCounter is an atomic
// counter incremented in the order players observe the slap;
// the last one to react loses the round.
//
// Deadlock analysis: every blocking channel operation is
// guarded by select against the slap close. Once a slap
// happens, every blocked operation immediately becomes
// ready (the receive on the closed slapCh returns), so no
// goroutine can ever wait forever.
func (p *Player) Play(
	wg *sync.WaitGroup,
	slapCh chan struct{},
	slapOnce *sync.Once,
	slapper *atomic.Int32,
	reactionCounter *atomic.Int32,
	resultsCh chan<- RoundResult,
	maxTurns int,
) {
	defer wg.Done()

	// Sanity check: a player must hold HandSize cards.
	if len(p.Hand) != HandSize {

		logger.Log(
			fmt.Sprintf(
				"Player %d started with %d cards (expected %d)",
				p.ID,
				len(p.Hand),
				HandSize,
			),
		)
	}

	// Check immediately whether the initial deal already
	// contains a trinca.
	if p.HasTrinca() {
		p.knock(slapCh, slapOnce, slapper)
	}

	for turn := 0; turn < maxTurns; turn++ {

		// ---- 1. Receive a card from the right neighbour ----

		var incoming Card
		gotCard := false

		select {
		case incoming = <-p.InCh:
			gotCard = true

		case <-slapCh:
			// Someone (possibly this very player) slapped:
			// stop circulating cards and react.
			p.reactToSlap(
				slapper,
				reactionCounter,
				resultsCh,
			)
			return
		}

		if gotCard {

			p.AddCard(incoming)

			logger.Log(
				fmt.Sprintf(
					"Player %d received %s, hand=%s",
					p.ID,
					incoming,
					p.HandString(),
				),
			)
		}

		// ---- 2. Check trinca after receiving ----

		if p.HasTrinca() {

			logger.Log(
				fmt.Sprintf(
					"Player %d formed a TRINCA with %s — KNOCK!",
					p.ID,
					p.HandString(),
				),
			)

			p.knock(slapCh, slapOnce, slapper)
			p.reactToSlap(slapper, reactionCounter, resultsCh)
			return
		}

		// ---- 3. Discard a card to the left neighbour ----

		// Small think delay to make the game observable
		// and to mimic per-player decision latency.
		utils.RandomDelay(80)

		discard, idx := p.chooseDiscard()
		p.removeAt(idx)

		select {
		case p.OutCh <- discard:

			logger.Log(
				fmt.Sprintf(
					"Player %d discarded %s to left, hand=%s",
					p.ID,
					discard,
					p.HandString(),
				),
			)

		case <-slapCh:
			// A slap was signalled before we managed to
			// pass the card on. Put the card back in hand
			// and react.
			p.AddCard(discard)

			p.reactToSlap(
				slapper,
				reactionCounter,
				resultsCh,
			)
			return
		}
	}

	// Safety net: if no trinca occurred after maxTurns,
	// the round is forced to terminate. Trigger a slap so
	// every player exits cleanly.
	logger.Log(
		fmt.Sprintf(
			"Player %d reached max rounds without trinca",
			p.ID,
		),
	)

	p.knock(slapCh, slapOnce, slapper)
	p.reactToSlap(
		slapper,
		reactionCounter,
		resultsCh,
	)
}

// knock signals the slap, atomically recording who got there
// first. Multiple players may form a trinca on the same
// circulation: sync.Once guarantees that only the first
// caller's closure of slapCh wins.
func (p *Player) knock(
	slapCh chan struct{},
	slapOnce *sync.Once,
	slapper *atomic.Int32,
) {
	slapOnce.Do(func() {
		slapper.Store(int32(p.ID))

		logger.Log(
			fmt.Sprintf(
				"*** Player %d KNOCKED first ***",
				p.ID,
			),
		)

		close(slapCh)
	})
}

// reactToSlap records the player's reaction. The first
// player to call reactToSlap after the slap is the fastest,
// the last one to call it loses the round.
//
// The reaction time is measured against the absolute time
// at which this method is called: since every blocked
// operation in Play is unblocked by the slapCh close, the
// difference in "time to react" comes from the random
// delays earlier in each player's loop. This mirrors the
// real-game intuition that players who were busy thinking
// react more slowly.
func (p *Player) reactToSlap(
	slapper *atomic.Int32,
	reactionCounter *atomic.Int32,
	resultsCh chan<- RoundResult,
) {
	knocked := int(slapper.Load()) == p.ID

	// A tiny additional jitter on the reaction. Players
	// further along their think delay react later. The
	// effect is dominated by where each player was inside
	// utils.RandomDelay at the moment of the slap, which
	// is itself non-deterministic — exactly the property
	// the game is meant to model.
	utils.RandomDelay(20)

	reactedAt := time.Now().UnixNano()
	order := reactionCounter.Add(1)

	logger.Log(
		fmt.Sprintf(
			"Player %d reacted (order #%d, knocked=%v)",
			p.ID,
			order,
			knocked,
		),
	)

	resultsCh <- RoundResult{
		PlayerID:  p.ID,
		Knocked:   knocked,
		ReactedAt: reactedAt,
	}
}
