package main

import (
	"bytes"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gobold"
)

// The 16-colour palette the Turbo Pascal Graph unit used.
var (
	black        = color.RGBA{0, 0, 0, 255}
	blue         = color.RGBA{0, 0, 170, 255}
	green        = color.RGBA{0, 170, 0, 255}
	red          = color.RGBA{170, 0, 0, 255}
	magenta      = color.RGBA{170, 0, 170, 255}
	brown        = color.RGBA{170, 85, 0, 255}
	lightGray    = color.RGBA{170, 170, 170, 255}
	darkGray     = color.RGBA{85, 85, 85, 255}
	lightBlue    = color.RGBA{85, 85, 255, 255}
	lightCyan    = color.RGBA{85, 255, 255, 255}
	lightGreen   = color.RGBA{85, 255, 85, 255}
	lightRed     = color.RGBA{255, 85, 85, 255}
	lightMagenta = color.RGBA{255, 85, 255, 255}
	yellow       = color.RGBA{255, 255, 85, 255}
	white        = color.RGBA{255, 255, 255, 255}
)

// Extra colours the original palette has no good match for.
var (
	reelBack = color.RGBA{16, 16, 24, 255}
	plumSkin = color.RGBA{120, 70, 220, 255}
	plumGlow = color.RGBA{175, 135, 255, 255}
	gold     = color.RGBA{255, 170, 0, 255}
	goldDark = color.RGBA{190, 110, 0, 255}
	chipFelt = color.RGBA{0, 0, 96, 255}
)

var fontSource *text.GoTextFaceSource

func init() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if err != nil {
		log.Fatal(err)
	}
	fontSource = src
}

const (
	alignStart  = text.AlignStart
	alignCenter = text.AlignCenter
	alignEnd    = text.AlignEnd
)

// drawText draws s with (x, y) placed according to the two alignments:
// h along the line, v down the page.
func drawText(dst *ebiten.Image, s string, x, y, size float64, c color.Color, h, v text.Align) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	op.PrimaryAlign = h
	op.SecondaryAlign = v
	text.Draw(dst, s, &text.GoTextFace{Source: fontSource, Size: size}, op)
}

func fillRect(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), c, false)
}

func frameRect(dst *ebiten.Image, x, y, w, h, width float64, c color.Color) {
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), float32(width), c, false)
}

func fillCircle(dst *ebiten.Image, cx, cy, r float64, c color.Color) {
	vector.FillCircle(dst, float32(cx), float32(cy), float32(r), c, true)
}

func drawLine(dst *ebiten.Image, x0, y0, x1, y1, width float64, c color.Color) {
	vector.StrokeLine(dst, float32(x0), float32(y0), float32(x1), float32(y1), float32(width), c, true)
}

// fillPoly fills the polygon whose corners are given as x, y pairs.
func fillPoly(dst *ebiten.Image, c color.Color, pts ...float64) {
	var p vector.Path
	p.MoveTo(float32(pts[0]), float32(pts[1]))
	for i := 2; i < len(pts); i += 2 {
		p.LineTo(float32(pts[i]), float32(pts[i+1]))
	}
	p.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(dst, &p, nil, op)
}
