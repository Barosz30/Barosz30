package engine

import "testing"

func TestNodeLabelPolish(t *testing.T) {
	if got := NodeLabel(CodeNode{Kind: HatStart}); got != "gdy gra się zaczyna" {
		t.Fatal(got)
	}
	if got := NodeLabel(CodeNode{Kind: AddScore, Amount: 10}); got != "dodaj 10 pkt" {
		t.Fatal(got)
	}
	if got := NodeLabel(CodeNode{Kind: SetMessage, Text: "Hej"}); got != `wiadomość: "Hej"` {
		t.Fatal(got)
	}
}

func TestSanitizeScriptsStripsLinksAndHatsInBody(t *testing.T) {
	in := []ScriptStack{{
		ID: "a",
		Blocks: []CodeNode{
			{Kind: HatStart},
			{Kind: SetMessage, Text: "wejdź na http://zly.pl"},
			{Kind: AddScore, Amount: 999},
			{Kind: IfScore, Gte: 10, Body: []CodeNode{
				{Kind: HatCoin},
				{Kind: Win},
			}},
			{Kind: CodeKind("hack")},
		},
	}}
	out := SanitizeScripts(in)
	if len(out) != 1 {
		t.Fatalf("stacks %d", len(out))
	}
	blocks := out[0].Blocks
	if blocks[1].Text != "Cześć!" {
		t.Fatalf("link text %q", blocks[1].Text)
	}
	if blocks[2].Amount != 100 {
		t.Fatalf("amount %d", blocks[2].Amount)
	}
	if len(blocks[3].Body) != 1 || blocks[3].Body[0].Kind != Win {
		t.Fatalf("nested %#v", blocks[3].Body)
	}
	for _, n := range blocks {
		if n.Kind == "hack" {
			t.Fatal("unknown kind kept")
		}
	}
}

func TestAdjustNodeCycles(t *testing.T) {
	n := NewFromToolbox(AddScore)
	AdjustNode(&n, 1)
	if n.Amount != 11 {
		t.Fatal(n.Amount)
	}
	s := NewFromToolbox(Sound)
	AdjustNode(&s, 1)
	if s.Sound == SndCoin {
		t.Fatal("sound should cycle")
	}
}

func TestRunScriptsStartMessage(t *testing.T) {
	api := &mockAPI{}
	RunScripts(stack(
		MakeNode(HatStart, nil),
		MakeNode(SetMessage, &CodeNode{Text: "Start!"}),
		MakeNode(AddScore, &CodeNode{Amount: 5}),
	), ScriptEvent{Kind: "start"}, api)
	if api.message != "Start!" || api.score != 5 {
		t.Fatalf("%q %d", api.message, api.score)
	}
}
