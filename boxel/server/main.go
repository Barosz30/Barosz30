package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type inbound struct {
	Type       string          `json:"type"`
	ID         json.RawMessage `json:"id"`
	Name       string          `json:"name"`
	FriendCode string          `json:"friendCode"`
	Code       string          `json:"code"`
	SizeX      int             `json:"sizeX"`
	SizeY      int             `json:"sizeY"`
	SizeZ      int             `json:"sizeZ"`
	Blocks     []int           `json:"blocks"`
	X          float64         `json:"x"`
	Y          float64         `json:"y"`
	Z          float64         `json:"z"`
	Yaw        float64         `json:"yaw"`
	Text       string          `json:"text"`
	ToID       string          `json:"toId"`
	FriendID   string          `json:"friendId"`
}

func parseInbound(raw []byte) (inbound, int, string) {
	var msg inbound
	_ = json.Unmarshal(raw, &msg)
	blockID := 0
	stringID := ""
	if len(msg.ID) > 0 && string(msg.ID) != "null" {
		if msg.ID[0] == '"' {
			_ = json.Unmarshal(msg.ID, &stringID)
		} else {
			_ = json.Unmarshal(msg.ID, &blockID)
		}
	}
	return msg, blockID, stringID
}

type Identity struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	FriendCode string `json:"friendCode"`
}

type FriendInfo struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	FriendCode string  `json:"friendCode"`
	Online     bool    `json:"online"`
	RoomCode   *string `json:"roomCode"`
}

type RemotePlayer struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Color string  `json:"color"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	Yaw   float64 `json:"yaw"`
}

type ChatLine struct {
	ID       string `json:"id"`
	FromID   string `json:"fromId"`
	FromName string `json:"fromName"`
	Text     string `json:"text"`
	Whisper  bool   `json:"whisper"`
	At       int64  `json:"at"`
}

type RoomInfo struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	HostName string `json:"hostName"`
	Players  int    `json:"players"`
}

type Player struct {
	id         string
	name       string
	friendCode string
	color      string
	x, y, z    float64
	yaw        float64
	roomID     string
	lastChat   time.Time
	lastPos    time.Time
	send       chan []byte
	conn       *websocket.Conn
}

type Room struct {
	id     string
	code   string
	name   string
	hostID string
	sizeX  int
	sizeY  int
	sizeZ  int
	blocks []int
}

type Hub struct {
	mu      sync.Mutex
	players map[string]*Player
	rooms   map[string]*Room
	byCode  map[string]string
	friends map[string]map[string]struct{}
	pending map[string]map[string]struct{}
}

func NewHub() *Hub {
	return &Hub{
		players: make(map[string]*Player),
		rooms:   make(map[string]*Room),
		byCode:  make(map[string]string),
		friends: make(map[string]map[string]struct{}),
		pending: make(map[string]map[string]struct{}),
	}
}

func (h *Hub) emit(p *Player, v any) {
	if p == nil {
		return
	}
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	select {
	case p.send <- b:
	default:
	}
}

func (h *Hub) Connect(hello inbound, stringID string, conn *websocket.Conn) *Player {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := stringID
	if !looksLikeID(id) {
		id = makeID()
	}
	name := sanitizeName(hello.Name)
	code := hello.FriendCode
	if !validFriendCode(code) {
		code = makeFriendCode()
	}

	p := h.players[id]
	if p == nil {
		p = &Player{
			id:         id,
			name:       name,
			friendCode: code,
			color:      colorForID(id),
			x:          12.5,
			y:          4,
			z:          12.5,
			send:       make(chan []byte, 64),
			conn:       conn,
		}
		h.players[id] = p
		h.byCode[code] = id
		if h.friends[id] == nil {
			h.friends[id] = make(map[string]struct{})
		}
	} else {
		p.name = name
		oldSend := p.send
		oldConn := p.conn
		p.send = make(chan []byte, 64)
		p.conn = conn
		if oldSend != nil {
			close(oldSend)
		}
		if oldConn != nil && oldConn != conn {
			_ = oldConn.Close()
		}
	}
	h.flushPending(p)
	h.emit(p, map[string]any{
		"type":  "helloOk",
		"you":   Identity{ID: p.id, Name: p.name, FriendCode: p.friendCode},
		"color": p.color,
	})
	h.pushFriends(p)
	h.pushRooms(p)
	return p
}

func (h *Hub) Disconnect(p *Player, conn *websocket.Conn) {
	if p == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if conn != nil && p.conn != conn {
		return
	}
	if p.roomID != "" {
		h.leaveRoomLocked(p)
	}
	delete(h.players, p.id)
	if p.send != nil {
		close(p.send)
		p.send = nil
	}
	p.conn = nil
	h.broadcastFriendsOfLocked(p.id)
}

func (h *Hub) Handle(p *Player, msg inbound, blockID int, stringID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	switch msg.Type {
	case "rename":
		p.name = sanitizeName(msg.Name)
		h.pushFriends(p)
		h.broadcastFriendsOfLocked(p.id)
		if p.roomID != "" {
			h.broadcastPlayers(p.roomID)
		}
	case "addFriend":
		h.addFriend(p, msg.Code)
	case "host":
		h.host(p, msg)
	case "join":
		h.join(p, msg.Code)
	case "leaveRoom":
		h.leaveRoomLocked(p)
	case "pos":
		h.pos(p, msg)
	case "place":
		h.place(p, msg, blockID)
	case "chat":
		h.chat(p, msg.Text, msg.ToID)
	case "invite":
		h.invite(p, msg.FriendID)
	}
}

func (h *Hub) addFriend(p *Player, raw string) {
	code := strings.ToUpper(strings.TrimSpace(raw))
	if code == p.friendCode {
		h.emit(p, map[string]any{"type": "error", "text": "To Twój własny kod."})
		return
	}
	otherID := h.byCode[code]
	other := h.players[otherID]
	if other != nil {
		h.linkFriends(p.id, other.id)
		h.pushFriends(p)
		h.pushFriends(other)
		h.emit(other, map[string]any{"type": "error", "text": p.name + " dodał Cię do znajomych!"})
		return
	}
	if h.pending[code] == nil {
		h.pending[code] = make(map[string]struct{})
	}
	h.pending[code][p.id] = struct{}{}
	h.emit(p, map[string]any{"type": "error", "text": "Kod zapisany. Znajomy pojawi się, gdy wejdzie do BOXEL."})
}

func (h *Hub) host(p *Player, msg inbound) {
	if p.roomID != "" {
		h.leaveRoomLocked(p)
	}
	sizeX := clampDim(msg.SizeX, 24)
	sizeY := clampDim(msg.SizeY, 16)
	sizeZ := clampDim(msg.SizeZ, 24)
	need := sizeX * sizeY * sizeZ
	blocks := append([]int(nil), msg.Blocks...)
	if len(blocks) != need {
		blocks = make([]int, need)
		copy(blocks, msg.Blocks)
	}
	room := &Room{
		id:     makeID(),
		code:   h.uniqueRoomCode(),
		name:   sanitizeName(firstNonEmpty(msg.Name, "Pokój")),
		hostID: p.id,
		sizeX:  sizeX,
		sizeY:  sizeY,
		sizeZ:  sizeZ,
		blocks: blocks,
	}
	h.rooms[room.id] = room
	p.roomID = room.id
	h.emit(p, map[string]any{
		"type":   "room",
		"code":   room.code,
		"name":   room.name,
		"host":   true,
		"sizeX":  room.sizeX,
		"sizeY":  room.sizeY,
		"sizeZ":  room.sizeZ,
		"blocks": room.blocks,
	})
	h.broadcastRooms()
	h.broadcastPlayers(room.id)
	h.broadcastFriendsOfLocked(p.id)
}

func (h *Hub) join(p *Player, code string) {
	var room *Room
	code = strings.TrimSpace(code)
	for _, r := range h.rooms {
		if r.code == code {
			room = r
			break
		}
	}
	if room == nil {
		h.emit(p, map[string]any{"type": "error", "text": "Nie ma takiego pokoju."})
		return
	}
	if len(h.roomPlayers(room.id)) >= 8 {
		h.emit(p, map[string]any{"type": "error", "text": "Pokój jest pełny."})
		return
	}
	if p.roomID != "" {
		h.leaveRoomLocked(p)
	}
	p.roomID = room.id
	h.emit(p, map[string]any{
		"type":   "room",
		"code":   room.code,
		"name":   room.name,
		"host":   room.hostID == p.id,
		"sizeX":  room.sizeX,
		"sizeY":  room.sizeY,
		"sizeZ":  room.sizeZ,
		"blocks": room.blocks,
	})
	h.broadcastPlayers(room.id)
	h.broadcastFriendsOfLocked(p.id)
}

func (h *Hub) leaveRoomLocked(p *Player) {
	roomID := p.roomID
	if roomID == "" {
		return
	}
	p.roomID = ""
	h.emit(p, map[string]any{"type": "leftRoom"})
	room := h.rooms[roomID]
	left := h.roomPlayers(roomID)
	if room != nil && len(left) == 0 {
		delete(h.rooms, roomID)
	} else if room != nil && room.hostID == p.id && len(left) > 0 {
		room.hostID = left[0].id
	}
	if h.rooms[roomID] != nil {
		h.broadcastPlayers(roomID)
	}
	h.broadcastRooms()
	h.broadcastFriendsOfLocked(p.id)
}

func (h *Hub) pos(p *Player, msg inbound) {
	now := time.Now()
	if now.Sub(p.lastPos) < 40*time.Millisecond {
		return
	}
	p.lastPos = now
	p.x, p.y, p.z, p.yaw = msg.X, msg.Y, msg.Z, msg.Yaw
	if p.roomID != "" {
		h.broadcastPlayers(p.roomID)
	}
}

func (h *Hub) place(p *Player, msg inbound, blockID int) {
	if p.roomID == "" {
		return
	}
	room := h.rooms[p.roomID]
	if room == nil {
		return
	}
	x, y, z := int(msg.X), int(msg.Y), int(msg.Z)
	if x < 0 || y < 0 || z < 0 || x >= room.sizeX || y >= room.sizeY || z >= room.sizeZ {
		return
	}
	idx := x + room.sizeX*(z+room.sizeZ*y)
	if idx < 0 || idx >= len(room.blocks) {
		return
	}
	block := blockID
	if block < 0 {
		block = 0
	}
	if block > 15 {
		block = 15
	}
	room.blocks[idx] = block
	for _, other := range h.roomPlayers(room.id) {
		if other.id == p.id {
			continue
		}
		h.emit(other, map[string]any{
			"type":   "place",
			"x":      x,
			"y":      y,
			"z":      z,
			"id":     block,
			"fromId": p.id,
		})
	}
}

func (h *Hub) chat(p *Player, raw, toID string) {
	now := time.Now()
	if !p.lastChat.IsZero() && now.Sub(p.lastChat) < 700*time.Millisecond {
		return
	}
	text, ok := sanitizeChat(raw)
	if !ok {
		h.emit(p, map[string]any{"type": "error", "text": "Tej wiadomości nie można wysłać."})
		return
	}
	line := ChatLine{
		ID: makeID(), FromID: p.id, FromName: p.name, Text: text, Whisper: toID != "", At: now.UnixMilli(),
	}
	if toID != "" {
		if !h.areFriends(p.id, toID) {
			h.emit(p, map[string]any{"type": "error", "text": "Szept tylko do znajomych."})
			return
		}
		p.lastChat = now
		h.emit(p, map[string]any{"type": "chat", "line": line})
		h.emit(h.players[toID], map[string]any{"type": "chat", "line": line})
		return
	}
	if p.roomID == "" {
		h.emit(p, map[string]any{"type": "error", "text": "Wejdź do pokoju, żeby pisać na czacie."})
		return
	}
	p.lastChat = now
	for _, other := range h.roomPlayers(p.roomID) {
		h.emit(other, map[string]any{"type": "chat", "line": line})
	}
}

func (h *Hub) invite(p *Player, friendID string) {
	if p.roomID == "" {
		h.emit(p, map[string]any{"type": "error", "text": "Najpierw załóż pokój."})
		return
	}
	if !h.areFriends(p.id, friendID) {
		h.emit(p, map[string]any{"type": "error", "text": "Możesz zapraszać tylko znajomych."})
		return
	}
	room := h.rooms[p.roomID]
	other := h.players[friendID]
	if room == nil || other == nil {
		h.emit(p, map[string]any{"type": "error", "text": "Znajomy jest offline."})
		return
	}
	h.emit(other, map[string]any{"type": "invite", "fromName": p.name, "roomCode": room.code})
}

func (h *Hub) roomPlayers(roomID string) []*Player {
	out := []*Player{}
	for _, p := range h.players {
		if p.roomID == roomID {
			out = append(out, p)
		}
	}
	return out
}

func (h *Hub) broadcastPlayers(roomID string) {
	members := h.roomPlayers(roomID)
	full := make([]RemotePlayer, 0, len(members))
	for _, p := range members {
		full = append(full, RemotePlayer{ID: p.id, Name: p.name, Color: p.color, X: p.x, Y: p.y, Z: p.z, Yaw: p.yaw})
	}
	for _, p := range members {
		list := make([]RemotePlayer, 0, len(full))
		for _, r := range full {
			if r.ID != p.id {
				list = append(list, r)
			}
		}
		h.emit(p, map[string]any{"type": "players", "players": list})
	}
}

func (h *Hub) friendList(p *Player) []FriendInfo {
	out := []FriendInfo{}
	for id := range h.friends[p.id] {
		other := h.players[id]
		info := FriendInfo{ID: id, Name: "Znajomy", FriendCode: "BX-??????", Online: other != nil}
		if other != nil {
			info.Name = other.name
			info.FriendCode = other.friendCode
			if other.roomID != "" {
				if room := h.rooms[other.roomID]; room != nil {
					c := room.code
					info.RoomCode = &c
				}
			}
		}
		out = append(out, info)
	}
	return out
}

func (h *Hub) pushFriends(p *Player) {
	h.emit(p, map[string]any{"type": "friends", "friends": h.friendList(p)})
}

func (h *Hub) roomList() []RoomInfo {
	out := []RoomInfo{}
	for _, r := range h.rooms {
		host := "?"
		if p := h.players[r.hostID]; p != nil {
			host = p.name
		}
		out = append(out, RoomInfo{Code: r.code, Name: r.name, HostName: host, Players: len(h.roomPlayers(r.id))})
	}
	return out
}

func (h *Hub) pushRooms(p *Player) {
	h.emit(p, map[string]any{"type": "rooms", "rooms": h.roomList()})
}

func (h *Hub) broadcastRooms() {
	rooms := h.roomList()
	for _, p := range h.players {
		h.emit(p, map[string]any{"type": "rooms", "rooms": rooms})
	}
}

func (h *Hub) broadcastFriendsOfLocked(id string) {
	if me := h.players[id]; me != nil {
		h.pushFriends(me)
	}
	for fid := range h.friends[id] {
		if f := h.players[fid]; f != nil {
			h.pushFriends(f)
		}
	}
}

func (h *Hub) linkFriends(a, b string) {
	if h.friends[a] == nil {
		h.friends[a] = make(map[string]struct{})
	}
	if h.friends[b] == nil {
		h.friends[b] = make(map[string]struct{})
	}
	h.friends[a][b] = struct{}{}
	h.friends[b][a] = struct{}{}
}

func (h *Hub) areFriends(a, b string) bool {
	_, ok := h.friends[a][b]
	return ok
}

func (h *Hub) flushPending(p *Player) {
	wait := h.pending[p.friendCode]
	if wait == nil {
		return
	}
	for id := range wait {
		other := h.players[id]
		if other == nil {
			continue
		}
		h.linkFriends(p.id, other.id)
		h.pushFriends(other)
	}
	delete(h.pending, p.friendCode)
	if len(wait) > 0 {
		h.pushFriends(p)
	}
}

func (h *Hub) uniqueRoomCode() string {
	for i := 0; i < 20; i++ {
		c := makeRoomCode()
		used := false
		for _, r := range h.rooms {
			if r.code == c {
				used = true
				break
			}
		}
		if !used {
			return c
		}
	}
	return makeRoomCode()
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func clampDim(n, fallback int) int {
	if n < 1 || n > 48 {
		return fallback
	}
	return n
}

func looksLikeID(id string) bool {
	if len(id) < 8 {
		return false
	}
	for _, r := range strings.ToLower(id) {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || r == '-' {
			continue
		}
		return false
	}
	return true
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func serveWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetReadLimit(2 << 20)

	var player *Player
	defer func() {
		if player != nil {
			hub.Disconnect(player, conn)
		}
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		msg, blockID, stringID := parseInbound(raw)
		if player == nil {
			if msg.Type != "hello" {
				continue
			}
			player = hub.Connect(msg, stringID, conn)
			go writePump(conn, player.send)
			continue
		}
		hub.Handle(player, msg, blockID, stringID)
	}
}

func writePump(conn *websocket.Conn, send <-chan []byte) {
	for buf := range send {
		_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, buf); err != nil {
			return
		}
	}
}

func main() {
	host := env("HOST", "0.0.0.0")
	port := env("PORT", "8787")
	hub := NewHub()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			serveWS(hub, w, r)
			return
		}
		w.Header().Set("content-type", "text/plain")
		_, _ = w.Write([]byte("BOXEL play server (Go)\n"))
	})
	addr := net.JoinHostPort(host, port)
	log.Printf("BOXEL Go server ws://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
