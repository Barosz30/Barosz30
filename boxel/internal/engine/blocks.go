package engine

type BlockID uint8

const (
	Air        BlockID = 0
	Grass      BlockID = 1
	Dirt       BlockID = 2
	Stone      BlockID = 3
	Brick      BlockID = 4
	Wood       BlockID = 5
	Sand       BlockID = 6
	Water      BlockID = 7
	Lava       BlockID = 8
	Ice        BlockID = 9
	Bounce     BlockID = 10
	Spike      BlockID = 11
	Coin       BlockID = 12
	Checkpoint BlockID = 13
	Spawn      BlockID = 14
	Goal       BlockID = 15
)

type RGB struct{ R, G, B uint8 }

func (c RGB) Float() (float32, float32, float32) {
	return float32(c.R) / 255, float32(c.G) / 255, float32(c.B) / 255
}

type BlockDef struct {
	ID     BlockID
	Name   string
	Hint   string
	Color  RGB
	Dark   RGB
	Light  RGB
	Solid  bool
	Hotbar bool
}

func rgb(n uint32) RGB {
	return RGB{R: uint8(n >> 16), G: uint8(n >> 8), B: uint8(n)}
}

var Blocks = []BlockDef{
	{0, "Gumka", "Usuwa klocki", rgb(0x3d3354), rgb(0x241b36), rgb(0x6a5a86), false, true},
	{1, "Trawa", "Zwykły klocek", rgb(0x5dad3e), rgb(0x3e7a28), rgb(0x8ad45e), true, true},
	{2, "Ziemia", "Zwykły klocek", rgb(0x8b5a2b), rgb(0x5e3b1b), rgb(0xc4894f), true, true},
	{3, "Kamień", "Zwykły klocek", rgb(0x8a8f99), rgb(0x5c616b), rgb(0xc5cad3), true, true},
	{4, "Cegła", "Zwykły klocek", rgb(0xc45c4a), rgb(0x8a3328), rgb(0xe08978), true, true},
	{5, "Drewno", "Zwykły klocek", rgb(0xc48a3a), rgb(0x8a5a1e), rgb(0xe0b56a), true, true},
	{6, "Piasek", "Zwykły klocek", rgb(0xe1c16e), rgb(0xb3903e), rgb(0xffe9a8), true, true},
	{7, "Woda", "Spowalnia", rgb(0x3a9ad9), rgb(0x1d6ea3), rgb(0x7ecfff), false, true},
	{8, "Lawa", "Odejmuje życie", rgb(0xff5a18), rgb(0xb32600), rgb(0xffb347), true, true},
	{9, "Lód", "Ślisko", rgb(0x9fe7ff), rgb(0x4fa9c8), rgb(0xe8fbff), true, true},
	{10, "Sprężyna", "Wyrzuca w górę", rgb(0xff4fa3), rgb(0xb21f6c), rgb(0xff9fd0), true, true},
	{11, "Kolce", "Odejmuje życie", rgb(0x3a3148), rgb(0x1a1424), rgb(0x6a5a80), true, true},
	{12, "Moneta", "Do zbierania", rgb(0xffd24a), rgb(0xc79212), rgb(0xfff1a3), false, true},
	{13, "Checkpoint", "Nowy start po upadku", rgb(0x7cff6b), rgb(0x2f9e2a), rgb(0xc8ffb8), true, true},
	{14, "Start", "Tu pojawia się gracz", rgb(0x5ec8ff), rgb(0x2178aa), rgb(0xb7e9ff), true, true},
	{15, "Meta", "Koniec poziomu", rgb(0xfff36a), rgb(0xc8b020), rgb(0xfff7b0), true, true},
}

func Def(id int) BlockDef {
	if id < 0 || id >= len(Blocks) {
		return Blocks[0]
	}
	return Blocks[id]
}

func IsSolid(id int) bool {
	if id == int(Air) {
		return false
	}
	return Def(id).Solid
}

func IsHazard(id int) bool {
	return id == int(Lava) || id == int(Spike)
}

func Hotbar() []BlockDef {
	out := make([]BlockDef, 0, len(Blocks))
	for _, b := range Blocks {
		if b.Hotbar {
			out = append(out, b)
		}
	}
	return out
}
