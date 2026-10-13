package machine

// The long-run odds of the machine, worked out from the reel strips and the
// rules rather than by playing. The tests check them against real spins.

// count is how many times sym is on a reel's strip.
func count(reel int, sym Symbol) int {
	n := 0
	for _, s := range Strips[reel] {
		if s == sym {
			n++
		}
	}
	return n
}

// lineReturn is the fraction of the bet on a line that payline wins pay
// back. Each reel stops independently, so every payline has the same odds
// and the figure holds however many lines are played.
func lineReturn() float64 {
	total := 0.0
	for sym := Ten; sym <= Dollar; sym++ {
		// lead is the chance the line starts with a run of at least n, and
		// plain the chance it does so without the help of a wild.
		lead := float64(count(0, sym)) / float64(len(Strips[0]))
		plain := lead
		for n := 1; n <= Reels; n++ {
			// The run goes on if the next reel shows the symbol or a wild.
			var symNext, wildNext float64
			if n < Reels {
				stops := float64(len(Strips[n]))
				symNext = float64(count(n, sym)) / stops
				wildNext = float64(count(n, Wild)) / stops
			}
			ends := 1 - symNext - wildNext
			total += float64(Pays[sym][n]) * ends * (plain + WildMultiplier*(lead-plain))
			lead *= symNext + wildNext
			plain *= symNext
		}
	}
	return total
}

// showing returns the chance of each number of sym showing on the reels
// after a spin.
func showing(sym Symbol) [Cells + 1]float64 {
	odds := [Cells + 1]float64{1}
	for reel := range Reels {
		// The chance of each number showing on this reel alone.
		var here [Rows + 1]float64
		for stop := range Strips[reel] {
			n := 0
			for _, s := range Window(reel, stop) {
				if s == sym {
					n++
				}
			}
			here[n] += 1 / float64(len(Strips[reel]))
		}
		var next [Cells + 1]float64
		for k, p := range odds {
			for n, q := range here {
				if k+n <= Cells {
					next[k+n] += p * q
				}
			}
		}
		odds = next
	}
	return odds
}

// FreeSpinOdds is the chance that a spin awards free spins.
func FreeSpinOdds() float64 {
	total := 0.0
	for k, p := range showing(Scatter) {
		if FreeSpinAwards[min(k, Reels)] > 0 {
			total += p
		}
	}
	return total
}

// LinkOdds is the chance that a spin starts the link feature.
func LinkOdds() float64 {
	total := 0.0
	for k, p := range showing(Link) {
		if k >= LinkTrigger {
			total += p
		}
	}
	return total
}

// linkEnd works out how a link feature ends from a given position: the
// number of chips it finishes with on average, and the chance it fills the
// grid. grand is the chance that a chip completing the grid is let land.
func linkEnd(filled, respins int, grand float64, memo map[[2]int][2]float64) (chips, full float64) {
	if filled == Cells {
		return Cells, 1
	}
	if respins == 0 {
		return float64(filled), 0
	}
	key := [2]int{filled, respins}
	if v, ok := memo[key]; ok {
		return v[0], v[1]
	}

	empty := Cells - filled
	odds := respinOdds
	// after adds in what follows when n chips land, weighted by chance p.
	after := func(n int, p float64) {
		var c, f float64
		if n == 0 {
			c, f = linkEnd(filled, respins-1, grand, memo)
		} else {
			c, f = linkEnd(filled+n, LinkRespins, grand, memo)
		}
		chips += p * c
		full += p * f
	}
	for n := 0; n <= empty; n++ {
		p := binomial(empty, n, odds)
		if n == empty {
			// The chip that would complete the grid may be held back.
			after(n, p*grand)
			after(n-1, p*(1-grand))
			continue
		}
		after(n, p)
	}
	memo[key] = [2]float64{chips, full}
	return chips, full
}

// binomial is the chance of exactly k successes in n tries, each with
// chance p.
func binomial(n, k int, p float64) float64 {
	c := 1.0
	for i := range k {
		c = c * float64(n-i) / float64(i+1) * p
	}
	for range n - k {
		c *= 1 - p
	}
	return c
}

// Returns breaks the long-run return to player down by where it comes
// from. Each part is a fraction of the money bet.
type Returns struct {
	Lines  float64 // line wins on paid spins
	Dice   float64 // scatter prizes, and everything free spins win on lines
	Link   float64 // link chips, with the mini and minor jackpots
	Major  float64 // the major pot's starting amount
	Grand  float64 // the grand pot's starting amount
	Growth float64 // what play adds to the two pots, all paid out in time
}

// Total is the whole return to player.
func (r Returns) Total() float64 {
	return r.Lines + r.Dice + r.Link + r.Major + r.Grand + r.Growth
}

// ReturnAt works out the return to player at a total bet. The bet matters
// only to the grand pot: it pays the same whatever the bet, and although it
// is made harder to win at small bets, that cannot level it completely.
func ReturnAt(bet int) Returns {
	var r Returns

	// What one spin pays on lines and scatters, and how many free spins it
	// awards on average.
	lines := lineReturn()
	scatters, awarded := 0.0, 0.0
	for k, p := range showing(Scatter) {
		scatters += p * float64(ScatterPays[min(k, Reels)])
		awarded += p * float64(FreeSpinAwards[min(k, Reels)])
	}
	// Free spins can award more free spins, so each paid spin leads to
	// 1/(1-awarded) spins in all, the rest of them free.
	spins := 1 / (1 - awarded)
	free := spins - 1
	r.Lines = lines
	r.Dice = scatters + FreeSpinMultiplier*free*(lines+scatters)

	// What an ordinary chip is worth on average, in total bets.
	weight, worth := 0, 0
	for _, c := range chipBets {
		weight += c.weight
		worth += c.weight * c.bets
	}

	major := majorOdds * betShare(bet)
	chip := (1 - major) * (miniOdds*MiniBets + minorOdds*MinorBets +
		(1-miniOdds-minorOdds)*float64(worth)/float64(weight))

	memo := map[[2]int][2]float64{}
	for k, p := range showing(Link) {
		if k < LinkTrigger {
			continue
		}
		chips, full := linkEnd(k, LinkRespins, grandChance(bet), memo)
		r.Link += p * spins * chips * chip
		r.Major += p * spins * chips * major * MajorStart / float64(bet)
		r.Grand += p * spins * full * GrandStart / float64(bet)
	}
	r.Growth = float64(majorGrowth+grandGrowth) / potUnit
	return r
}

// RTP is the long-run return to player at a total bet: the fraction of
// money bet that the machine pays back.
func RTP(bet int) float64 {
	return ReturnAt(bet).Total()
}
