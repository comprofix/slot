package machine

import (
	"math"
	"testing"
)

func TestEvaluate(t *testing.T) {
	const (
		t0 = Ten
		a  = Ace
		c  = Cherry
		p  = Plum
		b  = Bar
		s  = Seven
		d  = Dollar
		w  = Wild
		x  = Scatter
		k  = Link
	)
	tests := []struct {
		line    [Reels]Symbol
		pay     int
		sym     Symbol
		matched int
	}{
		{[Reels]Symbol{d, d, d, d, d}, 5000, d, 5},
		{[Reels]Symbol{d, d, d, d, s}, 500, d, 4},
		{[Reels]Symbol{d, d, d, b, d}, 100, d, 3},
		{[Reels]Symbol{s, s, s, s, s}, 1000, s, 5},
		{[Reels]Symbol{t0, t0, t0, a, a}, 5, t0, 3},
		{[Reels]Symbol{a, a, a, a, c}, 40, a, 4},
		{[Reels]Symbol{c, c, c, c, c}, 200, c, 5},
		{[Reels]Symbol{c, c, b, c, c}, 2, c, 2},
		{[Reels]Symbol{c, p, c, c, c}, 0, 0, 0},
		{[Reels]Symbol{p, p, s, p, p}, 0, 0, 0},
		{[Reels]Symbol{p, s, s, s, s}, 0, 0, 0},

		// Wilds stand in for the symbol the line starts with, and double the
		// win however many of them are in it.
		{[Reels]Symbol{b, w, b, p, p}, 50, b, 3},
		{[Reels]Symbol{b, w, w, w, b}, 1000, b, 5},
		{[Reels]Symbol{b, w, w, w, w}, 1000, b, 5},
		{[Reels]Symbol{c, w, p, p, p}, 4, c, 2},
		{[Reels]Symbol{p, w, c, c, c}, 0, 0, 0},
		{[Reels]Symbol{w, w, s, s, p}, 400, s, 4},
		{[Reels]Symbol{w, w, w, w, w}, 0, 0, 0},
		{[Reels]Symbol{b, b, b, p, w}, 25, b, 3}, // a wild outside the win

		// The ten, cherry, seven and dollar pay for a pair.
		{[Reels]Symbol{t0, t0, a, t0, t0}, 2, t0, 2},
		{[Reels]Symbol{s, s, p, p, p}, 2, s, 2},
		{[Reels]Symbol{d, d, p, p, p}, 5, d, 2},
		{[Reels]Symbol{d, w, p, p, p}, 10, d, 2},
		{[Reels]Symbol{a, a, p, p, p}, 0, 0, 0},

		// Scatters and chips never win along a line, with or without wilds.
		{[Reels]Symbol{x, x, x, x, x}, 0, 0, 0},
		{[Reels]Symbol{k, k, k, k, k}, 0, 0, 0},
		{[Reels]Symbol{x, w, w, w, w}, 0, 0, 0},
		{[Reels]Symbol{w, w, k, k, k}, 0, 0, 0},
		{[Reels]Symbol{b, b, x, b, b}, 0, 0, 0},
	}
	for _, tt := range tests {
		pay, sym, matched := Evaluate(tt.line)
		if pay != tt.pay || sym != tt.sym || matched != tt.matched {
			t.Errorf("Evaluate(%v) = %d, %v, %d; want %d, %v, %d",
				tt.line, pay, sym, matched, tt.pay, tt.sym, tt.matched)
		}
	}
}

func TestStrips(t *testing.T) {
	for reel, strip := range Strips {
		n := len(strip)
		stacks := 0
		for i, sym := range strip {
			if sym == Link {
				// Chips come in one stack of three.
				if strip[(i+n-1)%n] != Link {
					stacks++
					if strip[(i+1)%n] != Link || strip[(i+2)%n] != Link || strip[(i+3)%n] == Link {
						t.Errorf("reel %d: the chips at %d are not a stack of three", reel+1, i)
					}
				}
				continue
			}
			if sym == strip[(i+1)%n] || sym == strip[(i+2)%n] {
				t.Errorf("reel %d: %v comes twice within three stops at %d", reel+1, sym, i)
			}
		}
		if stacks != 1 {
			t.Errorf("reel %d has %d stacks of chips, want 1", reel+1, stacks)
		}
		if count(reel, Scatter) == 0 {
			t.Errorf("reel %d has no scatter", reel+1)
		}
	}
	if n := count(0, Wild); n != 0 {
		t.Errorf("reel 1 has %d wilds, want none", n)
	}
	for reel := 1; reel < Reels; reel++ {
		if count(reel, Wild) == 0 {
			t.Errorf("reel %d has no wild", reel+1)
		}
	}
}

func TestRTP(t *testing.T) {
	// The return is quoted at every line and one credit a line.
	bet := len(Paylines)
	r := ReturnAt(bet)
	t.Logf("return to player %.2f%% at a bet of %d: lines %.2f%%, dice %.2f%%, link %.2f%%, major %.2f%%, grand %.2f%%, pot growth %.2f%%",
		r.Total()*100, bet, r.Lines*100, r.Dice*100, r.Link*100, r.Major*100, r.Grand*100, r.Growth*100)
	t.Logf("free spins 1 in %.0f, link feature 1 in %.0f", 1/FreeSpinOdds(), 1/LinkOdds())
	t.Logf("return at a bet of 1: %.2f%%; at %d: %.2f%%", RTP(1)*100, MaxBet, RTP(MaxBet)*100)
	if got := r.Total(); math.Abs(got-0.92) > 0.005 {
		t.Errorf("RTP(%d) = %v, want about 0.92", bet, got)
	}
	// Only the grand pot should make the return depend on the bet.
	small, large := ReturnAt(1), ReturnAt(MaxBet)
	small.Grand, large.Grand = 0, 0
	if math.Abs(small.Total()-large.Total()) > 1e-4 {
		t.Errorf("return without the grand pot is %v at the smallest bet and %v at the largest", small.Total(), large.Total())
	}
}

// lineReturn works the odds out from symbol counts. Check it against every
// combination of stops on one payline.
func TestLineReturn(t *testing.T) {
	total, combos := 0, 0
	var line [Reels]Symbol
	for _, line[0] = range Strips[0] {
		for _, line[1] = range Strips[1] {
			for _, line[2] = range Strips[2] {
				for _, line[3] = range Strips[3] {
					for _, line[4] = range Strips[4] {
						pay, _, _ := Evaluate(line)
						total += pay
						combos++
					}
				}
			}
		}
	}
	want := float64(total) / float64(combos)
	if got := lineReturn(); math.Abs(got-want) > 1e-9 {
		t.Errorf("lineReturn() = %v, want %v", got, want)
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

// The result of every spin should agree with the grid it reports, line by
// line, and free spins should cost nothing and pay double.
func TestSpinAccounting(t *testing.T) {
	m := New(1, 2)
	freeSpins, links := 0, 0
	for n := range 100000 {
		if m.FreeSpins == 0 {
			m.Credits = StartCredits
			m.Lines = LineChoices[n%len(LineChoices)]
			m.BetPerLine = BetChoices[n/len(LineChoices)%len(BetChoices)]
		}
		before, waiting := m.Credits, m.FreeSpins
		major, grand := m.majorPot, m.grandPot

		r, ok := m.Spin()
		if !ok {
			t.Fatalf("Spin refused with %d credits and bet %d", before, m.TotalBet())
		}
		cost, multiplier := m.TotalBet(), 1
		if waiting > 0 {
			cost, multiplier = 0, FreeSpinMultiplier
			freeSpins++
		}
		if r.Free != (waiting > 0) {
			t.Fatalf("Free = %v with %d free spins waiting", r.Free, waiting)
		}
		if m.majorPot != major+cost*majorGrowth || m.grandPot != grand+cost*grandGrowth {
			t.Fatalf("pots grew by %d and %d on a spin costing %d", m.majorPot-major, m.grandPot-grand, cost)
		}

		scatters, chips := 0, 0
		for i, stop := range r.Stops {
			if r.Grid[i] != Window(i, stop) {
				t.Fatalf("reel %d grid %v does not match stop %d", i, r.Grid[i], stop)
			}
			for row, s := range r.Grid[i] {
				switch s {
				case Scatter:
					scatters++
				case Link:
					chips++
					if c := r.Chips[i][row]; c.Jackpot == NoJackpot && c.Bets == 0 {
						t.Fatalf("the chip at reel %d row %d carries nothing", i, row)
					}
				}
			}
		}

		var wins []LineWin
		total := 0
		for l := range m.Lines {
			var line [Reels]Symbol
			for i, row := range Paylines[l] {
				line[i] = r.Grid[i][row]
			}
			if pay, sym, matched := Evaluate(line); pay > 0 {
				win := pay * m.BetPerLine * multiplier
				wins = append(wins, LineWin{l, sym, matched, pay > Pays[sym][matched], win})
				total += win
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

		scatterWin := ScatterPays[scatters] * m.TotalBet() * multiplier
		total += scatterWin
		if r.Scatters != scatters || r.ScatterWin != scatterWin || r.FreeSpinsWon != FreeSpinAwards[scatters] {
			t.Fatalf("scatters %d paying %d with %d free spins; want %d, %d, %d",
				r.Scatters, r.ScatterWin, r.FreeSpinsWon, scatters, scatterWin, FreeSpinAwards[scatters])
		}
		if r.Win != total || m.Credits != before-cost+total {
			t.Fatalf("win %d and credits %d, want %d and %d", r.Win, m.Credits, total, before-cost+total)
		}
		if want := max(waiting-1, 0) + r.FreeSpinsWon; m.FreeSpins != want {
			t.Fatalf("%d free spins waiting, want %d", m.FreeSpins, want)
		}

		// Six or more chips start the link feature, which must be played
		// out before the next spin.
		if r.LinkStarted != (chips >= LinkTrigger) || (m.Link != nil) != r.LinkStarted {
			t.Fatalf("%d chips showing: LinkStarted = %v, feature in progress = %v", chips, r.LinkStarted, m.Link != nil)
		}
		if r.LinkStarted {
			links++
			if m.Link.Count() != chips || m.Link.Bet != m.TotalBet() {
				t.Fatalf("feature holds %d chips at bet %d, want %d at %d", m.Link.Count(), m.Link.Bet, chips, m.TotalBet())
			}
			if _, ok := m.Spin(); ok {
				t.Fatal("Spin succeeded during a link feature")
			}
			credits := m.Credits
			end := playLink(t, m)
			if m.Link != nil || m.Credits != credits+end.Win {
				t.Fatalf("feature paid %d: credits went from %d to %d", end.Win, credits, m.Credits)
			}
		}
	}
	if freeSpins == 0 || links == 0 {
		t.Errorf("played %d free spins and %d link features; want some of each", freeSpins, links)
	}
}

// A long run of real spins on all lines, features included, should pay back
// close to the calculated return and trigger the features about as often as
// calculated. The two growing pots are left out: they pay too rarely for
// any run to settle near their average.
func TestSpinReturn(t *testing.T) {
	const spins = 4_000_000
	m := New(3, 4)
	m.Lines = len(Paylines)
	bet, won, paid, frees, links := 0, 0, 0, 0, 0
	played := 0 // spins of either kind
	for paid < spins {
		m.Credits = StartCredits
		r, _ := m.Spin()
		played++
		if !r.Free {
			paid++
			bet += m.TotalBet()
			if r.FreeSpinsWon > 0 {
				frees++
			}
		}
		won += r.Win
		if r.LinkStarted {
			links++
			end := playLink(t, m)
			won += end.Win - end.MajorWin - end.GrandWin
		}
	}
	want := ReturnAt(m.TotalBet())
	got := float64(won) / float64(bet)
	freeOdds := float64(frees) / spins
	linkOdds := float64(links) / float64(played)
	t.Logf("paid back %.2f%% over %d spins of %d lines; free spins 1 in %.1f, link feature 1 in %.1f",
		got*100, spins, m.Lines, 1/freeOdds, 1/linkOdds)
	if w := want.Lines + want.Dice + want.Link; math.Abs(got-w) > 0.015 {
		t.Errorf("paid back %v, want about %v", got, w)
	}
	if w := FreeSpinOdds(); math.Abs(freeOdds-w) > w/20 {
		t.Errorf("free spins on %v of spins, want about %v", freeOdds, w)
	}
	if w := LinkOdds(); math.Abs(linkOdds-w) > w/20 {
		t.Errorf("link feature on %v of spins, want about %v", linkOdds, w)
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

	// A free spin needs no credits.
	m.Credits, m.FreeSpins = 0, 1
	if r, ok := m.Spin(); !ok || !r.Free {
		t.Error("free spin refused with no credits")
	}
}

func TestInsertCoin(t *testing.T) {
	m := New(1, 2)
	m.Credits = 0
	m.InsertCoin()
	m.InsertCoin()
	if m.Credits != 2*CoinCredits {
		t.Errorf("credits = %d after two coins, want %d", m.Credits, 2*CoinCredits)
	}
	if _, ok := m.Spin(); !ok {
		t.Error("Spin refused after inserting coins")
	}
}
