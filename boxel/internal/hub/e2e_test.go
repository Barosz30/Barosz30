package hub

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWSFriendsRoomChatPlace(t *testing.T) {
	srv := httptest.NewServer(Handler(NewHub()))
	defer srv.Close()
	u := "ws" + strings.TrimPrefix(srv.URL, "http")

	dial := func(name string) *websocket.Conn {
		c, _, err := websocket.DefaultDialer.Dial(u, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := c.WriteJSON(map[string]any{"type": "hello", "name": name}); err != nil {
			t.Fatal(err)
		}
		return c
	}
	wait := func(c *websocket.Conn, typ string) map[string]any {
		t.Helper()
		for i := 0; i < 20; i++ {
			_, raw, err := c.ReadMessage()
			if err != nil {
				t.Fatal(err)
			}
			var m map[string]any
			_ = json.Unmarshal(raw, &m)
			if m["type"] == typ {
				return m
			}
		}
		t.Fatalf("no %s", typ)
		return nil
	}

	a := dial("Ania")
	b := dial("Kuba")
	defer a.Close()
	defer b.Close()
	wait(a, "helloOk")
	hb := wait(b, "helloOk")
	you := hb["you"].(map[string]any)
	_ = a.WriteJSON(map[string]any{"type": "addFriend", "code": you["friendCode"]})
	wait(a, "friends")
	_ = a.WriteJSON(map[string]any{
		"type": "host", "name": "Test", "sizeX": 2, "sizeY": 2, "sizeZ": 2,
		"blocks": []int{0, 0, 0, 0, 0, 0, 0, 0},
		"scripts": []map[string]any{
			{"id": "s", "blocks": []map[string]any{
				{"kind": "hatStart"},
				{"kind": "setMessage", "text": "Cześć!"},
			}},
		},
	})
	room := wait(a, "room")
	if room["scripts"] == nil {
		t.Fatal("host room missing scripts")
	}
	_ = b.WriteJSON(map[string]any{"type": "join", "code": room["code"]})
	joined := wait(b, "room")
	if joined["scripts"] == nil {
		t.Fatal("join room missing scripts")
	}
	_ = a.WriteJSON(map[string]any{"type": "chat", "text": "Cześć!"})
	chat := wait(b, "chat")
	line := chat["line"].(map[string]any)
	if line["text"] != "Cześć!" {
		t.Fatalf("chat %#v", line)
	}
	_ = a.WriteJSON(map[string]any{"type": "place", "x": 1, "y": 0, "z": 0, "id": 4})
	place := wait(b, "place")
	if place["id"].(float64) != 4 {
		t.Fatalf("place %#v", place)
	}
}
