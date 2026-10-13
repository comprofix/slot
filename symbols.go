package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"slot/machine"
)

// The colour of each card symbol.
var cardInk = map[machine.Symbol]color.Color{
	machine.Ten:   lightCyan,
	machine.Jack:  lightGreen,
	machine.Queen: lightMagenta,
	machine.King:  lightBlue,
	machine.Ace:   white,
}

var cardText = map[machine.Symbol]string{
	machine.Ten:   "10",
	machine.Jack:  "J",
	machine.Queen: "Q",
	machine.King:  "K",
	machine.Ace:   "A",
}

// drawSymbol draws a symbol centred on (cx, cy). At scale 1 it fills a
// 120x90 cell; the reels, paytable and menu use smaller scales.
func drawSymbol(dst *ebiten.Image, sym machine.Symbol, cx, cy, scale float64) {
	// x and y turn offsets from the centre into screen positions.
	x := func(dx float64) float64 { return cx + dx*scale }
	y := func(dy float64) float64 { return cy + dy*scale }
	s := scale

	switch sym {
	case machine.Ten, machine.Jack, machine.Queen, machine.King, machine.Ace:
		drawText(dst, cardText[sym], x(3), y(3), 64*s, darkGray, alignCenter, alignCenter)
		drawText(dst, cardText[sym], x(0), y(0), 64*s, cardInk[sym], alignCenter, alignCenter)

	case machine.Cherry:
		drawLine(dst, x(-14), y(8), x(4), y(-32), 3*s, brown)
		drawLine(dst, x(14), y(10), x(4), y(-32), 3*s, brown)
		fillCircle(dst, x(14), y(-30), 7*s, green)
		fillCircle(dst, x(-14), y(12), 16*s, red)
		fillCircle(dst, x(14), y(14), 16*s, red)
		fillCircle(dst, x(-19), y(7), 4*s, lightRed)
		fillCircle(dst, x(9), y(9), 4*s, lightRed)

	case machine.Plum:
		drawLine(dst, x(4), y(-22), x(10), y(-38), 3*s, brown)
		fillCircle(dst, x(18), y(-34), 7*s, green)
		fillCircle(dst, x(0), y(4), 28*s, plumSkin)
		fillCircle(dst, x(-10), y(-7), 7*s, plumGlow)

	case machine.Bell:
		fillCircle(dst, x(0), y(-30), 5*s, gold)
		fillCircle(dst, x(0), y(28), 7*s, goldDark)
		fillCircle(dst, x(0), y(-8), 20*s, gold)
		fillPoly(dst, gold, x(-20), y(-8), x(20), y(-8), x(28), y(20), x(-28), y(20))
		fillRect(dst, x(-32), y(18), 64*s, 7*s, goldDark)

	case machine.Dollar:
		// The S is built from straight strokes, as in the original.
		w := 6 * s
		drawLine(dst, x(-18), y(-30), x(18), y(-30), w, yellow)
		drawLine(dst, x(-18), y(-33), x(-18), y(3), w, yellow)
		drawLine(dst, x(18), y(-33), x(18), y(-22), w, yellow)
		drawLine(dst, x(-18), y(0), x(18), y(0), w, yellow)
		drawLine(dst, x(18), y(-3), x(18), y(33), w, yellow)
		drawLine(dst, x(-18), y(30), x(18), y(30), w, yellow)
		drawLine(dst, x(-18), y(33), x(-18), y(22), w, yellow)
		drawLine(dst, x(-6), y(-40), x(-6), y(40), 3*s, yellow)
		drawLine(dst, x(7), y(-40), x(7), y(40), 3*s, yellow)

	case machine.Wild:
		// A jester's cap of three points, each tipped with a bell, over a
		// banner that says what the joker does.
		fillPoly(dst, lightMagenta, x(-36), y(10), x(-8), y(10), x(-42), y(-26))
		fillPoly(dst, lightMagenta, x(8), y(10), x(36), y(10), x(42), y(-26))
		fillPoly(dst, lightGreen, x(-16), y(10), x(16), y(10), x(0), y(-38))
		fillCircle(dst, x(-42), y(-28), 5*s, gold)
		fillCircle(dst, x(42), y(-28), 5*s, gold)
		fillCircle(dst, x(0), y(-40), 5*s, gold)
		fillRect(dst, x(-38), y(8), 76*s, 7*s, gold)
		fillRect(dst, x(-42), y(18), 84*s, 24*s, white)
		drawText(dst, "WILD", x(0), y(30), 20*s, red, alignCenter, alignCenter)

	case machine.Scatter:
		drawDie(dst, x(-17), y(-10), 19*s, -0.22, 5)
		drawDie(dst, x(17), y(2), 19*s, 0.18, 3)
		drawText(dst, "FREE GAMES", x(0), y(36), 12*s, yellow, alignCenter, alignCenter)

	case machine.Link:
		drawChip(dst, cx, cy, scale, "")
	}
}

// pips gives the spots on each face of a die, as positions on a 3x3 grid
// centred on the face.
var pips = map[int][][2]float64{
	3: {{-1, -1}, {0, 0}, {1, 1}},
	5: {{-1, -1}, {1, -1}, {0, 0}, {-1, 1}, {1, 1}},
}

// drawDie draws a die face centred on (cx, cy) and turned by angle radians.
// half is half the length of a side.
func drawDie(dst *ebiten.Image, cx, cy, half, angle float64, face int) {
	sin, cos := math.Sincos(angle)
	at := func(dx, dy float64) (float64, float64) {
		return cx + dx*cos - dy*sin, cy + dx*sin + dy*cos
	}
	square := func(h float64, c color.Color) {
		x0, y0 := at(-h, -h)
		x1, y1 := at(h, -h)
		x2, y2 := at(h, h)
		x3, y3 := at(-h, h)
		fillPoly(dst, c, x0, y0, x1, y1, x2, y2, x3, y3)
	}
	square(half, lightGray)
	square(half*0.9, white)
	for _, p := range pips[face] {
		px, py := at(p[0]*half*0.52, p[1]*half*0.52)
		fillCircle(dst, px, py, half*0.17, black)
	}
}

// drawChip draws a poker chip centred on (cx, cy) with a label in the
// middle: the credits or the jackpot it carries.
func drawChip(dst *ebiten.Image, cx, cy, scale float64, label string) {
	s := scale
	fillCircle(dst, cx, cy, 39*s, white)
	fillCircle(dst, cx, cy, 37*s, red)
	// White blocks round the rim.
	for i := range 8 {
		sin, cos := math.Sincos(float64(i) * math.Pi / 4)
		drawLine(dst, cx+cos*28*s, cy+sin*28*s, cx+cos*37*s, cy+sin*37*s, 10*s, white)
	}
	fillCircle(dst, cx, cy, 27*s, white)
	fillCircle(dst, cx, cy, 24*s, chipFelt)

	size := 20.0
	switch {
	case len(label) > 5:
		size = 11
	case len(label) > 3:
		size = 13
	}
	drawText(dst, label, cx, cy, size*s, yellow, alignCenter, alignCenter)
}
