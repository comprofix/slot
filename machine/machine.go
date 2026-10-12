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
	MaxBet       = 3
	StartCredits = 100
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

// Pays is the payout per credit bet, by symbol and by how many of it run
// along the payline from the left reel. Only cherries pay for fewer than
// three.
var Pays = [NumSymbols][Reels + 1]int{
	Cherry: {0, 1, 2, 5, 15, 50},
	Plum:   {0, 0, 0, 4, 20, 75},
	Bell:   {0, 0, 0, 8, 40, 150},
	Bar:    {0, 0, 0, 18, 75, 300},
	Seven:  {0, 0, 0, 30, 150, 750},
	Dollar: {0, 0, 0, 100, 500, 5000},
}

// Evaluate returns the payout per credit bet for a payline, and how many
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
// it rests at the given stop. The middle one is on the payline.
func Window(reel, stop int) [3]Symbol {
	s := Strips[reel]
	n := len(s)
	return [3]Symbol{s[(stop+1)%n], s[stop], s[(stop-1+n)%n]}
}

// RTP is the long-run return to player: the fraction of money bet that the
// machine pays back, worked out by trying every combination of stops.
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

// Result describes one spin.
type Result struct {
	Stops   [Reels]int
	Line    [Reels]Symbol
	Matched int // winning reels, counted from the left
	Win     int // credits won
}

// Machine is the state of a game in progress.
type Machine struct {
	Credits int
	Bet     int
	Stops   [Reels]int
	rng     *rand.Rand
}

// New returns a machine with the starting credits. The seeds fix the
// sequence of spins, which the tests rely on.
func New(seed1, seed2 uint64) *Machine {
	return &Machine{
		Credits: StartCredits,
		Bet:     1,
		rng:     rand.New(rand.NewPCG(seed1, seed2)),
	}
}

// SetBet changes the bet, keeping it between 1 and MaxBet and no more than
// the credits left.
func (m *Machine) SetBet(bet int) {
	bet = min(bet, MaxBet, m.Credits)
	m.Bet = max(bet, 1)
}

// Spin takes the bet, stops each reel at random and pays any win. It
// reports false and changes nothing if there are too few credits.
func (m *Machine) Spin() (Result, bool) {
	if m.Credits < m.Bet {
		return Result{}, false
	}
	m.Credits -= m.Bet

	var r Result
	for i := range Reels {
		r.Stops[i] = m.rng.IntN(len(Strips[i]))
		r.Line[i] = Strips[i][r.Stops[i]]
	}
	pay, matched := Evaluate(r.Line)
	r.Matched = matched
	r.Win = pay * m.Bet

	m.Stops = r.Stops
	m.Credits += r.Win
	if m.Credits > 0 {
		m.SetBet(m.Bet)
	}
	return r, true
}
