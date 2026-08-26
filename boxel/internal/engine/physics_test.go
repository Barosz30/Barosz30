package engine

import "testing"

func still() InputState { return InputState{} }

func TestStandOnGrass(t *testing.T) {
	w := CreateWorld(SizeX, SizeY, SizeZ)
	FillBox(w, 0, 0, 0, 10, 0, 10, int(Grass))
	p := CreatePlayer()
	p.X, p.Y, p.Z = 5.5, 1, 5.5
	for i := 0; i < 40; i++ {
		var ev []PhysEvent
		StepPlayer(p, w, still(), 0.016, &ev)
	}
	if p.Y <= 0.95 || p.Y >= 1.2 || !p.OnGround {
		t.Fatalf("y=%v ground=%v", p.Y, p.OnGround)
	}
}

func TestCoinPickup(t *testing.T) {
	w := CreateWorld(SizeX, SizeY, SizeZ)
	FillBox(w, 0, 0, 0, 10, 0, 10, int(Grass))
	SetBlock(w, 5, 1, 5, int(Coin))
	p := CreatePlayer()
	p.X, p.Y, p.Z = 5.5, 1, 5.5
	var ev []PhysEvent
	StepPlayer(p, w, still(), 0.016, &ev)
	found := false
	for _, e := range ev {
		if e.Type == "coin" {
			found = true
		}
	}
	if !found {
		t.Fatalf("events %#v", ev)
	}
}

func TestParkourTemplate(t *testing.T) {
	w, scripts, name := LoadTemplate("parkour")
	if name == "" || scripts[0].Blocks[0].Kind != HatStart {
		t.Fatal("scripts")
	}
	var spawn, lava, coin, goal int
	for _, id := range w.Blocks {
		switch BlockID(id) {
		case Spawn:
			spawn++
		case Lava:
			lava++
		case Coin:
			coin++
		case Goal:
			goal++
		}
	}
	if spawn == 0 || lava <= 10 || coin < 3 || goal != 1 {
		t.Fatalf("counts spawn=%d lava=%d coin=%d goal=%d", spawn, lava, coin, goal)
	}
	if GetBlock(w, 3, 2, 3) != int(Spawn) {
		t.Fatal("spawn cell")
	}
}

func TestPhysicsSpeed(t *testing.T) {
	w := CreateWorld(SizeX, SizeY, SizeZ)
	FillBox(w, 0, 0, 0, 12, 0, 12, int(Grass))
	p := CreatePlayer()
	p.X, p.Y, p.Z = 6.5, 1, 6.5
	for i := 0; i < 2000; i++ {
		var ev []PhysEvent
		StepPlayer(p, w, still(), 0.016, &ev)
	}
	if !p.OnGround {
		t.Fatal("should stand")
	}
}
