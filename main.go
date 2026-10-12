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
	panelH = 354 // the paytable runs down beside the reels and the figures

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

	free freeSpins
	link linkView
	spun bool // a spin has been played, so the chips showing were dealt

	byeTicks int
}

// freeSpins tracks a run of free spins, from the spin that wins them until
// the last one has been played.
type freeSpins struct {
	active bool
	total  int // spins awarded so far, counting any won along the way
	played int
	win    int // credits won by the free spins
}

func newGame() *Game {
	g := &Game{snd: newSounds(), rtp: machine.RTP(len(machine.Paylines))}
	g.newMachine()
	return g
}

func (g *Game) newMachine() {
	g.m = machine.New(rand.Uint64(), rand.Uint64())
	g.credits = g.m.Credits
	g.free = freeSpins{}
	g.link = linkView{}
	g.spun = false
	g.clearWin()
	g.message = "CHOOSE YOUR BET AND PRESS SPIN"
}

// clearWin takes the last spin's win off the display.
func (g *Game) clearWin() {
	g.result.Wins = nil
	g.result.Scatters = 0
	g.link = linkView{}
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

// The spin button sits under the paytable beside the two rows, with the coin
// and reset buttons below it.
const (
	coinH  = 24
	resetW = 64
)

var (
	spinButton  = button{panelX, betRowY, panelW, lineRowY + buttonH - betRowY - coinH - 4}
	coinButton  = button{panelX, lineRowY + buttonH - coinH, panelW - resetW - 4, coinH}
	resetButton = button{panelX + panelW - resetW, lineRowY + buttonH - coinH, resetW, coinH}
)

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
	click := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	x, y := ebiten.CursorPosition()
	if pressed(ebiten.KeyC) || click && coinButton.hit(x, y) {
		g.insertCoin()
	}

	// A link feature runs to its end before anything else can happen.
	if g.linkBusy() {
		g.updateLink(click && spinButton.hit(x, y))
		return
	}

	// Free spins play themselves at the bet that won them, after a pause to
	// show the last result. Pressing spin skips the pause.
	if g.m.FreeSpins > 0 {
		pause := 50
		switch {
		case g.link.show, g.result.FreeSpinsWon > 0:
			pause = 180
		case g.win > 0:
			pause = 110
		}
		switch {
		case pressed(ebiten.KeyEscape):
			g.scene = sceneMenu
		case g.ticks-g.stopTick > pause, click && spinButton.hit(x, y),
			pressed(ebiten.KeySpace, ebiten.KeyEnter, ebiten.KeyNumpadEnter):
			g.spin()
		}
		return
	}

	if click {
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
		if resetButton.hit(x, y) {
			g.reset()
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
	case pressed(ebiten.KeyR):
		g.reset()
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

const outOfCredits = "OUT OF CREDITS - INSERT A COIN OR START A NEW GAME"

// reset starts a new game with the starting credits.
func (g *Game) reset() {
	g.newMachine()
	g.snd.play(g.snd.click)
}

// insertCoin adds credits, so play can go on after they run out.
func (g *Game) insertCoin() {
	g.m.InsertCoin()
	g.credits = g.m.Credits
	g.message = fmt.Sprintf("COIN INSERTED - %d CREDITS ADDED", machine.CoinCredits)
	g.snd.play(g.snd.coin)
}

// spin decides the outcome straight away, then sets each reel turning
// towards its stop. Reels further right travel further and stop later.
func (g *Game) spin() {
	if g.m.Credits == 0 && g.m.FreeSpins == 0 {
		g.reset()
		return
	}
	res, ok := g.m.Spin()
	if !ok {
		g.clearWin()
		g.message = "NOT ENOUGH CREDITS - INSERT A COIN OR BET LESS"
		if g.m.Credits == 0 {
			g.message = outOfCredits
		}
		return
	}
	if res.Free {
		g.free.played++
	} else {
		g.free = freeSpins{}
	}
	g.clearWin()
	g.result = res
	g.spun = true
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
	if g.result.Free {
		g.free.win += g.win
	}
	if won := g.result.FreeSpinsWon; won > 0 {
		g.free.active = true
		g.free.total += won
	}

	switch {
	case g.result.FreeSpinsWon > 0:
		g.message = fmt.Sprintf("%d DICE! %d FREE SPINS", g.result.Scatters, g.result.FreeSpinsWon)
		if g.win > 0 {
			g.message += fmt.Sprintf(" AND YOU WIN %d!", g.win)
		}
	case g.win > 0:
		g.message = fmt.Sprintf("YOU WIN %d!", g.win)
	case g.m.Credits == 0 && g.m.FreeSpins == 0:
		g.message = outOfCredits
	default:
		g.message = "NO WIN"
	}
	// The last free spin: report what the whole run won.
	if g.free.active && g.m.FreeSpins == 0 {
		g.free.active = false
		g.message = fmt.Sprintf("FREE SPINS OVER - THEY WON %d CREDITS", g.free.win)
	}

	switch {
	case g.result.FreeSpinsWon > 0, g.win >= 10*g.m.TotalBet():
		g.snd.play(g.snd.bigWin)
	case g.win > 0:
		g.snd.play(g.snd.win)
	}
	if g.result.LinkStarted {
		g.startLink()
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

	for s := machine.Ten; s < machine.NumSymbols; s++ {
		drawSymbol(screen, s, screenW/2-338+float64(s)*52, 470, 0.42)
	}
}

func (g *Game) drawHelp(screen *ebiten.Image) {
	drawText(screen, "HOW TO PLAY", screenW/2, 34, 36, yellow, alignCenter, alignStart)

	lines := []string{
		"Choose a bet per line and how many lines to play, then spin.",
		"A spin costs the bet per line times the number of lines.",
		"Each line is a path across the five reels, read left to right.",
		"Three or more of a symbol in a row from the left reel win.",
		"Four or five in a row pay much more. A few symbols pay for two.",
		"A win on a line is multiplied by the bet per line.",
		"The joker is wild on reels 2 to 5 and doubles any win it helps make.",
		"Three or more dice anywhere win free spins, with all wins doubled.",
		"Six or more chips lock in place and respin for the jackpots.",
	}
	for i, l := range lines {
		drawText(screen, l, screenW/2-280, 86+float64(i)*24, 16, white, alignStart, alignStart)
	}

	keys := [][2]string{
		{"SPACE or ENTER", "spin the reels"},
		{"UP / DOWN", "change the bet per line"},
		{"LEFT / RIGHT", "change the number of lines"},
		{"M", "bet the maximum"},
		{"C", "insert a coin for more credits"},
		{"R", "reset to a new game"},
		{"S", "sound on or off"},
		{"ESC", "back to the menu"},
	}
	for i, k := range keys {
		y := 304 + float64(i)*23
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
	g.drawJackpots(screen)
	inFree := g.free.active || g.m.FreeSpins > 0
	g.drawStatus(screen, inFree)

	// The magenta frame shows through the gaps between the reels.
	fillRect(screen, reelsX-5, reelsY-5, reelsW+10, reelsH+10, magenta)
	var win *machine.LineWin
	if g.link.show {
		g.drawLinkGrid(screen)
	} else {
		for i := range g.reels {
			g.drawReel(screen, i)
		}
		win = g.shownWin()
		g.drawWins(screen, win)
	}

	g.drawPaytable(screen, win)

	drawStat(screen, "CREDITS", g.credits, 0, lightGreen)
	drawStat(screen, "BET / LINE", g.m.BetPerLine, 1, white)
	drawStat(screen, "LINES", g.m.Lines, 2, white)
	if inFree {
		drawStat(screen, "FREE SPINS WIN", g.free.win, 3, yellow)
	} else {
		drawStat(screen, "TOTAL BET", g.m.TotalBet(), 3, white)
	}
	drawStat(screen, "WIN", g.win, 4, yellow)

	message := g.message
	if win != nil && len(g.result.Wins) > 1 {
		message += fmt.Sprintf("   LINE %d PAYS %d", win.Line+1, win.Win)
	}
	if win != nil && win.Doubled {
		message += fmt.Sprintf("   JOKER x%d", machine.WildMultiplier)
	}
	drawText(screen, message, reelsX+reelsW/2, 420, 16, lightRed, alignCenter, alignStart)

	// The bet cannot be changed while a feature is being played.
	locked := g.m.FreeSpins > 0 || g.linkBusy()
	for i, n := range machine.BetChoices {
		drawButton(screen, betButton(i), n == g.m.BetPerLine, locked, 11, fmt.Sprintf("PLAY %d", n), "PER LINE")
	}
	for i, n := range machine.LineChoices {
		label := fmt.Sprintf("%d LINES", n)
		if n == 1 {
			label = "1 LINE"
		}
		drawButton(screen, lineButton(i), n == g.m.Lines, locked, 14, label)
	}
	spinLabel := "SPIN"
	switch {
	case g.linkBusy():
		spinLabel = "RESPIN"
	case locked:
		spinLabel = "FREE SPIN"
	case g.m.Credits == 0 && !g.spinning:
		spinLabel = "NEW GAME"
	}
	fillRect(screen, spinButton.x, spinButton.y, spinButton.w, spinButton.h, green)
	frameRect(screen, spinButton.x, spinButton.y, spinButton.w, spinButton.h, 2, lightGreen)
	drawText(screen, spinLabel, spinButton.x+spinButton.w/2, spinButton.y+spinButton.h/2, 30, white, alignCenter, alignCenter)

	fillRect(screen, coinButton.x, coinButton.y, coinButton.w, coinButton.h, gold)
	frameRect(screen, coinButton.x, coinButton.y, coinButton.w, coinButton.h, 2, yellow)
	coinLabel := fmt.Sprintf("INSERT COIN  +%d", machine.CoinCredits)
	drawText(screen, coinLabel, coinButton.x+coinButton.w/2, coinButton.y+coinButton.h/2, 12, black, alignCenter, alignCenter)

	fillRect(screen, resetButton.x, resetButton.y, resetButton.w, resetButton.h, red)
	frameRect(screen, resetButton.x, resetButton.y, resetButton.w, resetButton.h, 2, lightRed)
	drawText(screen, "RESET", resetButton.x+resetButton.w/2, resetButton.y+resetButton.h/2, 12, white, alignCenter, alignCenter)

	sound := "ON"
	if !g.snd.on {
		sound = "OFF"
	}
	hint := fmt.Sprintf("SPACE spin    UP/DOWN bet per line    LEFT/RIGHT lines    M max bet    C coin    R reset    S sound %s    ESC menu", sound)
	drawText(screen, hint, screenW/2, 550, 11, lightGray, alignCenter, alignStart)
}

// drawWins marks what the last spin won over the reels: the payline on
// show, or the lines being played after a change of bet.
func (g *Game) drawWins(screen *ebiten.Image, win *machine.LineWin) {
	switch {
	case win != nil:
		drawPayline(screen, win.Line, 4)
		for i := range win.Matched {
			boxCell(screen, i, machine.Paylines[win.Line][i], lineColor(win.Line))
		}
	case g.showLines > 0 && !g.spinning:
		for l := range g.m.Lines {
			drawPayline(screen, l, 2)
		}
	}
	// Winning scatters flash wherever they landed.
	if !g.spinning && g.result.FreeSpinsWon > 0 && g.ticks/15%2 == 0 {
		for i, col := range g.result.Grid {
			for row, sym := range col {
				if sym == machine.Scatter {
					boxCell(screen, i, row, white)
				}
			}
		}
	}
}

// boxCell outlines the cell at the given reel and row.
func boxCell(screen *ebiten.Image, reel, row int, c color.Color) {
	x, y := cellCentre(reel, row)
	frameRect(screen, x-reelW/2+2, y-rowH/2+2, reelW-4, rowH-4, 4, c)
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
		stop := ((int(base)+k)%n + n) % n
		y := cy - (float64(k)-frac)*rowH
		if strip[stop] == machine.Link {
			drawChip(window, cx, y, reelScale, g.chipLabel(g.chipAt(i, stop), g.m.TotalBet()))
			continue
		}
		drawSymbol(window, strip[stop], cx, y, reelScale)
	}
}

// chipAt is what the chip at a reel stop shows. The chips the last spin
// landed carry what they were dealt; any others, seen only as the reel
// turns, are dressed up from a fixed list.
func (g *Game) chipAt(reel, stop int) machine.Chip {
	if g.spun {
		n := len(machine.Strips[reel])
		for row := range machine.Rows {
			// Rows run top to bottom, against the order of the strip.
			if stop == ((g.result.Stops[reel]+1-row)%n+n)%n {
				return g.result.Chips[reel][row]
			}
		}
	}
	return passingChips[(reel*5+stop)%len(passingChips)]
}

var passingChips = []machine.Chip{
	{Bets: 1}, {Bets: 2}, {Bets: 1}, {Jackpot: machine.Mini}, {Bets: 3},
	{Bets: 1}, {Bets: 5}, {Bets: 2}, {Jackpot: machine.Minor}, {Bets: 1},
	{Bets: 10}, {Bets: 2}, {Jackpot: machine.Major},
}

// chipLabel is the text on a link chip: the name of its jackpot, or its
// value in credits at the given total bet.
func (g *Game) chipLabel(c machine.Chip, bet int) string {
	if c.Jackpot != machine.NoJackpot {
		return c.Jackpot.String()
	}
	return fmt.Sprint(g.m.ChipValue(c, bet))
}

// drawPaytable draws what each win pays at the current bet, picking out the
// entry for win if there is one.
func (g *Game) drawPaytable(screen *ebiten.Image, win *machine.LineWin) {
	const (
		top     = reelsY - 4
		rowsY   = top + 42
		rowStep = 22
		icon    = 0.22
	)

	// How far the right-hand edge of the column for a run of n symbols is
	// from the edge of the panel. Columns for longer runs are wider, as
	// their payouts have more digits.
	edges := [machine.Reels + 1]float64{2: 126, 3: 91, 4: 49, 5: 0}
	col := func(n int) float64 { return panelX + panelW - 12 - edges[n] }

	frameRect(screen, panelX, top, panelW, panelH, 2, magenta)
	drawText(screen, "PAYS", panelX+panelW/2, top+5, 15, yellow, alignCenter, alignStart)
	for n := 2; n <= machine.Reels; n++ {
		drawText(screen, fmt.Sprintf("x%d", n), col(n), top+25, 12, yellow, alignEnd, alignStart)
	}

	// The symbols are listed from the best paying down.
	i := 0
	for sym := machine.Dollar; sym >= machine.Ten; sym-- {
		cy := float64(rowsY + i*rowStep + rowStep/2)
		lit := win != nil && win.Symbol == sym
		if lit {
			fillRect(screen, panelX+3, cy-rowStep/2, panelW-6, rowStep, darkGray)
		}
		drawSymbol(screen, sym, panelX+22, cy, icon)
		for n := 2; n <= machine.Reels; n++ {
			pay := machine.Pays[sym][n] * g.m.BetPerLine
			if pay == 0 {
				continue
			}
			c := color.Color(white)
			if lit && n == win.Matched {
				c = yellow
			}
			drawText(screen, fmt.Sprint(pay), col(n), cy, 12, c, alignEnd, alignCenter)
		}
		i++
	}

	y := float64(rowsY + i*rowStep + 4)
	note := fmt.Sprintf("AT %d PER LINE", g.m.BetPerLine)
	drawText(screen, note, panelX+panelW-12, y, 10, lightGray, alignEnd, alignStart)

	bet := g.m.TotalBet()
	notes := []string{
		fmt.Sprintf("JOKER: WILD ON REELS 2-5, PAYS x%d", machine.WildMultiplier),
		fmt.Sprintf("3+ DICE: FREE SPINS, WINS x%d", machine.FreeSpinMultiplier),
		fmt.Sprintf("%d+ CHIPS: LINK FEATURE", machine.LinkTrigger),
		fmt.Sprintf("DICE PAY %d / %d / %d",
			machine.ScatterPays[3]*bet, machine.ScatterPays[4]*bet, machine.ScatterPays[5]*bet),
	}
	for j, l := range notes {
		drawText(screen, l, panelX+10, y+14+float64(j)*12, 9, lightGray, alignStart, alignStart)
	}
}

// drawStat draws a labelled number under the given reel column.
func drawStat(screen *ebiten.Image, label string, value, column int, c color.Color) {
	x, _ := cellCentre(column, 0)
	drawText(screen, label, x, 370, 11, lightGray, alignCenter, alignStart)
	drawText(screen, fmt.Sprint(value), x, 384, 24, c, alignCenter, alignStart)
}

// drawButton draws a button with one or two lines of text. The selected
// button of a row is shown lit, and the others are dimmed when the row is
// locked.
func drawButton(screen *ebiten.Image, b button, selected, locked bool, size float64, lines ...string) {
	fill, edge, ink := color.Color(blue), color.Color(lightBlue), color.Color(white)
	switch {
	case selected:
		fill, edge, ink = yellow, white, black
	case locked:
		fill, edge, ink = black, darkGray, darkGray
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
