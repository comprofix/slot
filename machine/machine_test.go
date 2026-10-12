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
		{[Reels]Symbol{c, p, c, c, c}, 0, 0},
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
	if math.Abs(got-0.72) > 0.005 {
		t.Errorf("RTP() = %v, want about 0.72", got)
	}
}

func TestWindow(t *testing.T) {
	for reel, strip := range Strips {
		n := len(strip)
		for stop := range strip {
			w := Window(reel, stop)
			want := [Rows]Symbol{strip[(stop+1)%n], strip[stop], strip[(stop+n-1)%n]}
			if w != want {
				t.Fatalf("Window(%d, %d) = %v, want %v", reel, stop, w, want)
			}
		}
	}
}

func TestPaylines(t *testing.T) {
	if len(Paylines) != 25 {
		t.Fatalf("%d paylines, want 25", len(Paylines))
	}
	seen := map[[Reels]int]int{}
	for i, line := range Paylines {
		for _, row := range line {
			if row < Top || row > Bottom {
				t.Errorf("payline %d has row %d", i+1, row)
			}
		}
		if j, ok := seen[line]; ok {
			t.Errorf("paylines %d and %d are the same", j+1, i+1)
		}
		seen[line] = i
	}
	if got := LineChoices[len(LineChoices)-1]; got != len(Paylines) {
		t.Errorf("largest line choice is %d, want %d", got, len(Paylines))
	}
}

// Every spin's result should agree with the grid it reports, line by line.
func TestSpinAccounting(t *testing.T) {
	m := New(1, 2)
	for n := range 20000 {
		m.Credits = StartCredits
		m.Lines = LineChoices[n%len(LineChoices)]
		m.BetPerLine = BetChoices[n/len(LineChoices)%len(BetChoices)]
		before := m.Credits

		r, ok := m.Spin()
		if !ok {
			t.Fatalf("Spin refused with %d credits and bet %d", before, m.TotalBet())
		}
		for i, stop := range r.Stops {
			if r.Grid[i] != Window(i, stop) {
				t.Fatalf("reel %d grid %v does not match stop %d", i, r.Grid[i], stop)
			}
		}

		var wins []LineWin
		total := 0
		for l := range m.Lines {
			var line [Reels]Symbol
			for i, row := range Paylines[l] {
				line[i] = r.Grid[i][row]
			}
			if pay, matched := Evaluate(line); pay > 0 {
				wins = append(wins, LineWin{l, line[0], matched, pay * m.BetPerLine})
				total += pay * m.BetPerLine
			}
		}
		if len(wins) != len(r.Wins) {
			t.Fatalf("%d winning lines, want %d", len(r.Wins), len(wins))
		}
		for i := range wins {
			if r.Wins[i] != wins[i] {
				t.Fatalf("win %d = %+v, want %+v", i, r.Wins[i], wins[i])
			}
		}
		if r.Win != total || m.Credits != before-m.TotalBet()+total {
			t.Fatalf("win %d and credits %d, want %d and %d", r.Win, m.Credits, total, before-m.TotalBet()+total)
		}
	}
}

// A long run of real spins on all lines should pay back close to the
// calculated RTP.
func TestSpinReturn(t *testing.T) {
	const spins = 1_000_000
	m := New(3, 4)
	m.Lines = len(Paylines)
	bet, won := 0, 0
	for range spins {
		m.Credits = StartCredits
		r, _ := m.Spin()
		bet += m.TotalBet()
		won += r.Win
	}
	got := float64(won) / float64(bet)
	t.Logf("paid back %.2f%% over %d spins of %d lines", got*100, spins, m.Lines)
	if math.Abs(got-RTP()) > 0.02 {
		t.Errorf("paid back %v, want about %v", got, RTP())
	}
}

func TestSpinNeedsCredits(t *testing.T) {
	m := New(1, 2)
	m.Lines, m.BetPerLine = 25, 20
	m.Credits = m.TotalBet() - 1
	if _, ok := m.Spin(); ok {
		t.Error("Spin succeeded with too few credits")
	}
	if m.Credits != m.TotalBet()-1 {
		t.Errorf("credits = %d after refused spin, want %d", m.Credits, m.TotalBet()-1)
	}
}
