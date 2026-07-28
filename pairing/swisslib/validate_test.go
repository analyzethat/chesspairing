// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"testing"

	"github.com/analyzethat/chesspairing"
)

func TestValidatePairing_Valid(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
		{ID: "p3", Active: true},
		{ID: "p4", Active: true},
	}
	result := &chesspairing.PairingResult{
		Pairings: []chesspairing.GamePairing{
			{Board: 1, WhiteID: "p1", BlackID: "p2"},
			{Board: 2, WhiteID: "p3", BlackID: "p4"},
		},
	}
	if err := ValidatePairing(players, result); err != nil {
		t.Errorf("valid pairing should not error: %v", err)
	}
}

func TestValidatePairing_DuplicatePlayer(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
		{ID: "p3", Active: true},
		{ID: "p4", Active: true},
	}
	result := &chesspairing.PairingResult{
		Pairings: []chesspairing.GamePairing{
			{Board: 1, WhiteID: "p1", BlackID: "p2"},
			{Board: 2, WhiteID: "p1", BlackID: "p3"}, // p1 paired twice!
		},
	}
	if err := ValidatePairing(players, result); err == nil {
		t.Error("duplicate player should fail validation")
	}
}

func TestValidatePairing_UnknownPlayer(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
	}
	result := &chesspairing.PairingResult{
		Pairings: []chesspairing.GamePairing{
			{Board: 1, WhiteID: "p1", BlackID: "p99"}, // p99 not in player list
		},
	}
	if err := ValidatePairing(players, result); err == nil {
		t.Error("unknown player should fail validation")
	}
}

func TestValidatePairing_OddWithBye(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
		{ID: "p3", Active: true},
	}
	result := &chesspairing.PairingResult{
		Pairings: []chesspairing.GamePairing{
			{Board: 1, WhiteID: "p1", BlackID: "p2"},
		},
		Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
	}
	if err := ValidatePairing(players, result); err != nil {
		t.Errorf("odd players with bye should be valid: %v", err)
	}
}

func TestValidatePairing_MissingPlayer(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
		{ID: "p3", Active: true},
		{ID: "p4", Active: true},
	}
	result := &chesspairing.PairingResult{
		Pairings: []chesspairing.GamePairing{
			{Board: 1, WhiteID: "p1", BlackID: "p2"},
		},
		// p3 and p4 are neither paired nor have bye
	}
	if err := ValidatePairing(players, result); err == nil {
		t.Error("missing players should fail validation")
	}
}

// Added in the analyzethat fork.

func TestValidatePairing_TooManyByes(t *testing.T) {
	// Handing every player a pairing-allocated bye accounts for all of them
	// exactly once, so every other check passes. It still means the pairer
	// paired nobody.
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
		{ID: "p3", Active: true},
		{ID: "p4", Active: true},
	}
	result := &chesspairing.PairingResult{
		Byes: []chesspairing.ByeEntry{
			{PlayerID: "p1", Type: chesspairing.ByePAB},
			{PlayerID: "p2", Type: chesspairing.ByePAB},
			{PlayerID: "p3", Type: chesspairing.ByePAB},
			{PlayerID: "p4", Type: chesspairing.ByePAB},
		},
	}
	if err := ValidatePairing(players, result); err == nil {
		t.Error("four byes for four players should fail validation")
	}
}

func TestValidateResult_IgnoresPreAssignedByes(t *testing.T) {
	// Pre-assigned byes are the caller's input: those players never entered
	// the matching pool, so they must not be held against the pairer.
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
	}
	preAssigned := []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByeHalf}}
	result := &chesspairing.PairingResult{
		Pairings: []chesspairing.GamePairing{{Board: 1, WhiteID: "p1", BlackID: "p2"}},
		Byes:     preAssigned,
	}
	if err := ValidateResult(players, preAssigned, result); err != nil {
		t.Errorf("pre-assigned bye should not fail validation: %v", err)
	}
}

func TestValidateResult_NilResult(t *testing.T) {
	if err := ValidateResult(nil, nil, nil); err == nil {
		t.Error("nil result should fail validation")
	}
}
