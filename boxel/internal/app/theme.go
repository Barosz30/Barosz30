package app

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

var (
	colBg     = color.RGBA{0x1a, 0x12, 0x28, 0xff}
	colPanel  = color.RGBA{0x31, 0x24, 0x4d, 0xff}
	colInk    = color.RGBA{0xff, 0xf6, 0xe8, 0xff}
	colMuted  = color.RGBA{0xcb, 0xbc, 0xe3, 0xff}
	colLine   = color.RGBA{0x0d, 0x08, 0x14, 0xff}
	colAccent = color.RGBA{0xff, 0xe1, 0x4a, 0xff}
	colPlay   = color.RGBA{0x3f, 0xd0, 0x7a, 0xff}
	colDanger = color.RGBA{0xff, 0x5a, 0x6a, 0xff}
	colBtn    = color.RGBA{0x4a, 0x3a, 0x68, 0xff}
)

var face14, face12, face18, face10 *text.GoXFace

func initFonts() {
	if face14 != nil {
		return
	}
	ttf, err := opentype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}
	mk := func(size float64) *text.GoXFace {
		f, err := opentype.NewFace(ttf, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
		if err != nil {
			panic(err)
		}
		return text.NewGoXFace(f)
	}
	face10, face12, face14, face18 = mk(11), mk(13), mk(15), mk(20)
}

type Rect struct{ X, Y, W, H int }

func (r Rect) Contains(x, y int) bool {
	return x >= r.X && y >= r.Y && x < r.X+r.W && y < r.Y+r.H
}

func fillRect(dst *ebiten.Image, r Rect, c color.Color) {
	vector.DrawFilledRect(dst, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), c, false)
}

func strokeRect(dst *ebiten.Image, r Rect, c color.Color, w float32) {
	vector.StrokeRect(dst, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), w, c, false)
}

func panel(dst *ebiten.Image, r Rect, fill color.Color) {
	fillRect(dst, Rect{r.X + 3, r.Y + 3, r.W, r.H}, colLine)
	fillRect(dst, r, fill)
	strokeRect(dst, r, colLine, 3)
}

func drawText(dst *ebiten.Image, s string, x, y int, face *text.GoXFace, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, face, op)
}

type click struct {
	x, y     int
	pressed  bool
	justDown bool
	justUp   bool
	right    bool
	dx, dy   int
}

func button(dst *ebiten.Image, r Rect, label string, on bool, fill color.Color, mouse click) bool {
	c := fill
	if on {
		c = colAccent
	}
	if mouse.pressed && r.Contains(mouse.x, mouse.y) {
		r = Rect{r.X + 2, r.Y + 2, r.W, r.H}
	}
	panel(dst, r, c)
	ink := colInk
	if on {
		ink = colBg
	}
	tw, th := text.Measure(label, face12, 0)
	drawText(dst, label, r.X+(r.W-int(tw))/2, r.Y+(r.H-int(th))/2, face12, ink)
	return mouse.justUp && r.Contains(mouse.x, mouse.y)
}
