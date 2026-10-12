// Package machine holds the slot machine rules: reel strips, paylines, the
// paytable, free spins, the link feature and credits. It has no graphics so
// it can be tested on its own.
package machine

import (
	"math/rand/v2"
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
	Link    // poker chip: six or more anywhere start the link feature
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
// The wild is kept off the first reel, so it can never start a win. The link
// chips sit together in a stack of three, so a reel can show a column of
// them. No other symbol comes twice within three stops, so a reel never
// shows two of the same.
var Strips = [Reels][]Symbol{
	{jk, pl, qn, ac, be, jk, kg, be, sv, kg, ch, ac, qn, ch, ba, ac, jk, tn, kg, sc, qn, ba, tn, sc, jk, tn, sv, ch, do, lk, lk, lk, kg, pl, ac, tn, sv, ch, pl, ba, tn, jk, qn, kg},
	{jk, kg, wd, ch, qn, ac, ch, tn, jk, ch, qn, pl, ch, ac, kg, tn, jk, sc, tn, kg, ch, jk, kg, sv, jk, tn, ac, be, do, tn, ba, sv, qn, pl, ch, qn, kg, wd, lk, lk, lk, sc, tn, be},
	{qn, jk, sv, ac, qn, kg, jk, ac, kg, jk, tn, kg, qn, pl, jk, qn, kg, ch, qn, pl, ch, be, kg, jk, qn, sc, be, ba, tn, do, ch, ba, wd, ac, qn, be, wd, kg, ac, jk, lk, lk, lk, tn},
	{jk, pl, be, qn, ac, tn, ch, ac, be, qn, do, pl, jk, ba, qn, ac, sc, pl, be, wd, ac, do, lk, lk, lk, jk, tn, kg, qn, ba, kg, jk, wd, ch, tn, pl, ch, ba, ac, sv, qn, kg, sc, be},
	{sc, ch, qn, pl, tn, ac, kg, wd, ac, jk, ch, ac, be, tn, qn, do, jk, sc, do, sv, be, ch, ac, jk, ba, qn, kg, jk, ba, qn, tn, be, ac, qn, kg, ba, pl, wd, lk, lk, lk, pl, qn, be},
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
	ScatterPays    = [Reels + 1]int{3: 2, 4: 10, 5: 60}
	FreeSpinAwards = [Reels + 1]int{3: 8, 4: 12, 5: 20}
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

	Chips       [Reels][Rows]Chip // what each link chip showing carries
	LinkStarted bool              // enough chips landed to start the link feature

	Win int // total credits won
}

// Machine is the state of a game in progress.
type Machine struct {
	Credits    int
	Lines      int // paylines being played
	BetPerLine int
	FreeSpins  int      // free spins waiting to be played
	Link       *Feature // the link feature in progress, if there is one
	Stops      [Reels]int

	majorPot, grandPot int // in thousandths of a credit
	rng                *rand.Rand
}

// New returns a machine with the starting credits. The seeds fix the
// sequence of spins, which the tests rely on.
func New(seed1, seed2 uint64) *Machine {
	return &Machine{
		Credits:    StartCredits,
		Lines:      1,
		BetPerLine: 1,
		majorPot:   MajorStart * potUnit,
		grandPot:   GrandStart * potUnit,
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
// nothing if there are too few credits. It also reports false while a link
// feature is in progress: that is played out with Respin.
//
// Lines and BetPerLine must not be changed while free spins are waiting or
// a link feature is in progress: they are played at the bet that won them.
func (m *Machine) Spin() (Result, bool) {
	if m.Link != nil {
		return Result{}, false
	}
	var r Result
	bet := m.TotalBet()
	multiplier := 1
	if m.FreeSpins > 0 {
		m.FreeSpins--
		r.Free = true
		multiplier = FreeSpinMultiplier
	} else {
		if m.Credits < bet {
			return Result{}, false
		}
		m.Credits -= bet
		// A cut of every bet goes into the two jackpot pots.
		m.majorPot += bet * majorGrowth
		m.grandPot += bet * grandGrowth
	}

	link := Feature{Bet: bet, Respins: LinkRespins}
	for i := range Reels {
		r.Stops[i] = m.rng.IntN(len(Strips[i]))
		r.Grid[i] = Window(i, r.Stops[i])
		for row, s := range r.Grid[i] {
			switch s {
			case Scatter:
				r.Scatters++
			case Link:
				link.Chips[i][row] = m.dealChip(bet, !link.hasMajor())
				link.Held[i][row] = true
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
	r.ScatterWin = ScatterPays[min(r.Scatters, Reels)] * bet * multiplier
	r.Win += r.ScatterWin
	r.FreeSpinsWon = FreeSpinAwards[min(r.Scatters, Reels)]

	// The chips pay nothing by themselves, but enough of them lock in
	// place and start the link feature.
	r.Chips = link.Chips
	if link.Count() >= LinkTrigger {
		r.LinkStarted = true
		m.Link = &link
	}

	m.Stops = r.Stops
	m.Credits += r.Win
	m.FreeSpins += r.FreeSpinsWon
	return r, true
}

// InsertCoin adds credits, so play can go on after they run out.
func (m *Machine) InsertCoin() {
	m.Credits += CoinCredits
}
