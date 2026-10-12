// Poker Machine: a five-reel slot machine, rewritten from the Turbo Pascal
// program SLOT.PAS.
package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"slot/machine"
)

// The game draws to a fixed 800x480 screen, the height of the VGA mode the
// original ran in but wider to fit five reels, and Ebitengine scales it to
// the window.
const (
	screenW = 800
	screenH = 480
	border  = 15

	reelsX  = 34
	reelsY  = 90
	reelW   = 100
	reelGap = 6
	rowH    = 90
	reelsW  = machine.Reels*reelW + (machine.Reels-1)*reelGap
	reelsH  = 3 * rowH
	lineY   = reelsY + reelsH/2 // centre of the payline

	reelScale = 0.9 // symbol size that fits a reel cell

	panelX = 582
	panelW = 198
)

type scene int

const (
	sceneMenu scene = iota
	sceneHelp
	sceneAbout
	scenePlay
	sceneBye
)

// reel is the animation state of one reel. pos is the strip index at the
// payline; it is fractional while the reel is moving.
type reel struct {
	pos      float64
	from     float64
	dist     float64
	tick     int
	dur      int
	spinning bool
}

type Game struct {
	scene scene
	ticks int
	snd   *sounds
	rtp   float64 // worked out once; it takes too long to redo every frame

	m        *machine.Machine
	reels    [machine.Reels]reel
	result   machine.Result
	spinning bool
	credits  int // credits on display; a win is added once the reels stop
	win      int
	message  string

	byeTicks int
}

func newGame() *Game {
	g := &Game{snd: newSounds(), rtp: machine.RTP()}
	g.newMachine()
	return g
}

func (g *Game) newMachine() {
	g.m = machine.New(rand.Uint64(), rand.Uint64())
	g.credits = g.m.Credits
	g.win = 0
	g.message = "PRESS SPACE TO SPIN"
}

func pressed(keys ...ebiten.Key) bool {
	for _, k := range keys {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}

func (g *Game) Update() error {
	g.ticks++
	if pressed(ebiten.KeyS) {
		g.snd.on = !g.snd.on
	}

	switch g.scene {
	case sceneMenu:
		switch {
		case pressed(ebiten.KeyDigit1, ebiten.KeyNumpad1):
			g.scene = sceneHelp
		case pressed(ebiten.KeyDigit2, ebiten.KeyNumpad2):
			g.scene = scenePlay
		case pressed(ebiten.KeyDigit3, ebiten.KeyNumpad3):
			g.scene = sceneAbout
		case pressed(ebiten.KeyDigit4, ebiten.KeyNumpad4, ebiten.KeyEscape):
			g.scene = sceneBye
			g.snd.play(g.snd.bye)
		}

	case sceneHelp, sceneAbout:
		if pressed(ebiten.KeyEnter, ebiten.KeyNumpadEnter, ebiten.KeySpace, ebiten.KeyEscape) {
			g.scene = sceneMenu
		}

	case scenePlay:
		g.updatePlay()

	case sceneBye:
		g.byeTicks++
		if g.byeTicks > 100 {
			return ebiten.Termination
		}
	}
	return nil
}

func (g *Game) updatePlay() {
	if g.spinning {
		g.advanceReels()
		return
	}
	switch {
	case pressed(ebiten.KeyEscape):
		g.scene = sceneMenu
	case pressed(ebiten.KeyArrowUp, ebiten.KeyArrowRight):
		g.m.SetBet(g.m.Bet + 1)
	case pressed(ebiten.KeyArrowDown, ebiten.KeyArrowLeft):
		g.m.SetBet(g.m.Bet - 1)
	case pressed(ebiten.KeyM):
		g.m.SetBet(machine.MaxBet)
	case pressed(ebiten.KeySpace, ebiten.KeyEnter, ebiten.KeyNumpadEnter):
		if g.m.Credits == 0 {
			g.newMachine()
		} else {
			g.startSpin()
		}
	}
}

// startSpin decides the outcome straight away, then sets each reel turning
// towards its stop. Reels further right travel further and stop later.
func (g *Game) startSpin() {
	res, ok := g.m.Spin()
	if !ok {
		return
	}
	g.result = res
	g.spinning = true
	g.credits = g.m.Credits - res.Win
	g.win = 0
	g.message = ""
	g.snd.play(g.snd.start)

	for i := range g.reels {
		r := &g.reels[i]
		n := len(machine.Strips[i])
		ahead := ((res.Stops[i]-int(r.pos))%n + n) % n
		r.from = r.pos
		r.dist = float64(n*(i+1) + ahead)
		r.tick = 0
		r.dur = 60 + 25*i
		r.spinning = true
	}
}

func (g *Game) advanceReels() {
	for i := range g.reels {
		r := &g.reels[i]
		if !r.spinning {
			continue
		}
		r.tick++
		if r.tick >= r.dur {
			r.pos = float64(g.result.Stops[i])
			r.spinning = false
			g.snd.play(g.snd.stop)
			continue
		}
		// Ease out: fast at first, slowing into the stop.
		u := float64(r.tick) / float64(r.dur)
		r.pos = r.from + r.dist*(1-(1-u)*(1-u))
	}
	if g.reels[machine.Reels-1].spinning {
		return
	}

	g.spinning = false
	g.credits = g.m.Credits
	g.win = g.result.Win
	switch {
	case g.win > 0:
		g.message = fmt.Sprintf("YOU WIN %d!", g.win)
		if g.result.Matched > 3 || g.result.Line[0] >= machine.Bar {
			g.snd.play(g.snd.bigWin)
		} else {
			g.snd.play(g.snd.win)
		}
	case g.m.Credits == 0:
		g.message = "OUT OF CREDITS - PRESS ENTER FOR A NEW GAME"
	default:
		g.message = "NO WIN"
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenW, screenH
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(black)
	fillRect(screen, 0, 0, screenW, border, green)
	fillRect(screen, 0, screenH-border, screenW, border, green)
	fillRect(screen, 0, 0, border, screenH, green)
	fillRect(screen, screenW-border, 0, border, screenH, green)

	switch g.scene {
	case sceneMenu:
		g.drawMenu(screen)
	case sceneHelp:
		g.drawHelp(screen)
	case sceneAbout:
		g.drawAbout(screen)
	case scenePlay:
		g.drawPlay(screen)
	case sceneBye:
		drawText(screen, "Thanks for playing", screenW/2, screenH/2, 44, lightBlue, alignCenter, alignCenter)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	drawText(screen, "POKER MACHINE MENU", screenW/2, 50, 40, yellow, alignCenter, alignStart)

	items := []string{"HOW TO PLAY GAME", "PLAY GAME", "ABOUT THIS GAME", "EXIT GAME"}
	for i, item := range items {
		y := 140 + float64(i)*36
		drawText(screen, fmt.Sprintf("%d.", i+1), screenW/2-130, y, 24, white, alignStart, alignStart)
		drawText(screen, item, screenW/2-90, y, 24, lightRed, alignStart, alignStart)
	}
	drawText(screen, "MAKE YOUR CHOICE", screenW/2, 300, 24, lightBlue, alignCenter, alignStart)

	for s := machine.Cherry; s < machine.NumSymbols; s++ {
		drawSymbol(screen, s, screenW/2-225+float64(s)*90, 400, 0.7)
	}
}

func (g *Game) drawHelp(screen *ebiten.Image) {
	drawText(screen, "HOW TO PLAY", screenW/2, 40, 36, yellow, alignCenter, alignStart)

	lines := []string{
		"Each spin costs your bet: 1, 2 or 3 credits.",
		"The reels stop at random. Only the middle row counts.",
		"Three or more of a symbol in a row from the left reel win.",
		"Four or five in a row pay much more. Cherries pay from one.",
		"Every win is multiplied by your bet.",
	}
	for i, l := range lines {
		drawText(screen, l, screenW/2-260, 110+float64(i)*26, 17, white, alignStart, alignStart)
	}

	keys := [][2]string{
		{"SPACE or ENTER", "spin the reels"},
		{"UP / DOWN", "change your bet"},
		{"M", "bet the maximum"},
		{"S", "sound on or off"},
		{"ESC", "back to the menu"},
	}
	for i, k := range keys {
		y := 265 + float64(i)*26
		drawText(screen, k[0], screenW/2-70, y, 17, lightRed, alignEnd, alignStart)
		drawText(screen, k[1], screenW/2-45, y, 17, white, alignStart, alignStart)
	}
	drawText(screen, "PRESS ENTER TO RETURN", screenW/2, 420, 20, lightBlue, alignCenter, alignStart)
}

func (g *Game) drawAbout(screen *ebiten.Image) {
	drawText(screen, "ABOUT THIS GAME", screenW/2, 40, 36, yellow, alignCenter, alignStart)

	lines := []string{
		"Poker Machine began as SLOT.PAS, a Turbo Pascal",
		"program drawn with the Graph unit.",
		"",
		"This version is written in Go with Ebitengine.",
		"The reels are random and wins are paid in credits.",
		"",
		fmt.Sprintf("Over a long run it pays back %.0f%% of what is bet.", g.rtp*100),
	}
	for i, l := range lines {
		drawText(screen, l, screenW/2, 130+float64(i)*28, 18, white, alignCenter, alignStart)
	}
	drawText(screen, "PRESS ENTER TO RETURN", screenW/2, 420, 20, lightBlue, alignCenter, alignStart)
}

func (g *Game) drawPlay(screen *ebiten.Image) {
	drawText(screen, "POKER MACHINE", screenW/2, 24, 44, lightBlue, alignCenter, alignStart)

	won := !g.spinning && g.win > 0
	blink := g.ticks/15%2 == 0

	// The magenta frame shows through the gaps between the reels.
	fillRect(screen, reelsX-5, reelsY-5, reelsW+10, reelsH+10, magenta)
	for i := range g.reels {
		x := reelsX + i*(reelW+reelGap)
		g.drawReel(screen, i, x)
		if won && blink && i < g.result.Matched {
			frameRect(screen, float64(x)+2, lineY-rowH/2+2, reelW-4, rowH-4, 4, yellow)
		}
	}
	fillPoly(screen, yellow, reelsX-18, lineY-9, reelsX-7, lineY, reelsX-18, lineY+9)
	fillPoly(screen, yellow, reelsX+reelsW+18, lineY-9, reelsX+reelsW+7, lineY, reelsX+reelsW+18, lineY+9)

	g.drawPaytable(screen, won && blink)

	drawStat(screen, "CREDITS", g.credits, reelsX+reelW/2, lightGreen)
	drawStat(screen, "BET", g.m.Bet, reelsX+reelsW/2, white)
	drawStat(screen, "WIN", g.win, reelsX+reelsW-reelW/2, yellow)

	drawText(screen, g.message, screenW/2, 424, 18, lightRed, alignCenter, alignStart)

	sound := "ON"
	if !g.snd.on {
		sound = "OFF"
	}
	hint := fmt.Sprintf("SPACE spin    UP/DOWN bet    M max bet    S sound %s    ESC menu", sound)
	drawText(screen, hint, screenW/2, 448, 11, lightGray, alignCenter, alignStart)
}

// drawReel draws the part of a reel's strip that shows in its window. The
// sub-image clips symbols that are part-way in or out.
func (g *Game) drawReel(screen *ebiten.Image, i, x int) {
	window := screen.SubImage(image.Rect(x, reelsY, x+reelW, reelsY+reelsH)).(*ebiten.Image)
	window.Fill(reelBack)

	strip := machine.Strips[i]
	n := len(strip)
	base := math.Floor(g.reels[i].pos)
	frac := g.reels[i].pos - base
	for k := -2; k <= 2; k++ {
		sym := strip[((int(base)+k)%n+n)%n]
		cy := lineY - (float64(k)-frac)*rowH
		drawSymbol(window, sym, float64(x)+reelW/2, cy, reelScale)
	}
}

// The paytable lists symbols from the best paying down.
var payOrder = [...]machine.Symbol{
	machine.Dollar, machine.Seven, machine.Bar, machine.Bell, machine.Plum, machine.Cherry,
}

func (g *Game) drawPaytable(screen *ebiten.Image, highlight bool) {
	const rowsY, rowStep = 118, 29

	// The right-hand edge of the column for a run of n symbols.
	col := func(n int) float64 { return panelX + panelW - 12 - float64(machine.Reels-n)*46 }

	frameRect(screen, panelX, reelsY-4, panelW, reelsH+8, 2, magenta)
	drawText(screen, "PAYS", panelX+12, reelsY+4, 16, yellow, alignStart, alignStart)
	for n := 3; n <= machine.Reels; n++ {
		drawText(screen, fmt.Sprintf("x%d", n), col(n), reelsY+4, 16, yellow, alignEnd, alignStart)
	}

	// row draws one line of the table and reports whether it is the win.
	row := func(i int, sym machine.Symbol, icons int, wins bool) (cy float64, lit bool) {
		cy = float64(rowsY + i*rowStep + rowStep/2)
		lit = highlight && wins
		if lit {
			fillRect(screen, panelX+3, cy-rowStep/2, panelW-6, rowStep, darkGray)
		}
		for j := range icons {
			drawSymbol(screen, sym, panelX+24+float64(j)*32, cy, 0.28)
		}
		return cy, lit
	}

	won, matched := g.result.Line[0], g.result.Matched
	i := 0
	for _, sym := range payOrder {
		cy, lit := row(i, sym, 1, sym == won && matched >= 3)
		for n := 3; n <= machine.Reels; n++ {
			c := color.Color(white)
			if lit && n == matched {
				c = yellow
			}
			drawText(screen, fmt.Sprint(machine.Pays[sym][n]), col(n), cy, 15, c, alignEnd, alignCenter)
		}
		i++
	}
	// Cherries are the only symbol to pay for a run of two or one.
	for n := 2; n >= 1; n-- {
		cy, _ := row(i, machine.Cherry, n, won == machine.Cherry && matched == n)
		drawText(screen, fmt.Sprint(machine.Pays[machine.Cherry][n]), col(machine.Reels), cy, 15, white, alignEnd, alignCenter)
		i++
	}
	drawText(screen, "x BET", panelX+panelW-12, reelsY+reelsH-3, 10, lightGray, alignEnd, alignCenter)
}

// drawStat draws a labelled number under the reels, centred on x.
func drawStat(screen *ebiten.Image, label string, value int, x float64, c color.Color) {
	drawText(screen, label, x, 372, 12, lightGray, alignCenter, alignStart)
	drawText(screen, fmt.Sprint(value), x, 386, 26, c, alignCenter, alignStart)
}

func main() {
	ebiten.SetWindowSize(screenW*3/2, screenH*3/2)
	ebiten.SetWindowTitle("Poker Machine")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
