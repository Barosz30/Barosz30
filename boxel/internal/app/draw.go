package app

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"boxel/internal/engine"
	"boxel/internal/hub"
	"boxel/internal/render"
)

func (g *Game) Draw(screen *ebiten.Image) {
	initFonts()
	screen.Fill(colBg)
	g.drawToolbar(screen)
	if g.rt.Mode == engine.ModeBuild {
		g.drawPalette(screen)
	}
	vx, vy, vw, vh := g.viewport()
	var ghost *[3]int
	if g.rt.Mode == engine.ModeBuild {
		ox, oy, oz, dx, dy, dz := g.cam.ScreenRay(float64(g.mouse.x-vx), float64(g.mouse.y-vy), vw, vh)
		hit := engine.VoxelRaycast(g.rt.World, ox, oy, oz, dx, dy, dz, 40)
		if g.rt.Brush == 0 {
			ghost = hit.Hit
		} else {
			ghost = hit.Prev
		}
	}
	blink := g.rt.Player.HitCooldown > 0 && time.Now().UnixMilli()/80%2 == 0
	render.DrawWorld(screen, vx, vy, vw, vh, g.cam, g.rt.World, g.rt.Player, g.rt.Remotes, ghost, g.rt.Brush, g.rt.Mode, g.rt.CoinSpin, blink)
	g.drawHUD(screen, vx, vy, vw, vh)
	if g.social {
		g.drawSocial(screen)
	}
	if g.rt.Mode == engine.ModeCode {
		g.drawCode(screen)
	} else {
		g.drawChat(screen)
	}
	if g.toast != "" {
		r := Rect{g.w/2 - 160, 70, 320, 36}
		panel(screen, r, colAccent)
		tw, th := text.Measure(g.toast, face12, 0)
		drawText(screen, g.toast, r.X+(r.W-int(tw))/2, r.Y+(r.H-int(th))/2, face12, colBg)
	}
	if g.help {
		g.drawHelp(screen)
	}
}

func (g *Game) drawToolbar(dst *ebiten.Image) {
	fillRect(dst, Rect{0, 0, g.w, toolbarH}, color.RGBA{0x2a, 0x1d, 0x42, 0xff})
	fillRect(dst, Rect{0, toolbarH - 3, g.w, 3}, colLine)
	fillRect(dst, Rect{10, 14, 28, 28}, color.RGBA{0x5d, 0xad, 0x3e, 0xff})
	strokeRect(dst, Rect{10, 14, 28, 28}, colLine, 3)
	drawText(dst, "BOXEL", 46, 8, face18, colAccent)
	sub := "Aplikacja klocków"
	if g.rt.LastFps > 0 {
		sub = fmt.Sprintf("Aplikacja klocków · %d fps", g.rt.LastFps)
	}
	drawText(dst, sub, 46, 32, face10, colMuted)

	x := 250
	for _, spec := range []struct {
		label string
		mode  engine.StudioMode
		play  bool
	}{
		{"Buduj", engine.ModeBuild, false},
		{"Klocki", engine.ModeCode, false},
		{"Graj", engine.ModePlay, true},
	} {
		label := spec.label
		on := g.rt.Mode == spec.mode
		fill := colBtn
		if spec.play {
			fill = color.RGBA{0x2e, 0x8a, 0x52, 0xff}
			if g.rt.Mode == engine.ModePlay {
				label = "Stop"
				on = true
				fill = colDanger
			}
		}
		r := Rect{x, 12, 78, 34}
		if button(dst, r, label, on && !spec.play, fill, g.mouse) {
			if spec.play {
				if g.rt.Mode == engine.ModePlay {
					g.rt.SetMode(engine.ModeBuild)
				} else {
					g.rt.SetScripts(g.rt.Scripts)
					g.rt.SetMode(engine.ModePlay)
				}
			} else {
				g.rt.SetMode(spec.mode)
			}
		}
		x += 84
	}

	drawText(dst, "GRA", x+4, 6, face10, colMuted)
	fr := Rect{x, 22, 150, 26}
	g.inputBox(dst, fr, g.nameEdit, fieldName)

	x += 160
	drawText(dst, "SZABLON", x+4, 6, face10, colMuted)
	tr := Rect{x, 22, 168, 26}
	tname := engine.Templates[g.tmpl].Name
	if button(dst, tr, tname, false, colBg, g.mouse) {
		g.tmpl = (g.tmpl + 1) % len(engine.Templates)
		w, s, n := engine.LoadTemplate(engine.Templates[g.tmpl].ID)
		g.rt.Stop(engine.ModeBuild)
		g.rt.LoadWorld(w)
		g.rt.SetScripts(s)
		g.rt.Name = n
		g.nameEdit = n
		g.pushWorld()
		g.flash("Wczytano szablon")
	}

	x = g.w - 8
	btns := []struct {
		label string
		fn    func()
	}{
		{"?", func() { g.help = !g.help }},
		{"Znajomi", func() { g.social = !g.social }},
		{"Plik", func() {
			path := engine.LastExportPath(engine.SerializeGame(g.rt.Name, g.rt.World, g.rt.Scripts))
			_ = engine.ExportGame(engine.SerializeGame(g.rt.Name, g.rt.World, g.rt.Scripts), path)
			g.flash("Zapisano " + path)
		}},
		{"Wczytaj", func() {
			if data, ok := engine.LoadGameLocal(); ok {
				w, s, n := engine.DeserializeGame(data)
				g.rt.Stop(engine.ModeBuild)
				g.rt.LoadWorld(w)
				g.rt.SetScripts(s)
				g.rt.Name, g.nameEdit = n, n
				g.pushWorld()
				g.flash("Wczytano zapis")
			} else {
				g.flash("Brak zapisu")
			}
		}},
		{"Zapisz", func() {
			_ = engine.SaveGameLocal(engine.SerializeGame(g.rt.Name, g.rt.World, g.rt.Scripts))
			g.flash("Zapisano na tym komputerze")
		}},
		{"Cofnij", func() { g.rt.UndoLast() }},
	}
	for i := range btns {
		b := btns[i]
		w := 70
		if b.label == "?" {
			w = 32
		}
		if b.label == "Znajomi" {
			w = 84
		}
		x -= w + 6
		r := Rect{x, 12, w, 34}
		on := b.label == "Znajomi" && g.social
		if button(dst, r, b.label, on, colPanel, g.mouse) {
			b.fn()
		}
	}
	if g.roomCode != "" {
		drawText(dst, "Pokój "+g.roomCode, x-110, 20, face12, colPlay)
	}
}

func (g *Game) inputBox(dst *ebiten.Image, r Rect, val string, id fieldID) {
	fill := colBg
	if g.focus == id {
		fill = color.RGBA{0x3a, 0x2a, 0x58, 0xff}
	}
	panel(dst, r, fill)
	txt := val
	if g.focus == id && time.Now().UnixMilli()/400%2 == 0 {
		txt += "|"
	}
	drawText(dst, txt, r.X+6, r.Y+5, face12, colInk)
	if g.mouse.justDown && r.Contains(g.mouse.x, g.mouse.y) {
		g.focus = id
	}
}

func (g *Game) drawPalette(dst *ebiten.Image) {
	fillRect(dst, Rect{0, toolbarH, paletteW, g.h - toolbarH}, color.RGBA{0x24, 0x18, 0x3a, 0xff})
	drawText(dst, "Klocki", 12, toolbarH+10, face14, colAccent)
	y := toolbarH + 36
	for _, b := range engine.Hotbar() {
		r := Rect{10, y, paletteW - 20, 28}
		fill := color.RGBA{b.Color.R, b.Color.G, b.Color.B, 255}
		if button(dst, r, b.Name, g.rt.Brush == int(b.ID), fill, g.mouse) {
			g.rt.Brush = int(b.ID)
		}
		y += 32
	}
	drawText(dst, "LPM stawia", 10, y+8, face10, colMuted)
	drawText(dst, "PPM kasuje", 10, y+22, face10, colMuted)
	drawText(dst, "Ctrl+Z cofnij", 10, y+36, face10, colMuted)
}

func (g *Game) drawHUD(dst *ebiten.Image, vx, vy, vw, vh int) {
	hud := g.rt.HUD()
	drawText(dst, fmt.Sprintf("PKT %d", hud.Score), vx+12, vy+10, face14, colInk)
	drawText(dst, fmt.Sprintf("MONETY %d", hud.Coins), vx+110, vy+10, face14, colAccent)
	drawText(dst, fmt.Sprintf("ŻYCIA %d", hud.Lives), vx+240, vy+10, face14, colDanger)
	rx := vx + 360
	for _, rem := range g.rt.Remotes {
		drawText(dst, fmt.Sprintf("%s %d", rem.Name, rem.Score), rx, vy+10, face10, colMuted)
		rx += 90
	}
	if hud.Message != "" {
		tw, _ := text.Measure(hud.Message, face14, 0)
		drawText(dst, hud.Message, vx+(vw-int(tw))/2, vy+40, face14, colAccent)
	}
	if g.rt.Mode == engine.ModePlay {
		pads := []struct {
			label string
			r     Rect
			set   func(bool)
		}{
			{"▲", Rect{vx + 48, vy + vh - 130, 44, 40}, func(v bool) { g.pad.Forward = v }},
			{"◀", Rect{vx + 8, vy + vh - 86, 44, 40}, func(v bool) { g.pad.Left = v }},
			{"▶", Rect{vx + 88, vy + vh - 86, 44, 40}, func(v bool) { g.pad.Right = v }},
			{"▼", Rect{vx + 48, vy + vh - 42, 44, 40}, func(v bool) { g.pad.Back = v }},
		}
		held := g.mouse.pressed
		g.pad = engine.InputState{}
		for _, p := range pads {
			on := held && p.r.Contains(g.mouse.x, g.mouse.y)
			p.set(on)
			button(dst, p.r, p.label, on, colPanel, g.mouse)
		}
		jr := Rect{vx + vw - 110, vy + vh - 90, 96, 64}
		if button(dst, jr, "SKOK", false, colPlay, g.mouse) || (held && jr.Contains(g.mouse.x, g.mouse.y) && g.mouse.justDown) {
			g.rt.JumpQueued = true
			g.pad.Jump = true
		}
		if held && jr.Contains(g.mouse.x, g.mouse.y) {
			g.pad.Jump = true
		}
	}
	if hud.Phase == "won" || hud.Phase == "lost" {
		title := "WYGRANA!"
		body := "Super, dobiegłeś do mety!"
		if hud.Phase == "lost" {
			title = "PRZEGRANA"
			body = "Spróbuj jeszcze raz albo ułatw parkour."
		}
		r := Rect{vx + vw/2 - 160, vy + vh/2 - 80, 320, 160}
		panel(dst, r, colPanel)
		drawText(dst, title, r.X+24, r.Y+16, face18, colAccent)
		drawText(dst, body, r.X+24, r.Y+50, face12, colInk)
		drawText(dst, fmt.Sprintf("Wynik: %d", hud.Score), r.X+24, r.Y+74, face12, colMuted)
		if button(dst, Rect{r.X + 24, r.Y + 108, 120, 32}, "Jeszcze raz", false, colPlay, g.mouse) {
			g.rt.Play()
		}
		if button(dst, Rect{r.X + 156, r.Y + 108, 140, 32}, "Do studia", false, colBtn, g.mouse) {
			g.rt.Stop(engine.ModeBuild)
		}
	}
}

func (g *Game) drawSocial(dst *ebiten.Image) {
	x := g.w - socialW
	y0 := toolbarH
	h := g.h - toolbarH
	h -= g.bottomH()
	fillRect(dst, Rect{x, y0, socialW, h}, color.RGBA{0x24, 0x18, 0x3a, 0xff})
	drawText(dst, "Znajomi", x+12, y0+8, face14, colAccent)
	net := "Łączenie…"
	nc := colDanger
	if g.connected {
		net = "W aplikacji / online"
		nc = colPlay
	}
	drawText(dst, net, x+12, y0+30, face10, nc)
	drawText(dst, "Imię", x+12, y0+50, face10, colMuted)
	g.inputBox(dst, Rect{x + 12, y0 + 66, socialW - 24, 26}, g.renameEd, fieldRename)
	drawText(dst, "Twój kod  "+g.profile.FriendCode, x+12, y0+98, face12, colAccent)

	g.inputBox(dst, Rect{x + 12, y0 + 122, socialW - 80, 26}, g.friendEd, fieldFriend)
	if button(dst, Rect{x + socialW - 64, y0 + 122, 52, 26}, "Dodaj", false, colBtn, g.mouse) {
		g.focus = fieldFriend
		g.commitField()
	}
	y := y0 + 156
	if len(g.friends) == 0 {
		drawText(dst, "Dodaj znajomego kodem.", x+12, y, face10, colMuted)
		y += 22
	}
	for _, f := range g.friends {
		st := "offline"
		if f.Online {
			st = "online"
		}
		if f.RoomCode != nil {
			st += " · pokój " + *f.RoomCode
		}
		drawText(dst, f.Name, x+12, y, face12, colInk)
		drawText(dst, st, x+12, y+16, face10, colMuted)
		fid := f.ID
		if button(dst, Rect{x + socialW - 118, y, 48, 22}, "Czat", false, colBtn, g.mouse) && f.Online {
			g.whisperTo = fid
			g.whisperNm = f.Name
		}
		if button(dst, Rect{x + socialW - 64, y, 52, 22}, "Zaproś", false, colBtn, g.mouse) && f.Online && g.roomCode != "" {
			g.net.Send(map[string]any{"type": "invite", "friendId": fid})
		}
		y += 40
	}

	drawText(dst, "Wspólny świat", x+12, y+8, face14, colAccent)
	y += 32
	if g.roomCode != "" {
		drawText(dst, "Pokój "+g.roomCode, x+12, y, face12, colPlay)
		if button(dst, Rect{x + socialW - 70, y - 4, 58, 24}, "Wyjdź", false, colDanger, g.mouse) {
			g.net.Send(map[string]any{"type": "leaveRoom"})
		}
		y += 28
	} else {
		if button(dst, Rect{x + 12, y, socialW - 24, 28}, "Załóż pokój", false, colPlay, g.mouse) {
			g.net.Send(map[string]any{
				"type": "host", "name": g.rt.Name,
				"sizeX": g.rt.World.SizeX, "sizeY": g.rt.World.SizeY, "sizeZ": g.rt.World.SizeZ,
				"blocks":  bytesToInt(g.rt.World.Blocks),
				"scripts": g.rt.Scripts,
			})
		}
		y += 36
		g.inputBox(dst, Rect{x + 12, y, socialW - 84, 26}, g.joinEd, fieldJoin)
		if button(dst, Rect{x + socialW - 68, y, 56, 26}, "Dołącz", false, colBtn, g.mouse) {
			g.focus = fieldJoin
			g.commitField()
		}
		y += 32
	}
	for _, r := range g.rooms {
		drawText(dst, fmt.Sprintf("%s · %s · %d", r.Name, r.Code, r.Players), x+12, y, face10, colMuted)
		y += 16
	}
}

func bytesToInt(b []uint8) []int {
	out := make([]int, len(b))
	for i, v := range b {
		out[i] = int(v)
	}
	return out
}

func (g *Game) drawChat(dst *ebiten.Image) {
	y := g.h - chatH
	fillRect(dst, Rect{0, y, g.w, chatH}, color.RGBA{0x24, 0x18, 0x3a, 0xff})
	ly := y + 8
	start := 0
	if len(g.chat) > 5 {
		start = len(g.chat) - 5
	}
	for _, line := range g.chat[start:] {
		prefix := line.FromName + ": "
		if line.Whisper {
			prefix = line.FromName + " (szept): "
		}
		drawText(dst, prefix+line.Text, 12, ly, face12, colInk)
		ly += 16
	}
	px := 8
	for _, p := range hub.ChatPresets {
		w := 78
		if button(dst, Rect{px, g.h - 64, w, 22}, p, false, colBtn, g.mouse) {
			msg := map[string]any{"type": "chat", "text": p}
			if g.whisperTo != "" {
				msg["toId"] = g.whisperTo
			}
			g.net.Send(msg)
		}
		px += w + 6
	}
	ph := "Napisz na czacie pokoju…"
	if g.whisperTo != "" {
		ph = "Do " + g.whisperNm + "…"
	}
	shown := g.chatEd
	if shown == "" && g.focus != fieldChat {
		shown = ph
	}
	g.inputBox(dst, Rect{8, g.h - 36, g.w - 180, 26}, shown, fieldChat)
	if g.focus != fieldChat && g.chatEd == "" {
		// placeholder painted as value; don't keep it
	}
	if button(dst, Rect{g.w - 168, g.h - 36, 72, 26}, "Wyślij", false, colPlay, g.mouse) {
		g.focus = fieldChat
		g.commitField()
	}
	if g.whisperTo != "" {
		if button(dst, Rect{g.w - 90, g.h - 36, 82, 26}, "Anuluj", false, colDanger, g.mouse) {
			g.whisperTo, g.whisperNm = "", ""
		}
	}
}
