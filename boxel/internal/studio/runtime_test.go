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
	rt.Play()
	rt.Player.X, rt.Player.Y, rt.Player.Z = 4.5, 1, 4.5
	for i := 0; i < 8; i++ {
		rt.Tick(0.016, engine.InputState{})
	}
	if rt.Player.Coins < 1 {
		t.Fatalf("coins=%d score=%d", rt.Player.Coins, rt.Player.Score)
	}
}
