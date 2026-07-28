// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0
//
// Modified in the analyzethat fork: file added.

package lexswiss

import (
	"fmt"

	"github.com/analyzethat/chesspairing"
)

// ValidateResult mirrors swisslib.ValidateResult for the lexicographic Swiss
// systems, which carry their own participant state. See the swisslib version
// for the rationale; the two are kept independent on purpose, as with
// FilterPreAssignedByes.
//
// Bye entries for participants filtered out before pairing are the caller's
// input rather than the pairer's output, so they are excluded from the check.
func ValidateResult(participants []ParticipantState, preAssigned []chesspairing.ByeEntry, result *chesspairing.PairingResult) error {
	if result == nil {
		return fmt.Errorf("pairer returned no result")
	}

	skip := make(map[string]bool, len(preAssigned))
	for _, b := range preAssigned {
		skip[b.PlayerID] = true
	}

	activeIDs := make(map[string]bool, len(participants))
	for _, p := range participants {
		if p.Active {
			activeIDs[p.ID] = true
		}
	}

	seen := make(map[string]bool, len(activeIDs))
	for i, pair := range result.Pairings {
		if pair.Board != i+1 {
			return fmt.Errorf("board number mismatch: expected %d, got %d", i+1, pair.Board)
		}
		for _, id := range []string{pair.WhiteID, pair.BlackID} {
			if !activeIDs[id] {
				return fmt.Errorf("unknown or inactive participant in pairing: %s", id)
			}
			if seen[id] {
				return fmt.Errorf("participant %s appears in multiple pairings", id)
			}
			seen[id] = true
		}
	}

	for _, bye := range result.Byes {
		if skip[bye.PlayerID] {
			continue
		}
		if !activeIDs[bye.PlayerID] {
			return fmt.Errorf("unknown or inactive participant in bye: %s", bye.PlayerID)
		}
		if seen[bye.PlayerID] {
			return fmt.Errorf("participant %s appears in both pairing and bye", bye.PlayerID)
		}
		seen[bye.PlayerID] = true
	}

	for id := range activeIDs {
		if !seen[id] {
			return fmt.Errorf("active participant %s not paired or given bye", id)
		}
	}

	// A round carries exactly one pairing-allocated bye when the field is odd
	// and none when it is even. Handing out more is how a pairer can satisfy
	// every check above while having paired nobody at all.
	pab := 0
	for _, bye := range result.Byes {
		if !skip[bye.PlayerID] && bye.Type == chesspairing.ByePAB {
			pab++
		}
	}
	if pab != len(activeIDs)%2 {
		return fmt.Errorf("%d pairing-allocated byes for %d active participants", pab, len(activeIDs))
	}

	return nil
}
