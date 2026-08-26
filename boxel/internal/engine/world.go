package engine

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
)

const (
	SizeX = 24
	SizeY = 16
	SizeZ = 24
)

func NewID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func MakeNode(kind CodeKind, extra *CodeNode) CodeNode {
	n := CodeNode{ID: NewID(), Kind: kind}
	if extra != nil {
		n.Amount = extra.Amount
		n.Text = extra.Text
		n.Block = extra.Block
		n.Sound = extra.Sound
		n.Seconds = extra.Seconds
		n.Gte = extra.Gte
		n.Times = extra.Times
		n.Body = extra.Body
	}
	return n
}

func CreateWorld(sizeX, sizeY, sizeZ int) World {
	if sizeX <= 0 {
		sizeX = SizeX
	}
	if sizeY <= 0 {
		sizeY = SizeY
	}
	if sizeZ <= 0 {
		sizeZ = SizeZ
	}
	return World{
		SizeX:  sizeX,
		SizeY:  sizeY,
		SizeZ:  sizeZ,
		Blocks: make([]uint8, sizeX*sizeY*sizeZ),
	}
}

func WorldIndex(w World, x, y, z int) int {
	return x + w.SizeX*(z+w.SizeZ*y)
}

func InBounds(w World, x, y, z int) bool {
	return x >= 0 && y >= 0 && z >= 0 && x < w.SizeX && y < w.SizeY && z < w.SizeZ
}

func GetBlock(w World, x, y, z int) int {
	if !InBounds(w, x, y, z) {
		return int(Air)
	}
	return int(w.Blocks[WorldIndex(w, x, y, z)])
}

func SetBlock(w World, x, y, z, id int) bool {
	if !InBounds(w, x, y, z) {
		return false
	}
	w.Blocks[WorldIndex(w, x, y, z)] = uint8(id)
	return true
}

func FillBox(w World, x1, y1, z1, x2, y2, z2, id int) {
	xa, xb := min(x1, x2), max(x1, x2)
	ya, yb := min(y1, y2), max(y1, y2)
	za, zb := min(z1, z2), max(z1, z2)
	for y := ya; y <= yb; y++ {
		for z := za; z <= zb; z++ {
			for x := xa; x <= xb; x++ {
				SetBlock(w, x, y, z, id)
			}
		}
	}
}

func CloneWorld(w World) World {
	cp := CreateWorld(w.SizeX, w.SizeY, w.SizeZ)
	copy(cp.Blocks, w.Blocks)
	return cp
}

func CopyWorld(from World, to *World) {
	to.SizeX, to.SizeY, to.SizeZ = from.SizeX, from.SizeY, from.SizeZ
	to.Blocks = append([]uint8(nil), from.Blocks...)
}

func FindSpawn(w World) (x, y, z float64) {
	for y := 0; y < w.SizeY; y++ {
		for z := 0; z < w.SizeZ; z++ {
			for x := 0; x < w.SizeX; x++ {
				if GetBlock(w, x, y, z) == int(Spawn) {
					return float64(x) + 0.5, float64(y) + 1, float64(z) + 0.5
				}
			}
		}
	}
	for y := w.SizeY - 2; y >= 0; y-- {
		for z := 0; z < w.SizeZ; z++ {
			for x := 0; x < w.SizeX; x++ {
				if GetBlock(w, x, y, z) != int(Air) && GetBlock(w, x, y+1, z) == int(Air) {
					return float64(x) + 0.5, float64(y) + 1, float64(z) + 0.5
				}
			}
		}
	}
	return float64(w.SizeX) / 2, 4, float64(w.SizeZ) / 2
}

func SerializeGame(name string, w World, scripts []ScriptStack) SavedGame {
	blocks := make([]int, len(w.Blocks))
	for i, b := range w.Blocks {
		blocks[i] = int(b)
	}
	return SavedGame{
		Version: 1,
		Name:    name,
		SizeX:   w.SizeX,
		SizeY:   w.SizeY,
		SizeZ:   w.SizeZ,
		Blocks:  blocks,
		Scripts: CloneScripts(scripts),
	}
}

func DeserializeGame(data SavedGame) (World, []ScriptStack, string) {
	w := CreateWorld(data.SizeX, data.SizeY, data.SizeZ)
	n := min(len(w.Blocks), len(data.Blocks))
	for i := 0; i < n; i++ {
		w.Blocks[i] = uint8(data.Blocks[i])
	}
	name := data.Name
	if name == "" {
		name = "Moja gra"
	}
	return w, CloneScripts(data.Scripts), name
}

func CloneScripts(scripts []ScriptStack) []ScriptStack {
	b, _ := json.Marshal(scripts)
	var out []ScriptStack
	_ = json.Unmarshal(b, &out)
	if out == nil {
		return []ScriptStack{}
	}
	return out
}
