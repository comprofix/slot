package machine

import (
	"encoding/json"
	"testing"
)

// A game saved and restored, by way of JSON as the game does it, should
// carry on exactly where it left off.
func TestSaveRestore(t *testing.T) {
	m := New(1, 2)
	m.Lines, m.BetPerLine = 25, 5
	for range 200 {
		m.Credits = StartCredits
		m.Spin()
		for m.Link != nil {
			m.Respin()
		}
	}
	m.Credits = 4321
	m.FreeSpins = 3
	startLink(m, 7, m.TotalBet())
	m.Link.Chips[0][0] = Chip{Jackpot: Minor}
	m.Link.Respins = 2
	if m.Major() == MajorStart || m.Grand() == GrandStart {
		t.Fatal("the pots did not grow during play")
	}

	data, err := json.Marshal(m.Save())
	if err != nil {
		t.Fatal(err)
	}
	var s Saved
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	fresh := New(3, 4)
	fresh.Restore(s)

	if fresh.Credits != 4321 || fresh.Lines != 25 || fresh.BetPerLine != 5 || fresh.FreeSpins != 3 {
		t.Errorf("restored %d credits, %d lines at %d, %d free spins", fresh.Credits, fresh.Lines, fresh.BetPerLine, fresh.FreeSpins)
	}
	if fresh.majorPot != m.majorPot || fresh.grandPot != m.grandPot {
		t.Errorf("restored pots %d and %d, want %d and %d", fresh.majorPot, fresh.grandPot, m.majorPot, m.grandPot)
	}
	if fresh.Link == nil || *fresh.Link != *m.Link {
		t.Errorf("restored link feature %+v, want %+v", fresh.Link, m.Link)
	}
	if fresh.Link == m.Link {
		t.Error("the restored feature shares its state with the saved game")
	}
}

// A save that play could not have produced should be put right.
func TestRestoreRepairs(t *testing.T) {
	m := New(1, 2)
	bad := &Feature{Bet: 25, Respins: 3} // holds no chips
	m.Restore(Saved{Credits: -5, Lines: 7, BetPerLine: 3, FreeSpins: -1, Link: bad})

	if m.Credits != 0 || m.FreeSpins != 0 {
		t.Errorf("restored %d credits and %d free spins, want none of either", m.Credits, m.FreeSpins)
	}
	if m.Lines != LineChoices[0] || m.BetPerLine != BetChoices[0] {
		t.Errorf("restored %d lines at %d, want the smallest bet", m.Lines, m.BetPerLine)
	}
	if m.Major() != MajorStart || m.Grand() != GrandStart {
		t.Errorf("restored pots of %d and %d, want their starting amounts", m.Major(), m.Grand())
	}
	if m.Link != nil {
		t.Error("restored a link feature with too few chips")
	}
}

// Resetting the credits must leave the pots as they are.
func TestResetCredits(t *testing.T) {
	m := New(1, 2)
	m.Lines = 25
	for range 50 {
		m.Spin()
		for m.Link != nil {
			m.Respin()
		}
	}
	major, grand := m.majorPot, m.grandPot
	m.Credits = 3
	m.ResetCredits()
	if m.Credits != StartCredits {
		t.Errorf("credits = %d after a reset, want %d", m.Credits, StartCredits)
	}
	if m.majorPot != major || m.grandPot != grand || major == MajorStart*potUnit {
		t.Errorf("pots went from %d and %d to %d and %d on a reset", major, grand, m.majorPot, m.grandPot)
	}
}
