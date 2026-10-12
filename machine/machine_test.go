package machine

import (
	"math"
	"testing"
)

func TestEvaluate(t *testing.T) {
	const (
		c = Cherry
		p = Plum
		l = Bell
		b = Bar
		s = Seven
		d = Dollar
	)
	tests := []struct {
		line    [Reels]Symbol
		pay     int
		matched int
	}{
		{[Reels]Symbol{d, d, d, d, d}, 5000, 5},
		{[Reels]Symbol{d, d, d, d, s}, 500, 4},
		{[Reels]Symbol{d, d, d, b, d}, 100, 3},
		{[Reels]Symbol{s, s, s, s, s}, 750, 5},
		{[Reels]Symbol{s, s, s, p, p}, 30, 3},
		{[Reels]Symbol{b, b, b, b, c}, 75, 4},
		{[Reels]Symbol{l, l, l, c, c}, 8, 3},
		{[Reels]Symbol{p, p, p, p, p}, 75, 5},
		{[Reels]Symbol{c, c, c, c, c}, 50, 5},
		{[Reels]Symbol{c, c, c, b, c}, 5, 3},
		{[Reels]Symbol{c, c, b, c, c}, 2, 2},
		{[Reels]Symbol{c, p, c, c, c}, 1, 1},
		{[Reels]Symbol{p, c, c, c, c}, 0, 0},
		{[Reels]Symbol{b, b, s, b, b}, 0, 0},
		{[Reels]Symbol{p, s, s, s, s}, 0, 0},
		{[Reels]Symbol{p, l, d, b, s}, 0, 0},
	}
	for _, tt := range tests {
		pay, matched := Evaluate(tt.line)
		if pay != tt.pay || matched != tt.matched {
			t.Errorf("Evaluate(%v) = %d, %d; want %d, %d", tt.line, pay, matched, tt.pay, tt.matched)
		}
	}
}

func TestRTP(t *testing.T) {
	got := RTP()
	t.Logf("return to player: %.2f%%", got*100)
	if math.Abs(got-0.92) > 0.005 {
		t.Errorf("RTP() = %v, want about 0.92", got)
	}
}

func TestWindow(t *testing.T) {
	for reel, strip := range Strips {
		n := len(strip)
		for stop := range strip {
			w := Window(reel, stop)
			want := [3]Symbol{strip[(stop+1)%n], strip[stop], strip[(stop+n-1)%n]}
			if w != want {
				t.Fatalf("Window(%d, %d) = %v, want %v", reel, stop, w, want)
			}
		}
	}
}

func TestSpinAccounting(t *testing.T) {
	m := New(1, 2)
	m.SetBet(MaxBet)
	for range 10000 {
		if m.Credits == 0 {
			m.Credits = StartCredits
		}
		m.SetBet(MaxBet)
		before, bet := m.Credits, m.Bet
		r, ok := m.Spin()
		if !ok {
			t.Fatalf("Spin refused with %d credits and bet %d", before, bet)
		}
		pay, matched := Evaluate(r.Line)
		if r.Win != pay*bet || r.Matched != matched {
			t.Fatalf("result %+v does not match Evaluate = %d, %d at bet %d", r, pay, matched, bet)
		}
		if m.Credits != before-bet+r.Win {
			t.Fatalf("credits = %d, want %d", m.Credits, before-bet+r.Win)
		}
		for i, stop := range r.Stops {
			if Strips[i][stop] != r.Line[i] {
				t.Fatalf("reel %d stop %d shows %v, result says %v", i, stop, Strips[i][stop], r.Line[i])
			}
		}
	}
}

// A long run of real spins should pay back close to the calculated RTP.
func TestSpinReturn(t *testing.T) {
	const spins = 4_000_000
	m := New(3, 4)
	won := 0
	for range spins {
		m.Credits = StartCredits
		r, _ := m.Spin()
		won += r.Win
	}
	got := float64(won) / spins
	t.Logf("paid back %.2f%% over %d spins", got*100, spins)
	if math.Abs(got-RTP()) > 0.02 {
		t.Errorf("paid back %v, want about %v", got, RTP())
	}
}

func TestSpinNeedsCredits(t *testing.T) {
	m := New(1, 2)
	m.Credits = 0
	if _, ok := m.Spin(); ok {
		t.Error("Spin succeeded with no credits")
	}
	if m.Credits != 0 {
		t.Errorf("credits = %d after refused spin, want 0", m.Credits)
	}
}

func TestSetBet(t *testing.T) {
	m := New(1, 2)
	for _, tt := range []struct{ credits, bet, want int }{
		{100, 0, 1},
		{100, 2, 2},
		{100, 9, MaxBet},
		{2, 3, 2},
		{0, 3, 1},
	} {
		m.Credits = tt.credits
		m.SetBet(tt.bet)
		if m.Bet != tt.want {
			t.Errorf("SetBet(%d) with %d credits = %d, want %d", tt.bet, tt.credits, m.Bet, tt.want)
		}
	}
}
