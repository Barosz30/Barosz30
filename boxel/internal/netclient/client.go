package netclient

import (
	"encoding/json"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"boxel/internal/engine"
	"boxel/internal/hub"
)

type Handler func(map[string]any)

type Client struct {
	URL      string
	You      hub.Identity
	Color    string
	mu       sync.Mutex
	conn     *websocket.Conn
	handlers []Handler
	queue    []any
	profile  engine.Profile
	stop     chan struct{}
}

func New(profile engine.Profile, serverURL string) *Client {
	if serverURL == "" {
		serverURL = "ws://127.0.0.1:8787"
	}
	return &Client{URL: serverURL, profile: profile, stop: make(chan struct{})}
}

func (c *Client) On(h Handler) {
	c.mu.Lock()
	c.handlers = append(c.handlers, h)
	c.mu.Unlock()
}

func (c *Client) Connect() {
	go c.loop()
}

func (c *Client) loop() {
	for {
		select {
		case <-c.stop:
			return
		default:
		}
		u, err := url.Parse(c.URL)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			time.Sleep(1200 * time.Millisecond)
			continue
		}
		c.mu.Lock()
		c.conn = conn
		hello := map[string]any{"type": "hello", "id": c.profile.ID, "name": c.profile.Name, "friendCode": c.profile.FriendCode}
		_ = conn.WriteJSON(hello)
		for _, msg := range c.queue {
			_ = conn.WriteJSON(msg)
		}
		c.queue = nil
		c.mu.Unlock()

		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				break
			}
			var msg map[string]any
			if json.Unmarshal(raw, &msg) != nil {
				continue
			}
			if msg["type"] == "helloOk" {
				if you, ok := msg["you"].(map[string]any); ok {
					c.You = hub.Identity{
						ID:         str(you["id"]),
						Name:       str(you["name"]),
						FriendCode: str(you["friendCode"]),
					}
					c.profile.ID = c.You.ID
					c.profile.Name = c.You.Name
					c.profile.FriendCode = c.You.FriendCode
					_ = engine.SaveProfile(c.profile)
				}
				c.Color = str(msg["color"])
			}
			c.mu.Lock()
			hs := append([]Handler(nil), c.handlers...)
			c.mu.Unlock()
			for _, h := range hs {
				h(msg)
			}
		}
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		_ = conn.Close()
		time.Sleep(1200 * time.Millisecond)
	}
}

func (c *Client) Disconnect() {
	close(c.stop)
	c.mu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()
}

func (c *Client) Send(msg any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.WriteJSON(msg)
		return
	}
	c.queue = append(c.queue, msg)
}

func (c *Client) Profile() engine.Profile {
	return c.profile
}

func str(v any) string {
	s, _ := v.(string)
	return s
}
