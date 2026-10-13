package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"slot/machine"
)

// linkPhase is where the link feature has got to on screen.
type linkPhase int

const (
	linkPause linkPhase = iota // showing the chips held, before a respin
	linkSpin                   // the empty cells are spinning
	linkDone                   // paid; the grid stays up until the next spin
)

// How long each phase lasts, in ticks.
const (
	linkIntroTicks = 150 // after the chips that started the feature land
	linkPauseTicks = 45
	linkSpinTicks  = 50
	linkFlashTicks = 40 // how long a newly landed chip is picked out
)

// linkView is the link feature as it appears on screen. It runs a step
// behind the machine: a respin is decided when the cells start to spin and
// shown when they stop.
type linkView struct {
	show  bool // the held chips are on screen in place of the reels
	phase linkPhase
	timer int

	state  machine.Feature                  // the chips on screen
	next   machine.Respin                   // the respin being played, shown when it stops
	total  int                              // what the chips on screen are worth
	landed [machine.Reels][machine.Rows]int // when each chip landed
}

// startLink puts the link feature on screen, once the spin that won it has
// stopped.
func (g *Game) startLink() {
	g.link = linkView{
		show:  true,
		timer: linkIntroTicks,
		state: *g.m.Link,
		total: g.m.LinkTotal(),
	}
	g.message = fmt.Sprintf("%d CHIPS! LINK FEATURE", g.m.Link.Count())
	if g.m.Link.Big {
		g.message = "GIANT CHIP! LINK FEATURE"
	}
	g.snd.play(g.snd.bigWin)
}

// updateLink moves the link feature on by a tick. Respins play themselves;
// hurry, or the spin keys, skip the pause before the next one.
func (g *Game) updateLink(hurry bool) {
	if pressed(ebiten.KeyEscape) {
		g.scene = sceneMenu
		return
	}
	hurry = hurry || pressed(ebiten.KeySpace, ebiten.KeyEnter, ebiten.KeyNumpadEnter)

	l := &g.link
	l.timer--
	switch l.phase {
	case linkPause:
		if l.timer > 0 && !hurry {
			return
		}
		l.next, _ = g.m.Respin()
		g.save()
		l.phase = linkSpin
		l.timer = linkSpinTicks
		g.message = ""
		g.snd.play(g.snd.start)

	case linkSpin:
		if l.timer > 0 {
			return
		}
		landed := false
		for i, col := range l.next.New {
			for row, isNew := range col {
				if isNew {
					l.landed[i][row] = g.ticks
					landed = true
				}
			}
		}
		l.state = l.next.Link
		if landed {
			g.snd.play(g.snd.coin)
		} else {
			g.snd.play(g.snd.stop)
		}
		if !l.next.Done {
			l.total = g.m.LinkTotal()
			l.phase = linkPause
			l.timer = linkPauseTicks
			return
		}
		g.finishLink()
	}
}

// finishLink shows what the link feature paid.
func (g *Game) finishLink() {
	end := g.link.next
	g.link.phase = linkDone
	g.link.total = end.Win
	g.stopTick = g.ticks
	g.credits = g.m.Credits
	g.win = end.Win
	if g.free.active {
		g.free.win += end.Win
	}

	g.message = fmt.Sprintf("LINK FEATURE PAYS %d!", end.Win)
	switch {
	case end.GrandWin > 0:
		g.message = fmt.Sprintf("GRAND JACKPOT! LINK FEATURE PAYS %d!", end.Win)
	case end.MajorWin > 0:
		g.message = fmt.Sprintf("MAJOR JACKPOT! LINK FEATURE PAYS %d!", end.Win)
	}
	// A giant chip on the last free game: the free games end here.
	if g.free.active && g.m.FreeSpins == 0 {
		g.free.active = false
		g.message += fmt.Sprintf("   FREE GAMES WON %d", g.free.win)
	}
	g.snd.play(g.snd.bigWin)
}

// drawLinkGrid draws the link feature in place of the reels: the chips
// held, and the empty cells turning while a respin is played.
func (g *Game) drawLinkGrid(screen *ebiten.Image) {
	l := &g.link
	for i := range machine.Reels {
		if l.state.Big && i >= machine.BigFirst && i <= machine.BigLast {
			continue // under the giant chip, drawn below
		}
		for row := range machine.Rows {
			x, y := cellCentre(i, row)
			fillRect(screen, x-reelW/2, y-rowH/2, reelW, rowH, reelBack)
			// Thin lines mark off the cells, which spin one by one.
			frameRect(screen, x-reelW/2, y-rowH/2, reelW, rowH, 1, darkGray)

			switch {
			case l.state.Held[i][row]:
				label := g.chipLabel(l.state.Chips[i][row], l.state.Bet)
				drawChip(screen, x, y, reelScale, label)
				if at := l.landed[i][row]; at > 0 && g.ticks-at < linkFlashTicks && g.ticks/5%2 == 0 {
					boxCell(screen, i, row, yellow)
				}
			case l.phase == linkSpin:
				// A blur of chips passing: each cell shows a faint one now
				// and then, out of step with its neighbours.
				if (g.ticks/3+i*2+row*3)%4 == 0 {
					fillCircle(screen, x, y, 37*reelScale, darkGray)
					fillCircle(screen, x, y, 24*reelScale, reelBack)
				}
			}
		}
	}
	if l.state.Big {
		// The giant chip that started the feature fills the middle reels,
		// with its one prize on it.
		rect := bigRect()
		screen.SubImage(rect).(*ebiten.Image).Fill(reelBack)
		x, y := cellCentre(machine.BigReel, machine.Middle)
		drawChip(screen, x, y, bigScale, g.chipLabel(l.state.Chips[machine.BigReel][machine.Middle], l.state.Bet))
	}
}

// The jackpots are shown across the top, best first.
var jackpots = [...]struct {
	jackpot machine.Jackpot
	fill    color.Color
}{
	{machine.Grand, red},
	{machine.Major, magenta},
	{machine.Minor, blue},
	{machine.Mini, green},
}

// jackpotCredits is what a jackpot would pay at the bet being played.
func (g *Game) jackpotCredits(j machine.Jackpot) int {
	if j == machine.Grand {
		return g.m.Grand()
	}
	return g.m.ChipValue(machine.Chip{Jackpot: j}, g.m.TotalBet())
}

// drawJackpots draws the four jackpot meters across the top of the screen.
func (g *Game) drawJackpots(screen *ebiten.Image) {
	const (
		left  = reelsX - 5
		right = panelX + panelW
		top   = 22
		h     = 38
		gap   = 6
	)
	w := (float64(right-left) - gap*float64(len(jackpots)-1)) / float64(len(jackpots))
	for i, j := range jackpots {
		x := left + float64(i)*(w+gap)
		fillRect(screen, x, top, w, h, j.fill)
		frameRect(screen, x, top, w, h, 2, yellow)
		drawText(screen, j.jackpot.String(), x+8, top+h/2, 13, yellow, alignStart, alignCenter)
		drawText(screen, money(g.jackpotCredits(j.jackpot)), x+w-8, top+h/2, 18, white, alignEnd, alignCenter)
	}
}

// drawStatus draws the line under the jackpots that says what is being
// played: the game's name, or how a feature is going.
func (g *Game) drawStatus(screen *ebiten.Image, inFree bool) {
	text, c := "POKER MACHINE", color.Color(lightBlue)
	switch {
	case g.link.show && g.link.phase == linkDone:
		text, c = fmt.Sprintf("LINK FEATURE PAID %d", g.link.total), yellow
	case g.link.show:
		text = fmt.Sprintf("LINK FEATURE     RESPINS LEFT %d     TOTAL %d", g.link.state.Respins, g.link.total)
		c = yellow
	case inFree && g.free.played == 0:
		text = fmt.Sprintf("%d FREE GAMES     GIANT SYMBOL ON REELS 2 TO 4", g.free.total)
		c = yellow
	case inFree:
		text = fmt.Sprintf("FREE GAME %d OF %d", g.free.played, g.free.total)
		c = yellow
	}
	drawText(screen, text, screenW/2, 74, 15, c, alignCenter, alignCenter)
}

// money writes a number of credits as dollars and cents, at 100 credits to
// the dollar.
func money(credits int) string {
	dollars := fmt.Sprint(credits / 100)
	// Put a comma before each group of three digits.
	for i := len(dollars) - 3; i > 0; i -= 3 {
		dollars = dollars[:i] + "," + dollars[i:]
	}
	return fmt.Sprintf("$%s.%02d", dollars, credits%100)
}

// linkBusy reports whether a link feature is being played, up to the moment
// its last respin has been shown.
func (g *Game) linkBusy() bool {
	return g.m.Link != nil || g.link.show && g.link.phase != linkDone
}
