// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0
//
// Modified in the analyzethat fork: added ValidateResult, and wired it into
// the pairers so a result that drops players is reported instead of returned.

package swisslib

import (
	"fmt"

	"github.com/analyzethat/chesspairing"
)

// ValidatePairing checks that a PairingResult is structurally valid:
// - Every active player is either paired exactly once or has a bye.
// - No player appears in more than one pairing.
// - No unknown player IDs in pairings or byes.
// - Board numbers are sequential starting from 1.
func ValidatePairing(players []PlayerState, result *chesspairing.PairingResult) error {
	activeIDs := make(map[string]bool, len(players))
	for _, p := range players {
		if p.Active {
			activeIDs[p.ID] = true
		}
	}

	seen := make(map[string]bool)

	// Check pairings.
	for i, pair := range result.Pairings {
		if pair.Board != i+1 {
			return fmt.Errorf("board number mismatch: expected %d, got %d", i+1, pair.Board)
		}

		for _, id := range []string{pair.WhiteID, pair.BlackID} {
			if !activeIDs[id] {
				return fmt.Errorf("unknown or inactive player in pairing: %s", id)
			}
			if seen[id] {
				return fmt.Errorf("player %s appears in multiple pairings", id)
			}
			seen[id] = true
		}
	}

	// Check byes.
	for _, bye := range result.Byes {
		if !activeIDs[bye.PlayerID] {
			return fmt.Errorf("unknown or inactive player in bye: %s", bye.PlayerID)
		}
		if seen[bye.PlayerID] {
			return fmt.Errorf("player %s appears in both pairing and bye", bye.PlayerID)
		}
		seen[bye.PlayerID] = true
	}

	// Check all active players accounted for.
	for id := range activeIDs {
		if !seen[id] {
			return fmt.Errorf("active player %s not paired or given bye", id)
		}
	}

	// A round carries exactly one pairing-allocated bye when the field is odd
	// and none when it is even. Handing out more is how a pairer can satisfy
	// every check above while having paired nobody at all.
	pab := 0
	for _, bye := range result.Byes {
		if bye.Type == chesspairing.ByePAB {
			pab++
		}
	}
	if pab != len(activeIDs)%2 {
		return fmt.Errorf("%d pairing-allocated byes for %d active players", pab, len(activeIDs))
	}

	return nil
}

// ValidateResult checks a pairer's own output on the way out, against the
// players that entered its matching pool. Bye entries for players filtered out
// beforehand are the caller's input rather than the pairer's output, so they
// are excluded from the check.
//
// A Swiss system can legitimately run out of legal pairings — with a small
// field and many rounds every opponent eventually gets used up. FIDE C.04.3
// article 1.9.3 leaves that to the Chief Arbiter, which a library cannot
// decide. What it must not do is return a result in which active players
// silently go missing: a caller that trusts a nil error records a round with
// players dropped from it. This turns that into an error the caller can act on.
func ValidateResult(players []PlayerState, preAssigned []chesspairing.ByeEntry, result *chesspairing.PairingResult) error {
	if result == nil {
		return fmt.Errorf("pairer returned no result")
	}
	if len(preAssigned) == 0 {
		return ValidatePairing(players, result)
	}

	skip := make(map[string]bool, len(preAssigned))
	for _, b := range preAssigned {
		skip[b.PlayerID] = true
	}
	trimmed := *result
	trimmed.Byes = make([]chesspairing.ByeEntry, 0, len(result.Byes))
	for _, b := range result.Byes {
		if !skip[b.PlayerID] {
			trimmed.Byes = append(trimmed.Byes, b)
		}
	}
	return ValidatePairing(players, &trimmed)
}
