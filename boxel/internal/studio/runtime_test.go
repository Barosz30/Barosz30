package studio

import (
	"testing"

	"boxel/internal/engine"
)

func TestPlayPicksUpCoin(t *testing.T) {
	w := engine.CreateWorld(8, 8, 8)
	engine.FillBox(w, 0, 0, 0, 7, 0, 7, int(engine.Grass))
	engine.SetBlock(w, 3, 1, 4, int(engine.Spawn))
	engine.SetBlock(w, 4, 1, 4, int(engine.Coin))
	scripts := []engine.ScriptStack{
		{ID: "c", Blocks: []engine.CodeNode{
			engine.MakeNode(engine.HatCoin, nil),
			engine.MakeNode(engine.AddScore, &engine.CodeNode{Amount: 10}),
		}},
	}
	rt := New(w, scripts, "t")
	var edits [][4]int
	var events []string
	rt.OnEdit = func(x, y, z, id int) { edits = append(edits, [4]int{x, y, z, id}) }
	rt.OnScriptEvent = func(kind string, x, y, z, block int) { events = append(events, kind) }
	rt.Play()
	rt.Player.X, rt.Player.Y, rt.Player.Z = 4.5, 1, 4.5
	for i := 0; i < 8; i++ {
		rt.Tick(0.016, engine.InputState{})
	}
	if rt.Player.Coins < 1 {
		t.Fatalf("coins=%d score=%d", rt.Player.Coins, rt.Player.Score)
	}
	if rt.Player.Score != 10 {
		t.Fatalf("score %d", rt.Player.Score)
	}
	if len(edits) != 1 || edits[0][3] != 0 {
		t.Fatalf("expected air edit, got %#v", edits)
	}
	if len(events) != 1 || events[0] != "coin" {
		t.Fatalf("script events %#v", events)
	}
}

func TestPlayRunsStartScript(t *testing.T) {
	w := engine.CreateWorld(8, 8, 8)
	engine.FillBox(w, 0, 0, 0, 7, 0, 7, int(engine.Grass))
	engine.SetBlock(w, 3, 1, 4, int(engine.Spawn))
	scripts := []engine.ScriptStack{
		{ID: "s", Blocks: []engine.CodeNode{
			engine.MakeNode(engine.HatStart, nil),
			engine.MakeNode(engine.SetMessage, &engine.CodeNode{Text: "Start!"}),
			engine.MakeNode(engine.AddScore, &engine.CodeNode{Amount: 7}),
		}},
	}
	rt := New(w, scripts, "t")
	rt.Play()
	if rt.Player.Message != "Start!" {
		t.Fatalf("message %q", rt.Player.Message)
	}
	if rt.Player.Score != 7 {
		t.Fatalf("score %d", rt.Player.Score)
	}
}
