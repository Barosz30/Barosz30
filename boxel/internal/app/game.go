package app

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"boxel/internal/engine"
	"boxel/internal/hub"
	"boxel/internal/netclient"
	"boxel/internal/render"
	"boxel/internal/studio"
)

const (
	toolbarH = 58
	paletteW = 138
	socialW  = 250
	chatH    = 150
)

type fieldID int

const (
	fieldNone fieldID = iota
	fieldName
	fieldRename
	fieldFriend
	fieldJoin
	fieldChat
	fieldCodeText
)

type Game struct {
	rt   *studio.Runtime
	cam  render.Camera
	net  *netclient.Client
	w, h int

	social    bool
	help      bool
	connected bool
	profile   engine.Profile
	friends   []hub.FriendInfo
	rooms     []hub.RoomInfo
	roomCode  string
	roomHost  bool
	chat      []hub.ChatLine
	whisperTo string
	whisperNm string
	toast     string
	toastAt   time.Time

	focus    fieldID
	nameEdit string
	renameEd string
	friendEd string
	joinEd   string
	chatEd   string
	tmpl     int
	codeSel  int
	selPath  []int

	mouse        click
	prevMX       int
	prevMY       int
	dragging     bool
	downX, downY int
	rightHeld    bool
	pad          engine.InputState

	fpsN   int
	fpsAt  time.Time
	blinkT time.Time
}

func New() *Game {
	initFonts()
	w, scripts, name := engine.LoadTemplate("parkour")
	rt := studio.New(w, scripts, name)
	g := &Game{
		rt: rt, cam: render.NewOrbit(float64(w.SizeX)/2, 3, float64(w.SizeZ)/2),
		social: true, nameEdit: name, tmpl: 0,
		fpsAt: time.Now(), blinkT: time.Now(),
	}
	if p, ok := engine.LoadProfileFile(); ok {
		p.Name = hub.SanitizeName(p.Name)
		g.profile = p
	} else {
		g.profile = engine.Profile{ID: engine.NewID(), Name: "Gracz", FriendCode: hub.MakeFriendCode()}
		if g.profile.Name == "Gracz" {
			g.profile.Name = "Gracz" + fmt.Sprintf("%d", 10+int(time.Now().UnixNano()%89))
		}
		_ = engine.SaveProfile(g.profile)
	}
	g.renameEd = g.profile.Name
	ws := os.Getenv("BOXEL_WS")
	if ws == "" {
		ws = "ws://127.0.0.1:8787"
	}
	g.net = netclient.New(g.profile, ws)
	g.net.On(g.onNet)
	rt.OnEdit = func(x, y, z, id int) {
		g.net.Send(map[string]any{"type": "place", "x": x, "y": y, "z": z, "id": id})
	}
	rt.OnPos = func(x, y, z, yaw float64) {
		g.net.Send(map[string]any{"type": "pos", "x": x, "y": y, "z": z, "yaw": yaw, "score": g.rt.Player.Score})
	}
	rt.OnScriptEvent = func(kind string, x, y, z, block int) {
		g.net.Send(map[string]any{"type": "scriptEvent", "kind": kind, "x": x, "y": y, "z": z, "block": block})
	}
	g.net.Connect()
	return g
}

func (g *Game) Layout(ow, oh int) (int, int) {
	if ow < 960 {
		ow = 960
	}
	if oh < 640 {
		oh = 640
	}
	g.w, g.h = ow, oh
	return ow, oh
}

func (g *Game) Update() error {
	initFonts()
	g.readMouse()
	g.readKeys()
	dt := 1.0 / 60.0
	if tps := ebiten.ActualTPS(); tps > 1 {
		dt = 1.0 / tps
	}

	inUI := g.hitUI(g.mouse.x, g.mouse.y)
	vx, vy, vw, vh := g.viewport()

	if g.mouse.justDown && !inUI && g.mouse.x >= vx && g.mouse.x < vx+vw && g.mouse.y >= vy && g.mouse.y < vy+vh {
		g.downX, g.downY = g.mouse.x, g.mouse.y
		g.dragging = false
		if !g.mouse.right {
			g.focus = fieldNone
		}
	}
	if g.mouse.pressed && !inUI {
		if math.Hypot(float64(g.mouse.x-g.downX), float64(g.mouse.y-g.downY)) > 6 {
			g.dragging = true
		}
		if g.dragging {
			if g.rt.Mode == engine.ModePlay && !g.mouse.right {
				g.rt.Player.Yaw -= float64(g.mouse.dx) * 0.008
			} else if g.rt.Mode != engine.ModePlay {
				g.cam.OrbitDrag(float64(g.mouse.dx), float64(g.mouse.dy))
			}
		}
	}
	_, wy := ebiten.Wheel()
	if wy != 0 && !inUI && g.rt.Mode != engine.ModePlay {
		g.cam.Zoom(wy)
	}

	if g.mouse.justUp && !inUI && !g.dragging && g.rt.Mode == engine.ModeBuild {
		ox, oy, oz, dx, dy, dz := g.cam.ScreenRay(float64(g.mouse.x-vx), float64(g.mouse.y-vy), vw, vh)
		hit := engine.VoxelRaycast(g.rt.World, ox, oy, oz, dx, dy, dz, 40)
		g.rt.PlaceFromRay(hit.Hit, hit.Prev, g.mouse.right || g.rt.Brush == 0)
	}

	input := engine.InputState{
		Forward: ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) || g.pad.Forward,
		Back:    ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) || g.pad.Back,
		Left:    ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || g.pad.Left,
		Right:   ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) || g.pad.Right,
		Jump:    ebiten.IsKeyPressed(ebiten.KeySpace) || g.pad.Jump,
	}
	g.rt.Tick(dt, input)
	if g.rt.Mode == engine.ModePlay {
		g.cam.FollowPlay(g.rt.Player.X, g.rt.Player.Y, g.rt.Player.Z, g.rt.Player.Yaw, dt)
	} else {
		g.cam.SyncEye()
	}
	for _, s := range g.rt.DrainSounds() {
		playSound(s)
	}

	g.fpsN++
	if time.Since(g.fpsAt) > 500*time.Millisecond {
		g.rt.LastFps = int(float64(g.fpsN) / time.Since(g.fpsAt).Seconds())
		g.fpsN = 0
		g.fpsAt = time.Now()
	}
	if time.Since(g.toastAt) > 2600*time.Millisecond {
		g.toast = ""
	}
	g.typeIntoFocus()
	return nil
}

func (g *Game) readMouse() {
	x, y := ebiten.CursorPosition()
	justL := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	justR := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
	upL := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	upR := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight)
	g.mouse.dx, g.mouse.dy = x-g.prevMX, y-g.prevMY
	g.prevMX, g.prevMY = x, y
	g.mouse.x, g.mouse.y = x, y
	g.mouse.justDown = justL || justR
	g.mouse.justUp = upL || upR
	g.mouse.pressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	g.mouse.right = ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) || justR || (g.mouse.justUp && g.rightHeld)
	if justR {
		g.rightHeld = true
	}
	if upR {
		g.rightHeld = false
	}
	if upL || upR {
		// keep right flag for the justUp place action
		if upR {
			g.mouse.right = true
		}
		if upL && !upR {
			g.mouse.right = false
		}
	}
	if !g.mouse.pressed {
		g.dragging = false
	}
}

func (g *Game) readKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) && (ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMetaLeft)) {
		g.rt.UndoLast()
	}
	for k := ebiten.KeyDigit1; k <= ebiten.KeyDigit9; k++ {
		if inpututil.IsKeyJustPressed(k) && g.focus == fieldNone {
			g.rt.Brush = int(k - ebiten.KeyDigit0)
		}
	}
	if (inpututil.IsKeyJustPressed(ebiten.KeyDigit0) || inpututil.IsKeyJustPressed(ebiten.KeyX)) && g.focus == fieldNone {
		g.rt.Brush = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
		g.commitField()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.help = false
		g.focus = fieldNone
	}
}

func (g *Game) typeIntoFocus() {
	if g.focus == fieldNone {
		return
	}
	dst := g.fieldPtr()
	if dst == nil {
		return
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(*dst) > 0 {
		r := []rune(*dst)
		*dst = string(r[:len(r)-1])
		if g.focus == fieldCodeText {
			g.afterEdit()
		}
	}
	chars := ebiten.AppendInputChars(nil)
	for _, ch := range chars {
		if unicode.IsControl(ch) {
			continue
		}
		*dst += string(ch)
		if g.focus == fieldChat && len([]rune(*dst)) > 60 {
			*dst = string([]rune(*dst)[:60])
		}
		if (g.focus == fieldName || g.focus == fieldRename) && len([]rune(*dst)) > 16 {
			*dst = string([]rune(*dst)[:16])
		}
		if g.focus == fieldCodeText && len([]rune(*dst)) > 48 {
			*dst = string([]rune(*dst)[:48])
		}
	}
	if g.focus == fieldCodeText && len(chars) > 0 {
		g.afterEdit()
	}
}

func (g *Game) fieldPtr() *string {
	switch g.focus {
	case fieldName:
		return &g.nameEdit
	case fieldRename:
		return &g.renameEd
	case fieldFriend:
		return &g.friendEd
	case fieldJoin:
		return &g.joinEd
	case fieldChat:
		return &g.chatEd
	case fieldCodeText:
		if n := g.selectedNode(); n != nil && n.Kind == engine.SetMessage {
			return &n.Text
		}
	}
	return nil
}

func (g *Game) commitField() {
	switch g.focus {
	case fieldName:
		g.rt.Name = g.nameEdit
	case fieldRename:
		name := hub.SanitizeName(g.renameEd)
		g.profile.Name = name
		g.renameEd = name
		_ = engine.SaveProfile(g.profile)
		g.net.Send(map[string]any{"type": "rename", "name": name})
	case fieldFriend:
		if g.friendEd != "" {
			g.net.Send(map[string]any{"type": "addFriend", "code": g.friendEd})
			g.friendEd = ""
		}
	case fieldJoin:
		if g.joinEd != "" {
			g.net.Send(map[string]any{"type": "join", "code": g.joinEd})
			g.joinEd = ""
		}
	case fieldChat:
		if strings.TrimSpace(g.chatEd) != "" {
			msg := map[string]any{"type": "chat", "text": g.chatEd}
			if g.whisperTo != "" {
				msg["toId"] = g.whisperTo
			}
			g.net.Send(msg)
			g.chatEd = ""
		}
	case fieldCodeText:
		g.afterEdit()
	}
	g.focus = fieldNone
}

func (g *Game) bottomH() int {
	if g.rt.Mode != engine.ModeCode {
		return chatH
	}
	h := g.h * 2 / 5
	if h < 260 {
		h = 260
	}
	if h > 440 {
		h = 440
	}
	return h
}

func (g *Game) viewport() (x, y, w, h int) {
	x, y = 0, toolbarH
	w, h = g.w, g.h-toolbarH-g.bottomH()
	if g.rt.Mode == engine.ModeBuild {
		x += paletteW
		w -= paletteW
	}
	if g.social {
		w -= socialW
	}
	if w < 80 {
		w = 80
	}
	if h < 80 {
		h = 80
	}
	return
}

func (g *Game) hitUI(x, y int) bool {
	vx, vy, vw, vh := g.viewport()
	if x >= vx && y >= vy && x < vx+vw && y < vy+vh && !g.help {
		return false
	}
	return true
}

func (g *Game) flash(s string) {
	g.toast = s
	g.toastAt = time.Now()
}

func (g *Game) onNet(msg map[string]any) {
	switch str(msg["type"]) {
	case "helloOk":
		g.connected = true
		if you, ok := msg["you"].(map[string]any); ok {
			g.profile.ID = str(you["id"])
			g.profile.Name = str(you["name"])
			g.profile.FriendCode = str(you["friendCode"])
			g.renameEd = g.profile.Name
		}
	case "friends":
		g.friends = parseFriends(msg["friends"])
	case "rooms":
		g.rooms = parseRooms(msg["rooms"])
	case "room":
		code := str(msg["code"])
		first := g.roomCode != code
		g.roomCode = code
		g.roomHost = asBool(msg["host"])
		if g.rt.Mode == engine.ModePlay {
			g.rt.Stop(engine.ModeBuild)
		}
		g.rt.ApplyWorldData(asInt(msg["sizeX"]), asInt(msg["sizeY"]), asInt(msg["sizeZ"]), asIntSlice(msg["blocks"]))
		g.rt.SetScripts(parseScripts(msg["scripts"]))
		if first {
			g.flash("Pokój " + g.roomCode)
		}
	case "scripts":
		if str(msg["fromId"]) != g.profile.ID {
			g.rt.SetScripts(parseScripts(msg["scripts"]))
		}
		if _, ok := msg["host"]; ok {
			g.roomHost = asBool(msg["host"])
		}
	case "scriptFx":
		if str(msg["fromId"]) == g.profile.ID {
			break
		}
		name := str(msg["fromName"])
		if m := str(msg["message"]); m != "" {
			g.flash(name + ": " + m)
		} else if asBool(msg["win"]) {
			g.flash(name + " wygrał!")
		} else if asBool(msg["lose"]) {
			g.flash(name + " odpadł")
		}
	case "hostChanged":
		g.roomHost = asBool(msg["host"])
		if g.roomHost {
			g.flash("Jesteś gospodarzem klocków kodu")
		}
	case "players":
		g.rt.Remotes = parseRemotes(msg["players"])
	case "place":
		g.rt.ApplyNetworkEdit(asInt(msg["x"]), asInt(msg["y"]), asInt(msg["z"]), asInt(msg["id"]))
	case "chat":
		if line, ok := msg["line"].(map[string]any); ok {
			g.chat = append(g.chat, hub.ChatLine{
				ID: str(line["id"]), FromID: str(line["fromId"]), FromName: str(line["fromName"]),
				Text: str(line["text"]), Whisper: asBool(line["whisper"]),
			})
			if len(g.chat) > 40 {
				g.chat = g.chat[len(g.chat)-40:]
			}
		}
	case "invite":
		g.flash(str(msg["fromName"]) + " zaprasza do pokoju " + str(msg["roomCode"]))
	case "leftRoom":
		g.roomCode = ""
		g.roomHost = false
		g.rt.Remotes = nil
	case "error":
		g.flash(str(msg["text"]))
	}
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func asInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case string:
		i, _ := strconv.Atoi(n)
		return i
	}
	return 0
}

func asBool(v any) bool {
	b, _ := v.(bool)
	return b
}

func asIntSlice(v any) []int {
	arr, _ := v.([]any)
	out := make([]int, 0, len(arr))
	for _, x := range arr {
		out = append(out, asInt(x))
	}
	return out
}

func parseFriends(v any) []hub.FriendInfo {
	arr, _ := v.([]any)
	var out []hub.FriendInfo
	for _, x := range arr {
		m, _ := x.(map[string]any)
		var rc *string
		if s := str(m["roomCode"]); s != "" {
			rc = &s
		}
		out = append(out, hub.FriendInfo{
			ID: str(m["id"]), Name: str(m["name"]), FriendCode: str(m["friendCode"]),
			Online: asBool(m["online"]), RoomCode: rc,
		})
	}
	return out
}

func parseRooms(v any) []hub.RoomInfo {
	arr, _ := v.([]any)
	var out []hub.RoomInfo
	for _, x := range arr {
		m, _ := x.(map[string]any)
		out = append(out, hub.RoomInfo{Code: str(m["code"]), Name: str(m["name"]), HostName: str(m["hostName"]), Players: asInt(m["players"])})
	}
	return out
}

func parseRemotes(v any) []engine.RemoteAvatar {
	arr, _ := v.([]any)
	var out []engine.RemoteAvatar
	for _, x := range arr {
		m, _ := x.(map[string]any)
		out = append(out, engine.RemoteAvatar{
			ID: str(m["id"]), Name: str(m["name"]), Color: str(m["color"]),
			X: num(m["x"]), Y: num(m["y"]), Z: num(m["z"]), Yaw: num(m["yaw"]),
			Score: asInt(m["score"]),
		})
	}
	return out
}

func parseScripts(v any) []engine.ScriptStack {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var out []engine.ScriptStack
	if json.Unmarshal(b, &out) != nil {
		return nil
	}
	return engine.SanitizeScripts(out)
}

func num(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	}
	return 0
}
