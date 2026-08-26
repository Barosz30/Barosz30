package engine

import "testing"

func TestWorldSetAndRead(t *testing.T) {
	w := CreateWorld(8, 4, 8)
	if GetBlock(w, 1, 1, 1) != 0 {
		t.Fatal("empty")
	}
	SetBlock(w, 1, 1, 1, int(Grass))
	if GetBlock(w, 1, 1, 1) != int(Grass) {
		t.Fatal("set")
	}
	if w.Blocks[WorldIndex(w, 1, 1, 1)] != uint8(Grass) {
		t.Fatal("index")
	}
}

func TestWorldBounds(t *testing.T) {
	w := CreateWorld(4, 4, 4)
	if SetBlock(w, -1, 0, 0, int(Grass)) {
		t.Fatal("oob write")
	}
	if GetBlock(w, -1, 0, 0) != 0 {
		t.Fatal("oob read")
	}
}

func TestFillBox(t *testing.T) {
	w := CreateWorld(8, 4, 8)
	FillBox(w, 1, 0, 1, 3, 0, 2, int(Grass))
	if GetBlock(w, 1, 0, 1) != int(Grass) || GetBlock(w, 3, 0, 2) != int(Grass) {
		t.Fatal("fill")
	}
	if GetBlock(w, 4, 0, 2) != 0 {
		t.Fatal("outside fill")
	}
}

func TestSaveRoundtrip(t *testing.T) {
	w := CreateWorld(4, 4, 4)
	SetBlock(w, 0, 1, 0, int(Spawn))
	saved := SerializeGame("Test", w, nil)
	loaded, _, name := DeserializeGame(saved)
	if name != "Test" {
		t.Fatal(name)
	}
	if GetBlock(loaded, 0, 1, 0) != int(Spawn) {
		t.Fatal("spawn")
	}
}
