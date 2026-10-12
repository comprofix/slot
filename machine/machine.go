// Package machine holds the slot machine rules: reel strips, paylines, the
// paytable, free spins and credits. It has no graphics so it can be tested
// on its own.
package machine

import (
	"math/rand/v2"
	"slices"
)

// Symbol is one picture on a reel.
type Symbol int

// The paying symbols come first, from the lowest paying to the highest,
// followed by the special symbols.
const (
	Ten Symbol = iota
	Jack
	Queen
	King
	Ace
	Cherry
	Plum
	Bell
	Bar
	Seven
	Dollar
	Wild    // joker: stands in for any paying symbol and doubles the win
	Scatter // dice: three or more anywhere award free spins
	Link    // poker chip: will start the link feature, which is not built yet
	NumSymbols
)

func (s Symbol) String() string {
	return [...]string{
		"Ten", "Jack", "Queen", "King", "Ace", "Cherry", "Plum", "Bell",
		"Bar", "Seven", "Dollar", "Wild", "Scatter", "Link",
	}[s]
}

// Paying reports whether s is an ordinary symbol that wins along paylines.
func (s Symbol) Paying() bool {
	return s <= Dollar
}

const (
	Reels        = 5
	Rows         = 3 // symbols showing on each reel
	StartCredits = 1000
	CoinCredits  = 100 // credits added by inserting a coin
)

// Short names keep the strips below readable.
const (
	tn = Ten
	jk = Jack
	qn = Queen
	kg = King
	ac = Ace
	ch = Cherry
	pl = Plum
	be = Bell
	ba = Bar
	sv = Seven
	do = Dollar
	wd = Wild
	sc = Scatter
	lk = Link
)

// Strips is the order of symbols on each reel. How often a symbol appears
// on a strip sets how likely it is to land, so changing these changes the
// odds; TestRTP reports the resulting return to player.
//
// The wild is kept off the first reel, so it can never start a win. No
// symbol comes twice within three stops, so a reel never shows two of the
// same.
var Strips = [Reels][]Symbol{
	{tn, ch, jk, kg, ch, jk, lk, kg, pl, qn, ac, ch, kg, tn, sv, pl, qn, ac, be, ba, qn, kg, tn, sc, do, tn, lk, ch, jk, tn, ac, be, ba, ch, qn, jk},
	{jk, tn, pl, do, sc, be, kg, sv, jk, lk, ac, qn, wd, jk, kg, ch, ba, qn, tn, kg, ac, tn, ch, sv, tn, pl, qn, wd, jk, sc, ch, be, ac, qn, kg, lk},
	{sv, kg, lk, jk, be, qn, jk, tn, ch, qn, sc, ch, ac, qn, kg, tn, jk, ac, kg, jk, do, ac, ba, kg, qn, pl, tn, jk, kg, ch, qn, jk, wd, be, pl, lk},
	{ch, lk, pl, tn, jk, kg, ba, be, wd, sc, ac, qn, do, sv, ch, qn, ac, ch, pl, ba, tn, kg, jk, ac, tn, ba, wd, ac, sv, lk, sc, qn, jk, kg, pl, be},
	{wd, sv, pl, ac, ch, jk, ba, ch, wd, pl, be, jk, qn, pl, be, sc, qn, do, tn, ba, ac, kg, lk, qn, jk, be, kg, ac, tn, sv, ch, kg, tn, ba, do, lk},
}

// Row numbers within a reel's window.
const (
	Top = iota
	Middle
	Bottom
)

// Paylines gives, for each payline, the row it crosses on each reel from
// left to right. A player choosing N lines plays the first N.
var Paylines = [...][Reels]int{
	{1, 1, 1, 1, 1},
	{0, 0, 0, 0, 0},
	{2, 2, 2, 2, 2},
	{2, 1, 0, 1, 2},
	{0, 1, 2, 1, 0},
	{1, 2, 1, 0, 1},
	{1, 2, 1, 2, 1},
	{0, 1, 0, 1, 2},
	{2, 1, 2, 1, 0},
	{2, 1, 0, 1, 0},
	{0, 1, 2, 1, 2},
	{1, 0, 1, 2, 1},
	{2, 1, 2, 1, 2},
	{1, 0, 1, 0, 1},
	{0, 1, 0, 1, 0},
	{1, 1, 2, 1, 1},
	{0, 0, 1, 0, 0},
	{2, 2, 1, 2, 2},
	{1, 1, 0, 1, 1},
	{0, 0, 2, 0, 0},
	{2, 2, 0, 2, 2},
	{1, 1, 1, 2, 1},
	{0, 0, 0, 1, 2},
	{2, 2, 2, 1, 0},
	{2, 1, 1, 1, 0},
}

// The numbers of lines and the bets per line a player can choose from.
var (
	LineChoices = []int{1, 5, 10, 15, 25}
	BetChoices  = []int{1, 2, 5, 10, 20}
)

// Pays is the payout per credit bet on a line, by symbol and by how many of
// it run along the payline from the left reel. A few symbols pay for two;
// nothing pays for one.
var Pays = [NumSymbols][Reels + 1]int{
	Ten:    {0, 0, 2, 5, 20, 100},
	Jack:   {0, 0, 0, 5, 20, 100},
	Queen:  {0, 0, 0, 5, 25, 100},
	King:   {0, 0, 0, 5, 25, 125},
	Ace:    {0, 0, 0, 10, 40, 150},
	Cherry: {0, 0, 2, 10, 50, 200},
	Plum:   {0, 0, 0, 15, 75, 250},
	Bell:   {0, 0, 0, 20, 100, 400},
	Bar:    {0, 0, 0, 25, 125, 500},
	Seven:  {0, 0, 2, 40, 200, 1000},
	Dollar: {0, 0, 5, 100, 500, 5000},
}

// Scatters pay wherever they land, not along paylines. By how many are
// showing, ScatterPays is the prize per credit of the total bet and
// FreeSpinAwards the free spins won.
var (
	ScatterPays    = [Reels + 1]int{3: 5, 4: 20, 5: 400}
	FreeSpinAwards = [Reels + 1]int{3: 15, 4: 20, 5: 25}
)

// FreeSpinMultiplier multiplies every win during free spins.
const FreeSpinMultiplier = 2

// WildMultiplier multiplies a line win that a wild helps to make. Several
// wilds in one win still multiply it only once.
const WildMultiplier = 2

// Evaluate returns the payout per credit bet for the symbols along one
// payline, the symbol that won, and how many reels (counted from the left)
// make up the win. Wilds count as the symbol the line starts with, and a
// win with a wild in it is multiplied by WildMultiplier.
func Evaluate(line [Reels]Symbol) (pay int, sym Symbol, matched int) {
	// The win belongs to the first symbol that is not a wild.
	sym = Wild
	for _, s := range line {
		if s != Wild {
			sym = s
			break
		}
	}
	if !sym.Paying() {
		return 0, 0, 0
	}
	run, wild := 0, false
	for run < Reels && (line[run] == sym || line[run] == Wild) {
		if line[run] == Wild {
			wild = true
		}
		run++
	}
	pay = Pays[sym][run]
	if pay == 0 {
		return 0, 0, 0
	}
	if wild {
		pay *= WildMultiplier
	}
	return pay, sym, run
}

// Window returns the three symbols showing on a reel, top to bottom, when
// it rests at the given stop.
func Window(reel, stop int) [Rows]Symbol {
	s := Strips[reel]
	n := len(s)
	return [Rows]Symbol{s[(stop+1)%n], s[stop], s[(stop-1+n)%n]}
}

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

// scatterOdds returns the chance of each number of scatters showing.
func scatterOdds() [Reels + 1]float64 {
	odds := [Reels + 1]float64{1}
	for reel := range Reels {
		// The strips keep scatters apart, so a reel shows one or none.
		showing := 0
		for stop := range Strips[reel] {
			w := Window(reel, stop)
			if slices.Contains(w[:], Scatter) {
				showing++
			}
		}
		p := float64(showing) / float64(len(Strips[reel]))
		for k := reel + 1; k > 0; k-- {
			odds[k] = odds[k]*(1-p) + odds[k-1]*p
		}
		odds[0] *= 1 - p
	}
	return odds
}

// FreeSpinOdds is the chance that a spin awards free spins.
func FreeSpinOdds() float64 {
	total := 0.0
	for k, p := range scatterOdds() {
		if FreeSpinAwards[k] > 0 {
			total += p
		}
	}
	return total
}

// RTP is the long-run return to player: the fraction of money bet that the
// machine pays back, counting line wins, scatter prizes and free spins.
func RTP() float64 {
	// What one spin pays before any free spins it awards, and how many free
	// spins it awards on average.
	spin, awarded := lineReturn(), 0.0
	for k, p := range scatterOdds() {
		spin += p * float64(ScatterPays[k])
		awarded += p * float64(FreeSpinAwards[k])
	}
	// Free spins can award more free spins, so one awarded spin leads to
	// 1/(1-awarded) spins played on average.
	return spin * (1 + FreeSpinMultiplier*awarded/(1-awarded))
}

// LinkPrize is what a link chip carries: one of the jackpots, or a number
// of total bets.
type LinkPrize struct {
	Jackpot string // "MINI", "MINOR", "MAJOR", "GRAND", or empty
	Bets    int
}

// linkPrizes are dealt out to the chips on the strips in turn.
var linkPrizes = []LinkPrize{
	{Bets: 1}, {Bets: 2}, {Jackpot: "MINI"}, {Bets: 1}, {Bets: 5},
	{Jackpot: "MINOR"}, {Bets: 3}, {Bets: 1}, {Jackpot: "MAJOR"}, {Bets: 10},
	{Bets: 2}, {Jackpot: "GRAND"},
}

// LinkPrizeAt is the prize shown on the chip at a reel stop. The chips do
// nothing yet: the link feature they belong to has not been built.
func LinkPrizeAt(reel, stop int) LinkPrize {
	return linkPrizes[(reel*5+stop)%len(linkPrizes)]
}

// LineWin is a win on one payline.
type LineWin struct {
	Line    int    // index into Paylines
	Symbol  Symbol // the winning symbol
	Matched int    // winning reels, counted from the left
	Doubled bool   // a wild helped, multiplying the win
	Win     int    // credits won
}

// Result describes one spin.
type Result struct {
	Stops [Reels]int
	Grid  [Reels][Rows]Symbol // what each reel shows, top to bottom
	Wins  []LineWin           // winning paylines, in payline order

	Free         bool // this was a free spin
	Scatters     int  // scatters showing
	ScatterWin   int  // credits won by the scatters
	FreeSpinsWon int  // free spins this spin awarded

	Win int // total credits won
}

// Machine is the state of a game in progress.
type Machine struct {
	Credits    int
	Lines      int // paylines being played
	BetPerLine int
	FreeSpins  int // free spins waiting to be played
	Stops      [Reels]int
	rng        *rand.Rand
}

// New returns a machine with the starting credits. The seeds fix the
// sequence of spins, which the tests rely on.
func New(seed1, seed2 uint64) *Machine {
	return &Machine{
		Credits:    StartCredits,
		Lines:      1,
		BetPerLine: 1,
		rng:        rand.New(rand.NewPCG(seed1, seed2)),
	}
}

// TotalBet is the cost of one spin.
func (m *Machine) TotalBet() int {
	return m.Lines * m.BetPerLine
}

// Spin stops each reel at random and pays the wins on every payline being
// played, along with any scatter prize. It plays a free spin if any are
// waiting; otherwise it takes the bet, and reports false and changes
// nothing if there are too few credits.
//
// Lines and BetPerLine must not be changed while free spins are waiting:
// they are played at the bet that won them.
func (m *Machine) Spin() (Result, bool) {
	var r Result
	multiplier := 1
	if m.FreeSpins > 0 {
		m.FreeSpins--
		r.Free = true
		multiplier = FreeSpinMultiplier
	} else {
		if m.Credits < m.TotalBet() {
			return Result{}, false
		}
		m.Credits -= m.TotalBet()
	}

	for i := range Reels {
		r.Stops[i] = m.rng.IntN(len(Strips[i]))
		r.Grid[i] = Window(i, r.Stops[i])
		for _, s := range r.Grid[i] {
			if s == Scatter {
				r.Scatters++
			}
		}
	}
	for l := range m.Lines {
		var line [Reels]Symbol
		for i, row := range Paylines[l] {
			line[i] = r.Grid[i][row]
		}
		if pay, sym, matched := Evaluate(line); pay > 0 {
			win := pay * m.BetPerLine * multiplier
			doubled := pay > Pays[sym][matched]
			r.Wins = append(r.Wins, LineWin{Line: l, Symbol: sym, Matched: matched, Doubled: doubled, Win: win})
			r.Win += win
		}
	}
	r.ScatterWin = ScatterPays[r.Scatters] * m.TotalBet() * multiplier
	r.Win += r.ScatterWin
	r.FreeSpinsWon = FreeSpinAwards[r.Scatters]

	m.Stops = r.Stops
	m.Credits += r.Win
	m.FreeSpins += r.FreeSpinsWon
	return r, true
}

// InsertCoin adds credits, so play can go on after they run out.
func (m *Machine) InsertCoin() {
	m.Credits += CoinCredits
}
