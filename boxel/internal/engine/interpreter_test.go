package engine

import "testing"

type mockAPI struct {
	score, teleports int
	stopped          bool
	message          string
	won, lost        bool
	sounds           []SoundName
}

func (m *mockAPI) AddScore(n int)           { m.score += n }
func (m *mockAPI) GetScore() int            { return m.score }
func (m *mockAPI) SetMessage(text string)   { m.message = text }
func (m *mockAPI) DoWin()                   { m.won = true; m.stopped = true }
func (m *mockAPI) DoLose()                  { m.lost = true; m.stopped = true }
func (m *mockAPI) TeleportSpawn()           { m.teleports++ }
func (m *mockAPI) PlaySound(name SoundName) { m.sounds = append(m.sounds, name) }
func (m *mockAPI) IsStopped() bool          { return m.stopped }

func stack(nodes ...CodeNode) []ScriptStack {
	return []ScriptStack{{ID: "s1", Blocks: nodes}}
}

func TestInterpreterScoreOnStart(t *testing.T) {
	api := &mockAPI{}
	vm := &ScriptVM{}
	vm.SetScripts(stack(MakeNode(HatStart, nil), MakeNode(AddScore, &CodeNode{Amount: 10})))
	vm.Emit(ScriptEvent{Kind: "start"}, api)
	vm.Tick(0.016, api)
	if api.score != 10 {
		t.Fatalf("score %d", api.score)
	}
}

func TestIfScoreThreshold(t *testing.T) {
	api := &mockAPI{}
	vm := &ScriptVM{}
	vm.SetScripts(stack(
		MakeNode(HatGoal, nil),
		MakeNode(IfScore, &CodeNode{Gte: 30, Body: []CodeNode{MakeNode(Win, nil)}}),
		MakeNode(SetMessage, &CodeNode{Text: "za mało"}),
	))
	vm.Emit(ScriptEvent{Kind: "goal"}, api)
	vm.Tick(0.016, api)
	if api.won || api.message != "za mało" {
		t.Fatalf("first %#v %s", api.won, api.message)
	}
	api.score = 30
	vm.Emit(ScriptEvent{Kind: "goal"}, api)
	vm.Tick(0.016, api)
	if !api.won {
		t.Fatal("expected win")
	}
}

func TestRepeatCap(t *testing.T) {
	api := &mockAPI{}
	vm := &ScriptVM{}
	vm.SetScripts(stack(
		MakeNode(HatStart, nil),
		MakeNode(Repeat, &CodeNode{Times: 50, Body: []CodeNode{MakeNode(AddScore, &CodeNode{Amount: 1})}}),
	))
	vm.Emit(ScriptEvent{Kind: "start"}, api)
	vm.Tick(0.016, api)
	if api.score != 20 {
		t.Fatalf("score %d", api.score)
	}
}

func TestInstructionBudget(t *testing.T) {
	api := &mockAPI{}
	vm := &ScriptVM{}
	vm.SetScripts(stack(
		MakeNode(HatAlways, nil),
		MakeNode(Repeat, &CodeNode{Times: 20, Body: []CodeNode{MakeNode(AddScore, &CodeNode{Amount: 1})}}),
	))
	vm.Tick(0.016, api)
	if api.score <= 0 || api.score > 80 {
		t.Fatalf("score %d", api.score)
	}
}

func TestTouchMatch(t *testing.T) {
	api := &mockAPI{}
	vm := &ScriptVM{}
	vm.SetScripts(stack(MakeNode(HatTouch, &CodeNode{Block: 8}), MakeNode(AddScore, &CodeNode{Amount: 5})))
	vm.Emit(ScriptEvent{Kind: "touch", Block: 3}, api)
	vm.Tick(0.016, api)
	if api.score != 0 {
		t.Fatal(api.score)
	}
	vm.Emit(ScriptEvent{Kind: "touch", Block: 8}, api)
	vm.Tick(0.016, api)
	if api.score != 5 {
		t.Fatal(api.score)
	}
}
