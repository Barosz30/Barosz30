package hub

import (
	"encoding/json"
	"testing"

	"boxel/internal/engine"
)

func drain(p *Player) []map[string]any {
	out := []map[string]any{}
	for {
		select {
		case b := <-p.send:
			var m map[string]any
			_ = json.Unmarshal(b, &m)
			out = append(out, m)
		default:
			return out
		}
	}
}

func lastOfType(msgs []map[string]any, typ string) map[string]any {
	var last map[string]any
	for _, m := range msgs {
		if m["type"] == typ {
			last = m
		}
	}
	return last
}

func TestSanitizeChat(t *testing.T) {
	if got, ok := sanitizeChat("Cześć! Chodź tu"); !ok || got != "Cześć! Chodź tu" {
		t.Fatalf("allowed chat: %q %v", got, ok)
	}
	if _, ok := sanitizeChat("wejdź na discord.gg/xyz"); ok {
		t.Fatal("discord should be blocked")
	}
	if _, ok := sanitizeChat("http://zly-link.pl"); ok {
		t.Fatal("url should be blocked")
	}
	if _, ok := sanitizeChat("kurwa lawy"); ok {
		t.Fatal("banned word")
	}
}

func TestSanitizeName(t *testing.T) {
	if sanitizeName("Ania!!!") != "Ania" {
		t.Fatal(sanitizeName("Ania!!!"))
	}
	if sanitizeName("x") != "Gracz" {
		t.Fatal("short name")
	}
}

func TestFriendCode(t *testing.T) {
	c := makeFriendCode()
	if !validFriendCode(c) {
		t.Fatalf("bad code %s", c)
	}
}

func TestFriendsRoomChatPlace(t *testing.T) {
	h := NewHub()
	a := h.Connect(inbound{Type: "hello", Name: "Ania"}, "", nil)
	b := h.Connect(inbound{Type: "hello", Name: "Kuba"}, "", nil)
	drain(a)
	h.Handle(a, inbound{Type: "addFriend", Code: b.friendCode}, 0, "")
	msgs := drain(a)
	friends := lastOfType(msgs, "friends")
	if friends == nil {
		t.Fatal("no friends message")
	}
	list, _ := friends["friends"].([]any)
	found := false
	for _, item := range list {
		m := item.(map[string]any)
		if m["name"] == "Kuba" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Kuba missing: %#v", friends)
	}

	h.Handle(a, inbound{Type: "host", Name: "Parkour", SizeX: 2, SizeY: 2, SizeZ: 2, Blocks: []int{0, 0, 0, 0, 0, 0, 0, 0}}, 0, "")
	room := lastOfType(drain(a), "room")
	if room == nil {
		t.Fatal("no room")
	}
	code, _ := room["code"].(string)
	if len(code) != 4 {
		t.Fatalf("code %s", code)
	}

	h.Handle(b, inbound{Type: "join", Code: code}, 0, "")
	drain(b)
	h.Handle(a, inbound{Type: "chat", Text: "Cześć!"}, 0, "")
	chat := lastOfType(drain(b), "chat")
	if chat == nil {
		t.Fatal("no chat")
	}
	line := chat["line"].(map[string]any)
	if line["text"] != "Cześć!" {
		t.Fatalf("chat %v", line)
	}

	h.Handle(a, inbound{Type: "place", X: 1, Y: 0, Z: 0}, 4, "")
	place := lastOfType(drain(b), "place")
	if place == nil || place["id"].(float64) != 4 {
		t.Fatalf("place %v", place)
	}
}

func TestWhisperRequiresFriends(t *testing.T) {
	h := NewHub()
	a := h.Connect(inbound{Type: "hello", Name: "Ania"}, "", nil)
	b := h.Connect(inbound{Type: "hello", Name: "Kuba"}, "", nil)
	drain(a)
	drain(b)
	h.Handle(a, inbound{Type: "chat", Text: "psst", ToID: b.id}, 0, "")
	if lastOfType(drain(a), "error") == nil {
		t.Fatal("expected error")
	}
	if lastOfType(drain(b), "chat") != nil {
		t.Fatal("whisper leaked")
	}
}

func TestParsePlaceID(t *testing.T) {
	msg, blockID, stringID := parseInbound([]byte(`{"type":"place","x":1,"y":0,"z":0,"id":4}`))
	if msg.Type != "place" || blockID != 4 || stringID != "" {
		t.Fatalf("place parse %v %d %q", msg.Type, blockID, stringID)
	}
	msg, blockID, stringID = parseInbound([]byte(`{"type":"hello","id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","name":"Ania"}`))
	if msg.Type != "hello" || blockID != 0 || stringID != "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" {
		t.Fatalf("hello parse %v %d %q", msg.Type, blockID, stringID)
	}
}

func TestReconnectKeepsFriends(t *testing.T) {
	h := NewHub()
	a := h.Connect(inbound{Type: "hello", Name: "Ania"}, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", nil)
	b := h.Connect(inbound{Type: "hello", Name: "Kuba"}, "11111111-2222-3333-4444-555555555555", nil)
	drain(a)
	h.Handle(a, inbound{Type: "addFriend", Code: b.friendCode}, 0, "")
	again := h.Connect(inbound{Type: "hello", Name: "Ania"}, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", nil)
	if again.id != a.id {
		t.Fatal("id changed")
	}
	friends := lastOfType(drain(again), "friends")
	if friends == nil {
		t.Fatal("no friends after reconnect")
	}
	list, _ := friends["friends"].([]any)
	if len(list) != 1 {
		t.Fatalf("friends %v", friends)
	}
}

func TestRoomScriptsLocalAndServer(t *testing.T) {
	h := NewHub()
	a := h.Connect(inbound{Type: "hello", Name: "Ania"}, "", nil)
	b := h.Connect(inbound{Type: "hello", Name: "Kuba"}, "", nil)
	drain(a)
	drain(b)

	scripts := []engine.ScriptStack{{
		ID: "s1",
		Blocks: []engine.CodeNode{
			{Kind: engine.HatStart, Text: ""},
			{Kind: engine.SetMessage, Text: "Cześć z pokoju"},
			{Kind: engine.AddScore, Amount: 10},
			{Kind: engine.HatCoin},
		},
	}, {
		ID: "s2",
		Blocks: []engine.CodeNode{
			{Kind: engine.HatCoin},
			{Kind: engine.AddScore, Amount: 10},
			{Kind: engine.SetMessage, Text: "Moneta!"},
		},
	}}
	h.Handle(a, inbound{Type: "host", Name: "Parkour", SizeX: 2, SizeY: 2, SizeZ: 2, Blocks: []int{0, 0, 0, 0, 0, 0, 0, 0}, Scripts: scripts}, 0, "")
	room := lastOfType(drain(a), "room")
	if room == nil {
		t.Fatal("no room")
	}
	if _, ok := room["scripts"]; !ok {
		t.Fatalf("host room missing scripts: %#v", room)
	}
	code, _ := room["code"].(string)
	h.Handle(b, inbound{Type: "join", Code: code}, 0, "")
	joined := lastOfType(drain(b), "room")
	if joined == nil {
		t.Fatal("no join room")
	}
	raw, _ := json.Marshal(joined["scripts"])
	var got []engine.ScriptStack
	_ = json.Unmarshal(raw, &got)
	if len(got) != 2 {
		t.Fatalf("joiner scripts %d %s", len(got), raw)
	}
	if got[1].Blocks[0].Kind != engine.HatCoin {
		t.Fatalf("coin hat missing: %#v", got[1])
	}
	if asBoolMap(joined["host"]) {
		t.Fatal("joiner should not be host")
	}

	h.Handle(b, inbound{Type: "scripts", Scripts: []engine.ScriptStack{{ID: "x", Blocks: []engine.CodeNode{{Kind: engine.Win}}}}}, 0, "")
	if lastOfType(drain(b), "error") == nil {
		t.Fatal("joiner must not change room scripts")
	}

	updated := []engine.ScriptStack{{
		ID: "s3",
		Blocks: []engine.CodeNode{
			{Kind: engine.HatGoal},
			{Kind: engine.Win},
		},
	}}
	h.Handle(a, inbound{Type: "scripts", Scripts: updated}, 0, "")
	push := lastOfType(drain(b), "scripts")
	if push == nil {
		t.Fatal("joiner did not get script update")
	}

	h.Handle(a, inbound{Type: "scriptEvent", Kind: "coin", X: 1, Y: 0, Z: 0}, 0, "")
	after := drain(b)
	if lastOfType(after, "scriptFx") == nil {
		t.Fatal("no scriptFx")
	}
	place := lastOfType(after, "place")
	if place == nil || place["id"].(float64) != 0 {
		t.Fatalf("coin should clear block, got %#v", place)
	}

	api := EvalScripts(updated, engine.ScriptEvent{Kind: "goal"})
	if !api.won {
		t.Fatal("expected server eval win")
	}
}

func asBoolMap(v any) bool {
	b, _ := v.(bool)
	return b
}
