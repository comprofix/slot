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
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"slot/machine"
)

// The game draws to a fixed 800x600 screen and Ebitengine scales it to the
// window.
const (
	screenW = 800
	screenH = 600
	border  = 15

	reelsX  = 34
	reelsY  = 90
	reelW   = 100
	reelGap = 6
	rowH    = 90
	reelsW  = machine.Reels*reelW + (machine.Reels-1)*reelGap
	reelsH  = machine.Rows * rowH

	reelScale = 0.9 // symbol size that fits a reel cell

	panelX = 582
	panelW = 198

	// The two rows of buttons under the reels.
	betRowY  = 448
	lineRowY = 494
	buttonH  = 40
)

type scene int

const (
	sceneMenu scene = iota
	sceneHelp
	sceneAbout
	scenePlay
	sceneBye
)

// reel is the animation state of one reel. pos is the strip index on the
// middle row; it is fractional while the reel is moving.
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

	stopTick  int // when the reels last stopped, to time the win display
	showLines int // ticks left to show the lines being played

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
	g.clearWin()
	g.message = "CHOOSE YOUR BET AND PRESS SPIN"
}

// clearWin takes the last spin's win off the display.
func (g *Game) clearWin() {
	g.result.Wins = nil
	g.win = 0
	g.message = ""
}

func pressed(keys ...ebiten.Key) bool {
	for _, k := range keys {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}

// button is a clickable rectangle.
type button struct{ x, y, w, h float64 }

func (b button) hit(x, y int) bool {
	fx, fy := float64(x), float64(y)
	return fx >= b.x && fx < b.x+b.w && fy >= b.y && fy < b.y+b.h
}

// The bet and line buttons sit under the reels, one per reel column.
func betButton(i int) button {
	return button{reelsX + float64(i)*(reelW+reelGap), betRowY, reelW, buttonH}
}

func lineButton(i int) button {
	return button{reelsX + float64(i)*(reelW+reelGap), lineRowY, reelW, buttonH}
}

var spinButton = button{panelX, betRowY, panelW, lineRowY + buttonH - betRowY}

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
	if g.showLines > 0 {
		g.showLines--
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		for i, n := range machine.BetChoices {
			if betButton(i).hit(x, y) {
				g.setBet(n)
			}
		}
		for i, n := range machine.LineChoices {
			if lineButton(i).hit(x, y) {
				g.setLines(n)
			}
		}
		if spinButton.hit(x, y) {
			g.spin()
		}
	}

	switch {
	case pressed(ebiten.KeyEscape):
		g.scene = sceneMenu
	case pressed(ebiten.KeyArrowUp):
		g.setBet(step(machine.BetChoices, g.m.BetPerLine, 1))
	case pressed(ebiten.KeyArrowDown):
		g.setBet(step(machine.BetChoices, g.m.BetPerLine, -1))
	case pressed(ebiten.KeyArrowRight):
		g.setLines(step(machine.LineChoices, g.m.Lines, 1))
	case pressed(ebiten.KeyArrowLeft):
		g.setLines(step(machine.LineChoices, g.m.Lines, -1))
	case pressed(ebiten.KeyM):
		g.maxBet()
	case pressed(ebiten.KeySpace, ebiten.KeyEnter, ebiten.KeyNumpadEnter):
		g.spin()
	}
}

// step returns the choice one place up or down from current, stopping at
// either end of the list.
func step(choices []int, current, by int) int {
	i := slices.Index(choices, current) + by
	return choices[min(max(i, 0), len(choices)-1)]
}

func (g *Game) setBet(n int) {
	g.m.BetPerLine = n
	g.clearWin()
	g.snd.play(g.snd.click)
}

func (g *Game) setLines(n int) {
	g.m.Lines = n
	g.clearWin()
	g.showLines = 120
	g.snd.play(g.snd.click)
}

// maxBet plays every line at the highest bet per line the credits cover.
func (g *Game) maxBet() {
	lines := len(machine.Paylines)
	bet := machine.BetChoices[0]
	for _, n := range machine.BetChoices {
		if lines*n <= g.m.Credits {
			bet = n
		}
	}
	g.setLines(lines)
	g.m.BetPerLine = bet
}

// spin decides the outcome straight away, then sets each reel turning
// towards its stop. Reels further right travel further and stop later.
func (g *Game) spin() {
	if g.m.Credits == 0 {
		g.newMachine()
		return
	}
	res, ok := g.m.Spin()
	if !ok {
		g.clearWin()
		g.message = "NOT ENOUGH CREDITS - CHOOSE A SMALLER BET"
		return
	}
	g.clearWin()
	g.result = res
	g.spinning = true
	g.showLines = 0
	g.credits = g.m.Credits - res.Win
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
	g.stopTick = g.ticks
	g.credits = g.m.Credits
	g.win = g.result.Win
	switch {
	case g.win > 0:
		g.message = fmt.Sprintf("YOU WIN %d!", g.win)
		if g.win >= 10*g.m.TotalBet() {
			g.snd.play(g.snd.bigWin)
		} else {
			g.snd.play(g.snd.win)
		}
	case g.m.Credits == 0:
		g.message = "OUT OF CREDITS - PRESS SPIN FOR A NEW GAME"
	default:
		g.message = "NO WIN"
	}
}

// shownWin is the winning line on display. The wins take turns, about a
// second each.
func (g *Game) shownWin() *machine.LineWin {
	if g.spinning || len(g.result.Wins) == 0 {
		return nil
	}
	i := (g.ticks - g.stopTick) / 60 % len(g.result.Wins)
	return &g.result.Wins[i]
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
	drawText(screen, "POKER MACHINE MENU", screenW/2, 90, 40, yellow, alignCenter, alignStart)

	items := []string{"HOW TO PLAY GAME", "PLAY GAME", "ABOUT THIS GAME", "EXIT GAME"}
	for i, item := range items {
		y := 190 + float64(i)*36
		drawText(screen, fmt.Sprintf("%d.", i+1), screenW/2-130, y, 24, white, alignStart, alignStart)
		drawText(screen, item, screenW/2-90, y, 24, lightRed, alignStart, alignStart)
	}
	drawText(screen, "MAKE YOUR CHOICE", screenW/2, 355, 24, lightBlue, alignCenter, alignStart)

	for s := machine.Cherry; s < machine.NumSymbols; s++ {
		drawSymbol(screen, s, screenW/2-225+float64(s)*90, 470, 0.7)
	}
}

func (g *Game) drawHelp(screen *ebiten.Image) {
	drawText(screen, "HOW TO PLAY", screenW/2, 40, 36, yellow, alignCenter, alignStart)

	lines := []string{
		"Choose a bet per line and how many lines to play, then spin.",
		"A spin costs the bet per line times the number of lines.",
		"Each line is a path across the five reels, read left to right.",
		"Three or more of a symbol in a row from the left reel win.",
		"Four or five in a row pay much more. Cherries pay from one.",
		"A line's win is multiplied by the bet per line.",
	}
	for i, l := range lines {
		drawText(screen, l, screenW/2-265, 110+float64(i)*28, 17, white, alignStart, alignStart)
	}

	keys := [][2]string{
		{"SPACE or ENTER", "spin the reels"},
		{"UP / DOWN", "change the bet per line"},
		{"LEFT / RIGHT", "change the number of lines"},
		{"M", "bet the maximum"},
		{"S", "sound on or off"},
		{"ESC", "back to the menu"},
	}
	for i, k := range keys {
		y := 310 + float64(i)*28
		drawText(screen, k[0], screenW/2-70, y, 17, lightRed, alignEnd, alignStart)
		drawText(screen, k[1], screenW/2-45, y, 17, white, alignStart, alignStart)
	}
	drawText(screen, "The buttons on screen can be clicked with the mouse.", screenW/2, 490, 15, lightGray, alignCenter, alignStart)
	drawText(screen, "PRESS ENTER TO RETURN", screenW/2, 530, 20, lightBlue, alignCenter, alignStart)
}

func (g *Game) drawAbout(screen *ebiten.Image) {
	drawText(screen, "ABOUT THIS GAME", screenW/2, 60, 36, yellow, alignCenter, alignStart)

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
		drawText(screen, l, screenW/2, 170+float64(i)*28, 18, white, alignCenter, alignStart)
	}
	drawText(screen, "PRESS ENTER TO RETURN", screenW/2, 480, 20, lightBlue, alignCenter, alignStart)
}

func (g *Game) drawPlay(screen *ebiten.Image) {
	drawText(screen, "POKER MACHINE", screenW/2, 24, 44, lightBlue, alignCenter, alignStart)

	// The magenta frame shows through the gaps between the reels.
	fillRect(screen, reelsX-5, reelsY-5, reelsW+10, reelsH+10, magenta)
	for i := range g.reels {
		g.drawReel(screen, i)
	}

	win := g.shownWin()
	switch {
	case win != nil:
		drawPayline(screen, win.Line, 4)
		for i := range win.Matched {
			x, y := cellCentre(i, machine.Paylines[win.Line][i])
			frameRect(screen, x-reelW/2+2, y-rowH/2+2, reelW-4, rowH-4, 4, lineColor(win.Line))
		}
	case g.showLines > 0 && !g.spinning:
		for l := range g.m.Lines {
			drawPayline(screen, l, 2)
		}
	}

	g.drawPaytable(screen, win)

	drawStat(screen, "CREDITS", g.credits, 0, lightGreen)
	drawStat(screen, "BET / LINE", g.m.BetPerLine, 1, white)
	drawStat(screen, "LINES", g.m.Lines, 2, white)
	drawStat(screen, "TOTAL BET", g.m.TotalBet(), 3, white)
	drawStat(screen, "WIN", g.win, 4, yellow)

	message := g.message
	if win != nil && len(g.result.Wins) > 1 {
		message += fmt.Sprintf("   LINE %d PAYS %d", win.Line+1, win.Win)
	}
	drawText(screen, message, reelsX+reelsW/2, 420, 16, lightRed, alignCenter, alignStart)

	for i, n := range machine.BetChoices {
		drawButton(screen, betButton(i), n == g.m.BetPerLine, 11, fmt.Sprintf("PLAY %d", n), "PER LINE")
	}
	for i, n := range machine.LineChoices {
		label := fmt.Sprintf("%d LINES", n)
		if n == 1 {
			label = "1 LINE"
		}
		drawButton(screen, lineButton(i), n == g.m.Lines, 14, label)
	}
	spinLabel := "SPIN"
	if g.m.Credits == 0 && !g.spinning {
		spinLabel = "NEW GAME"
	}
	fillRect(screen, spinButton.x, spinButton.y, spinButton.w, spinButton.h, green)
	frameRect(screen, spinButton.x, spinButton.y, spinButton.w, spinButton.h, 2, lightGreen)
	drawText(screen, spinLabel, spinButton.x+spinButton.w/2, spinButton.y+spinButton.h/2, 30, white, alignCenter, alignCenter)

	sound := "ON"
	if !g.snd.on {
		sound = "OFF"
	}
	hint := fmt.Sprintf("SPACE spin    UP/DOWN bet per line    LEFT/RIGHT lines    M max bet    S sound %s    ESC menu", sound)
	drawText(screen, hint, screenW/2, 550, 11, lightGray, alignCenter, alignStart)
}

// cellCentre is the middle of the cell at the given reel and row.
func cellCentre(reel, row int) (x, y float64) {
	return reelsX + float64(reel)*(reelW+reelGap) + reelW/2,
		reelsY + float64(row)*rowH + rowH/2
}

// lineColor gives each payline its own colour. Stepping the hue by a large
// odd amount keeps neighbouring lines well apart.
func lineColor(line int) color.Color {
	h := float64(line*137%360) / 60
	x := uint8(255 * (1 - math.Abs(math.Mod(h, 2)-1)))
	switch int(h) {
	case 0:
		return color.RGBA{255, x, 0, 255}
	case 1:
		return color.RGBA{x, 255, 0, 255}
	case 2:
		return color.RGBA{0, 255, x, 255}
	case 3:
		return color.RGBA{0, x, 255, 255}
	case 4:
		return color.RGBA{x, 0, 255, 255}
	}
	return color.RGBA{255, 0, x, 255}
}

// drawPayline draws a payline's path across the reels, with its number at
// the left-hand end.
func drawPayline(screen *ebiten.Image, line int, width float64) {
	c := lineColor(line)
	rows := machine.Paylines[line]

	px, py := cellCentre(0, rows[0])
	px = reelsX - 5
	seg := func(x, y float64) {
		drawLine(screen, px, py, x, y, width+2, black)
		drawLine(screen, px, py, x, y, width, c)
		px, py = x, y
	}
	for i, row := range rows {
		seg(cellCentre(i, row))
	}
	seg(reelsX+reelsW+5, py)

	_, y := cellCentre(0, rows[0])
	drawText(screen, fmt.Sprint(line+1), reelsX-7, y, 10, c, alignEnd, alignCenter)
}

// drawReel draws the part of a reel's strip that shows in its window. The
// sub-image clips symbols that are part-way in or out.
func (g *Game) drawReel(screen *ebiten.Image, i int) {
	cx, cy := cellCentre(i, machine.Middle)
	x := int(cx) - reelW/2
	window := screen.SubImage(image.Rect(x, reelsY, x+reelW, reelsY+reelsH)).(*ebiten.Image)
	window.Fill(reelBack)

	strip := machine.Strips[i]
	n := len(strip)
	base := math.Floor(g.reels[i].pos)
	frac := g.reels[i].pos - base
	for k := -2; k <= 2; k++ {
		sym := strip[((int(base)+k)%n+n)%n]
		drawSymbol(window, sym, cx, cy-(float64(k)-frac)*rowH, reelScale)
	}
}

// The paytable lists symbols from the best paying down.
var payOrder = [...]machine.Symbol{
	machine.Dollar, machine.Seven, machine.Bar, machine.Bell, machine.Plum, machine.Cherry,
}

// drawPaytable draws what each win pays at the current bet per line,
// picking out the entry for win if there is one.
func (g *Game) drawPaytable(screen *ebiten.Image, win *machine.LineWin) {
	const rowsY, rowStep = 118, 29

	// The right-hand edge of the column for a run of n symbols. Columns for
	// longer runs are wider, as their payouts have more digits.
	edges := [machine.Reels + 1]float64{3: 104, 4: 58, 5: 0}
	col := func(n int) float64 { return panelX + panelW - 12 - edges[n] }
	pays := func(sym machine.Symbol, n int) string {
		return fmt.Sprint(machine.Pays[sym][n] * g.m.BetPerLine)
	}

	frameRect(screen, panelX, reelsY-4, panelW, reelsH+8, 2, magenta)
	drawText(screen, "PAYS", panelX+12, reelsY+4, 16, yellow, alignStart, alignStart)
	for n := 3; n <= machine.Reels; n++ {
		drawText(screen, fmt.Sprintf("x%d", n), col(n), reelsY+4, 16, yellow, alignEnd, alignStart)
	}

	// row draws one line of the table, lit if it is the win, and returns
	// the height of its centre.
	row := func(i int, sym machine.Symbol, icons int, lit bool) float64 {
		cy := float64(rowsY + i*rowStep + rowStep/2)
		if lit {
			fillRect(screen, panelX+3, cy-rowStep/2, panelW-6, rowStep, darkGray)
		}
		for j := range icons {
			drawSymbol(screen, sym, panelX+24+float64(j)*32, cy, 0.28)
		}
		return cy
	}

	i := 0
	for _, sym := range payOrder {
		lit := win != nil && win.Symbol == sym && win.Matched >= 3
		cy := row(i, sym, 1, lit)
		for n := 3; n <= machine.Reels; n++ {
			c := color.Color(white)
			if lit && n == win.Matched {
				c = yellow
			}
			drawText(screen, pays(sym, n), col(n), cy, 13, c, alignEnd, alignCenter)
		}
		i++
	}
	// Cherries are the only symbol to pay for a run of two or one.
	for n := 2; n >= 1; n-- {
		lit := win != nil && win.Symbol == machine.Cherry && win.Matched == n
		cy := row(i, machine.Cherry, n, lit)
		drawText(screen, pays(machine.Cherry, n), col(machine.Reels), cy, 13, white, alignEnd, alignCenter)
		i++
	}
	note := fmt.Sprintf("AT %d PER LINE", g.m.BetPerLine)
	drawText(screen, note, panelX+panelW-12, reelsY+reelsH-3, 10, lightGray, alignEnd, alignCenter)
}

// drawStat draws a labelled number under the given reel column.
func drawStat(screen *ebiten.Image, label string, value, column int, c color.Color) {
	x, _ := cellCentre(column, 0)
	drawText(screen, label, x, 370, 11, lightGray, alignCenter, alignStart)
	drawText(screen, fmt.Sprint(value), x, 384, 24, c, alignCenter, alignStart)
}

// drawButton draws a button with one or two lines of text. The selected
// button of a row is shown lit.
func drawButton(screen *ebiten.Image, b button, selected bool, size float64, lines ...string) {
	fill, edge, ink := color.Color(blue), color.Color(lightBlue), color.Color(white)
	if selected {
		fill, edge, ink = yellow, white, black
	}
	fillRect(screen, b.x, b.y, b.w, b.h, fill)
	frameRect(screen, b.x, b.y, b.w, b.h, 2, edge)

	step := size + 3
	y := b.y + b.h/2 - step*float64(len(lines)-1)/2
	for _, l := range lines {
		drawText(screen, l, b.x+b.w/2, y, size, ink, alignCenter, alignCenter)
		y += step
	}
}

func main() {
	// Use a bigger window where the monitor has room for it.
	scale := 1.0
	if _, h := ebiten.Monitor().Size(); h >= 1000 {
		scale = 1.5
	}
	ebiten.SetWindowSize(int(screenW*scale), int(screenH*scale))
	ebiten.SetWindowTitle("Poker Machine")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
