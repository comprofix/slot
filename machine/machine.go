// Package machine holds the slot machine rules: reel strips, the paytable,
// spinning and credits. It has no graphics so it can be tested on its own.
package machine

import "math/rand/v2"

// Symbol is one picture on a reel.
type Symbol int

const (
	Cherry Symbol = iota
	Plum
	Bell
	Bar
	Seven
	Dollar
	NumSymbols
)

func (s Symbol) String() string {
	return [...]string{"Cherry", "Plum", "Bell", "Bar", "Seven", "Dollar"}[s]
}

const (
	Reels        = 5
	Rows         = 3 // symbols showing on each reel
	StartCredits = 1000
)

// Short names keep the strips below readable.
const (
	ch = Cherry
	pl = Plum
	be = Bell
	ba = Bar
	sv = Seven
	do = Dollar
)

// Strips is the order of symbols on each reel. How often a symbol appears
// on a strip sets how likely it is to land, so changing these changes the
// odds; TestRTP reports the resulting return to player.
var Strips = [Reels][]Symbol{
	{ch, pl, be, ba, ch, pl, sv, be, ch, pl, ba, do, ch, be, pl, ba, sv, ch, pl, be},
	{pl, ch, ba, be, sv, pl, ch, be, ba, pl, sv, ch, be, pl, do, ba, ch, be, sv, pl},
	{be, ch, ba, pl, do, be, ch, ba, pl, sv, be, ch, ba, pl, do, be, ch, ba, pl, sv},
	{be, pl, sv, ch, ba, pl, be, ch, sv, pl, ba, be, ch, pl, do, sv, be, ch, ba, pl},
	{ch, ba, be, pl, sv, ch, ba, do, be, pl, ch, ba, be, pl, sv, ch, ba, be, do, pl},
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
// it run along the payline from the left reel. Only cherries pay for fewer
// than three.
var Pays = [NumSymbols][Reels + 1]int{
	Cherry: {0, 1, 2, 5, 15, 50},
	Plum:   {0, 0, 0, 4, 20, 75},
	Bell:   {0, 0, 0, 8, 40, 150},
	Bar:    {0, 0, 0, 18, 75, 300},
	Seven:  {0, 0, 0, 30, 150, 750},
	Dollar: {0, 0, 0, 100, 500, 5000},
}

// Evaluate returns the payout per credit bet for the symbols along one
// payline, and how many
// reels (counted from the left) make up the win.
func Evaluate(line [Reels]Symbol) (pay, matched int) {
	run := 1
	for run < Reels && line[run] == line[0] {
		run++
	}
	pay = Pays[line[0]][run]
	if pay == 0 {
		return 0, 0
	}
	return pay, run
}

// Window returns the three symbols showing on a reel, top to bottom, when
// it rests at the given stop.
func Window(reel, stop int) [Rows]Symbol {
	s := Strips[reel]
	n := len(s)
	return [Rows]Symbol{s[(stop+1)%n], s[stop], s[(stop-1+n)%n]}
}

// RTP is the long-run return to player: the fraction of money bet that the
// machine pays back, worked out by trying every combination of stops on
// one payline. Each reel stops independently, so every payline has the
// same odds and the figure holds however many lines are played.
func RTP() float64 {
	total, combos := 0, 0
	var stops [Reels]int
	for {
		var line [Reels]Symbol
		for i, stop := range stops {
			line[i] = Strips[i][stop]
		}
		pay, _ := Evaluate(line)
		total += pay
		combos++

		// Step to the next combination, like an odometer.
		i := Reels - 1
		for ; i >= 0; i-- {
			stops[i]++
			if stops[i] < len(Strips[i]) {
				break
			}
			stops[i] = 0
		}
		if i < 0 {
			break
		}
	}
	return float64(total) / float64(combos)
}

// LineWin is a win on one payline.
type LineWin struct {
	Line    int    // index into Paylines
	Symbol  Symbol // the winning symbol
	Matched int    // winning reels, counted from the left
	Win     int    // credits won
}

// Result describes one spin.
type Result struct {
	Stops [Reels]int
	Grid  [Reels][Rows]Symbol // what each reel shows, top to bottom
	Wins  []LineWin           // winning paylines, in payline order
	Win   int                 // total credits won
}

// Machine is the state of a game in progress.
type Machine struct {
	Credits    int
	Lines      int // paylines being played
	BetPerLine int
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

// Spin takes the bet, stops each reel at random and pays the win on every
// payline being played. It reports false and changes nothing if there are
// too few credits.
func (m *Machine) Spin() (Result, bool) {
	if m.Credits < m.TotalBet() {
		return Result{}, false
	}
	m.Credits -= m.TotalBet()

	var r Result
	for i := range Reels {
		r.Stops[i] = m.rng.IntN(len(Strips[i]))
		r.Grid[i] = Window(i, r.Stops[i])
	}
	for l := range m.Lines {
		var line [Reels]Symbol
		for i, row := range Paylines[l] {
			line[i] = r.Grid[i][row]
		}
		if pay, matched := Evaluate(line); pay > 0 {
			win := pay * m.BetPerLine
			r.Wins = append(r.Wins, LineWin{Line: l, Symbol: line[0], Matched: matched, Win: win})
			r.Win += win
		}
	}

	m.Stops = r.Stops
	m.Credits += r.Win
	return r, true
}
