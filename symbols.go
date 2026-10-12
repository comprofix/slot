package main

import (
	"github.com/hajimehoshi/ebiten/v2"

	"slot/machine"
)

// drawSymbol draws a symbol centred on (cx, cy). At scale 1 it fits a reel
// cell; the paytable and menu use smaller scales.
func drawSymbol(dst *ebiten.Image, sym machine.Symbol, cx, cy, scale float64) {
	// x and y turn offsets from the centre into screen positions.
	x := func(dx float64) float64 { return cx + dx*scale }
	y := func(dy float64) float64 { return cy + dy*scale }
	s := scale

	switch sym {
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

	case machine.Bar:
		fillRect(dst, x(-48), y(-24), 96*s, 48*s, yellow)
		frameRect(dst, x(-43), y(-19), 86*s, 38*s, 2*s, red)
		drawText(dst, "BAR", x(0), y(0), 32*s, red, alignCenter, alignCenter)

	case machine.Seven:
		drawText(dst, "7", x(3), y(3), 76*s, red, alignCenter, alignCenter)
		drawText(dst, "7", x(0), y(0), 76*s, lightRed, alignCenter, alignCenter)

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
	}
}
