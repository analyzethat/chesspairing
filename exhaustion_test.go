// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0
//
// Modified in the analyzethat fork: file added.

package chesspairing_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/analyzethat/chesspairing"
	"github.com/analyzethat/chesspairing/pairing/burstein"
	"github.com/analyzethat/chesspairing/pairing/doubleswiss"
	"github.com/analyzethat/chesspairing/pairing/dubov"
	"github.com/analyzethat/chesspairing/pairing/dutch"
	"github.com/analyzethat/chesspairing/pairing/lim"
	"github.com/analyzethat/chesspairing/pairing/roundrobin"
)

// buildExhaustedState returns a state for a field of n players in which every
// player has already met every other player, so that no legal pairing for the
// upcoming round exists. The history is a complete round robin produced by the
// round-robin pairer; every game is drawn so all players stay on equal score.
func buildExhaustedState(t *testing.T, n int) *chesspairing.TournamentState {
	t.Helper()

	players := make([]chesspairing.PlayerEntry, n)
	for i := range n {
		id := fmt.Sprintf("p%d", i+1)
		players[i] = chesspairing.PlayerEntry{ID: id, DisplayName: id, Rating: 2000 - i*10}
	}

	state := &chesspairing.TournamentState{
		Players:       players,
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingRoundRobin, Options: map[string]any{}},
		ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringStandard, Options: map[string]any{}},
	}

	rr := roundrobin.New(roundrobin.Options{})
	for round := 1; round <= n-1; round++ {
		state.CurrentRound = round
		res, err := rr.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("building history: round %d: %v", round, err)
		}
		rd := chesspairing.RoundData{Number: round}
		for _, p := range res.Pairings {
			rd.Games = append(rd.Games, chesspairing.GameData{
				WhiteID: p.WhiteID, BlackID: p.BlackID, Result: chesspairing.ResultDraw,
			})
		}
		rd.Byes = append(rd.Byes, res.Byes...)
		state.Rounds = append(state.Rounds, rd)
	}

	state.CurrentRound = n
	return state
}

// assertCompleteOrError fails unless the pairer either reported an error or
// returned a result in which every active player is accounted for exactly
// once. Returning a partial result with a nil error is the failure this test
// exists to catch: the caller has no way to notice that players were dropped.
func assertCompleteOrError(t *testing.T, name string, state *chesspairing.TournamentState, res *chesspairing.PairingResult, err error) {
	t.Helper()

	if err != nil {
		return // Reporting the impossibility is the correct behaviour.
	}
	if res == nil {
		t.Errorf("%s: nil result and nil error", name)
		return
	}

	seen := make(map[string]int, len(state.Players))
	for _, p := range res.Pairings {
		seen[p.WhiteID]++
		seen[p.BlackID]++
	}
	byes := 0
	for _, b := range res.Byes {
		seen[b.PlayerID]++
		if b.Type == chesspairing.ByePAB {
			byes++
		}
	}

	for _, p := range state.Players {
		switch seen[p.ID] {
		case 1:
			// Accounted for.
		case 0:
			t.Errorf("%s: no error, but player %s is neither paired nor given a bye "+
				"(%d pairings, %d byes for %d players)",
				name, p.ID, len(res.Pairings), len(res.Byes), len(state.Players))
		default:
			t.Errorf("%s: no error, but player %s appears %d times", name, p.ID, seen[p.ID])
		}
	}
	if byes > 1 {
		t.Errorf("%s: no error, but %d players received a pairing-allocated bye", name, byes)
	}
}

// TestSwissPairers_ExhaustedFieldIsReported checks that a Swiss pairer facing a
// round it cannot legally pair says so, instead of silently returning a result
// that leaves active players out of the round.
//
// FIDE C.04.3 article 1.9.3 leaves an impossible round-pairing to the Chief
// Arbiter. A library cannot make that call, but it must surface the situation:
// a caller that trusts a nil error will record a round in which players
// vanished.
func TestSwissPairers_ExhaustedFieldIsReported(t *testing.T) {
	pairers := []struct {
		name   string
		pairer chesspairing.Pairer
	}{
		{"dutch", dutch.New(dutch.Options{})},
		{"burstein", burstein.New(burstein.Options{})},
		{"dubov", dubov.New(dubov.Options{})},
		{"lim", lim.New(lim.Options{})},
		{"doubleswiss", doubleswiss.New(doubleswiss.Options{})},
	}

	for _, n := range []int{4, 6, 8} {
		for _, tc := range pairers {
			t.Run(fmt.Sprintf("%s/%dplayers", tc.name, n), func(t *testing.T) {
				state := buildExhaustedState(t, n)
				res, err := tc.pairer.Pair(context.Background(), state)
				assertCompleteOrError(t, tc.name, state, res, err)
			})
		}
	}
}
