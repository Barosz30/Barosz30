package app

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"boxel/internal/engine"
)

func brickColor(kind engine.CodeKind) color.RGBA {
	switch engine.NodeCategory(kind) {
	case "event":
		return color.RGBA{0xe8, 0xb0, 0x3a, 0xff}
	case "looks":
		return color.RGBA{0x9b, 0x6b, 0xd6, 0xff}
	case "sound":
		return color.RGBA{0xd0, 0x5a, 0xb8, 0xff}
	case "motion":
		return color.RGBA{0x4a, 0xa0, 0xe8, 0xff}
	case "sensing":
		return color.RGBA{0x4e, 0xc4, 0xc8, 0xff}
	default:
		return color.RGBA{0xd0, 0x7a, 0x3f, 0xff}
	}
}

func (g *Game) canEditCode() bool {
	return g.roomCode == "" || g.roomHost
}

func (g *Game) afterEdit() {
	g.rt.SetScripts(g.rt.Scripts)
	g.pushScripts()
}

func (g *Game) pushScripts() {
	if !g.roomHost || g.roomCode == "" {
		return
	}
	g.net.Send(map[string]any{"type": "scripts", "scripts": g.rt.Scripts})
}

func (g *Game) pushWorld() {
	if !g.roomHost || g.roomCode == "" {
		return
	}
	g.net.Send(map[string]any{
		"type":    "world",
		"sizeX":   g.rt.World.SizeX,
		"sizeY":   g.rt.World.SizeY,
		"sizeZ":   g.rt.World.SizeZ,
		"blocks":  bytesToInt(g.rt.World.Blocks),
		"scripts": g.rt.Scripts,
	})
}

func (g *Game) selectedNode() *engine.CodeNode {
	list, idx := g.selectedList()
	if list == nil || idx < 0 || idx >= len(*list) {
		return nil
	}
	return &(*list)[idx]
}

func (g *Game) selectedList() (*[]engine.CodeNode, int) {
	if g.codeSel < 0 || g.codeSel >= len(g.rt.Scripts) {
		return nil, -1
	}
	cur := &g.rt.Scripts[g.codeSel].Blocks
	if len(g.selPath) == 0 {
		return cur, -1
	}
	for i := 0; i < len(g.selPath)-1; i++ {
		idx := g.selPath[i]
		if idx < 0 || idx >= len(*cur) {
			return cur, -1
		}
		cur = &(*cur)[idx].Body
	}
	return cur, g.selPath[len(g.selPath)-1]
}

func pathEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (g *Game) addCode(kind engine.CodeKind) {
	if !g.canEditCode() {
		g.flash("Tylko gospodarz zmienia klocki w pokoju")
		return
	}
	n := engine.NewFromToolbox(kind)
	if engine.IsHat(kind) {
		g.rt.Scripts = append(g.rt.Scripts, engine.ScriptStack{ID: engine.NewID(), Blocks: []engine.CodeNode{n}})
		g.codeSel = len(g.rt.Scripts) - 1
		g.selPath = []int{0}
		g.afterEdit()
		return
	}
	if len(g.rt.Scripts) == 0 {
		g.flash("Najpierw kliknij żółte zdarzenie (np. gdy gra się zaczyna)")
		return
	}
	if g.codeSel < 0 || g.codeSel >= len(g.rt.Scripts) {
		g.codeSel = len(g.rt.Scripts) - 1
		g.selPath = nil
	}
	sel := g.selectedNode()
	if sel != nil && (sel.Kind == engine.IfScore || sel.Kind == engine.Repeat) {
		sel.Body = append(sel.Body, n)
		g.selPath = append(append([]int{}, g.selPath...), len(sel.Body)-1)
		g.afterEdit()
		return
	}
	list, idx := g.selectedList()
	if list == nil {
		st := &g.rt.Scripts[g.codeSel]
		st.Blocks = append(st.Blocks, n)
		g.selPath = []int{len(st.Blocks) - 1}
		g.afterEdit()
		return
	}
	insertAt := len(*list)
	if idx >= 0 {
		insertAt = idx + 1
	}
	*list = append(*list, engine.CodeNode{})
	copy((*list)[insertAt+1:], (*list)[insertAt:])
	(*list)[insertAt] = n
	if len(g.selPath) == 0 {
		g.selPath = []int{insertAt}
	} else {
		g.selPath = append(append([]int{}, g.selPath[:len(g.selPath)-1]...), insertAt)
	}
	g.afterEdit()
}

func (g *Game) deleteStack(idx int) {
	if !g.canEditCode() || idx < 0 || idx >= len(g.rt.Scripts) {
		return
	}
	g.rt.Scripts = append(g.rt.Scripts[:idx], g.rt.Scripts[idx+1:]...)
	if g.codeSel >= len(g.rt.Scripts) {
		g.codeSel = len(g.rt.Scripts) - 1
	}
	g.selPath = nil
	g.afterEdit()
}

func (g *Game) deleteSelectedNode() {
	if !g.canEditCode() {
		return
	}
	list, idx := g.selectedList()
	if list == nil || idx < 0 || idx >= len(*list) {
		return
	}
	if len(g.selPath) == 1 && idx == 0 && engine.IsHat((*list)[idx].Kind) {
		g.deleteStack(g.codeSel)
		return
	}
	*list = append((*list)[:idx], (*list)[idx+1:]...)
	if idx >= len(*list) {
		idx = len(*list) - 1
	}
	if idx < 0 {
		g.selPath = g.selPath[:len(g.selPath)-1]
	} else {
		g.selPath[len(g.selPath)-1] = idx
	}
	g.afterEdit()
}

func (g *Game) adjustSelected(dir int) {
	if !g.canEditCode() {
		return
	}
	n := g.selectedNode()
	if n == nil || !engine.CanAdjust(n.Kind) {
		return
	}
	engine.AdjustNode(n, dir)
	g.afterEdit()
}

func (g *Game) drawCode(dst *ebiten.Image) {
	h := g.bottomH()
	y0 := g.h - h
	fillRect(dst, Rect{0, y0, g.w, h}, color.RGBA{0x1e, 0x14, 0x32, 0xff})
	hint := "Klocki kodu — żółte zdarzenie, potem akcja. + / − zmienia liczby."
	if g.roomCode != "" && g.roomHost {
		hint = "Gospodarz: klocki lecą do całego pokoju. Zapis lokalny: przycisk Zapisz."
	} else if g.roomCode != "" && !g.roomHost {
		hint = "Oglądasz klocki gospodarza. Sam możesz je zmienić, gdy założysz własny pokój."
	}
	drawText(dst, hint, 12, y0+6, face10, colMuted)

	x := 8
	yy := y0 + 24
	for _, item := range engine.Toolbox {
		fill := brickColor(item.Kind)
		if !g.canEditCode() {
			fill = color.RGBA{0x5a, 0x4a, 0x70, 0xff}
		}
		r := Rect{x, yy, 148, 22}
		it := item
		if button(dst, r, it.Label, false, fill, g.mouse) {
			g.addCode(it.Kind)
		}
		x += 154
		if x > g.w-160 {
			x = 8
			yy += 26
		}
	}

	yy += 30
	if len(g.rt.Scripts) == 0 {
		drawText(dst, "Tu będą Twoje skrypty. Kliknij „gdy gra się zaczyna”, potem „ustaw wiadomość”.", 12, yy+8, face12, colInk)
		drawText(dst, "W GRAJ klocki działają na tym komputerze. W pokoju — u wszystkich graczy.", 12, yy+28, face12, colMuted)
		return
	}

	left := 8
	for i, st := range g.rt.Scripts {
		if yy > g.h-28 {
			break
		}
		label := "Skrypt"
		if len(st.Blocks) > 0 {
			label = engine.NodeLabel(st.Blocks[0])
		}
		hdr := Rect{left, yy, g.w - 90, 20}
		on := i == g.codeSel
		col := colPanel
		if on {
			col = color.RGBA{0x4a, 0x3a, 0x68, 0xff}
		}
		if button(dst, hdr, label, on, col, g.mouse) {
			g.codeSel = i
			g.selPath = nil
		}
		idx := i
		if g.canEditCode() && button(dst, Rect{g.w - 74, yy, 66, 20}, "Usuń", false, colDanger, g.mouse) {
			g.deleteStack(idx)
			return
		}
		yy += 22
		yy = g.drawNodes(dst, st.Blocks, left+10, yy, g.w-100, nil, i)
		yy += 8
	}
}

func (g *Game) drawNodes(dst *ebiten.Image, nodes []engine.CodeNode, x, y, w int, prefix []int, stackIdx int) int {
	for i, n := range nodes {
		if y > g.h-26 {
			return y
		}
		path := append(append([]int{}, prefix...), i)
		sel := stackIdx == g.codeSel && pathEqual(path, g.selPath)
		r := Rect{x, y, w, 22}
		fill := brickColor(n.Kind)
		if sel {
			strokeRect(dst, Rect{r.X - 2, r.Y - 2, r.W + 4, r.H + 4}, colAccent, 2)
		}
		label := engine.NodeLabel(n)
		if button(dst, r, label, sel, fill, g.mouse) {
			g.codeSel = stackIdx
			g.selPath = path
			if n.Kind == engine.SetMessage && g.canEditCode() {
				g.focus = fieldCodeText
			} else if g.focus == fieldCodeText {
				g.focus = fieldNone
			}
		}
		if sel && g.canEditCode() {
			bx := x + w + 4
			if engine.CanAdjust(n.Kind) {
				if button(dst, Rect{bx, y, 22, 22}, "−", false, colBtn, g.mouse) {
					g.adjustSelected(-1)
				}
				if button(dst, Rect{bx + 24, y, 22, 22}, "+", false, colBtn, g.mouse) {
					g.adjustSelected(1)
				}
				bx += 52
			}
			if n.Kind == engine.SetMessage {
				shown := n.Text
				if shown == "" {
					shown = "napisz tu…"
				}
				g.inputBox(dst, Rect{bx, y, 140, 22}, shown, fieldCodeText)
			}
			if button(dst, Rect{g.w - 28, y, 20, 22}, "×", false, colDanger, g.mouse) {
				g.deleteSelectedNode()
				return y
			}
		}
		y += 24
		if n.Kind == engine.IfScore || n.Kind == engine.Repeat {
			y = g.drawNodes(dst, n.Body, x+18, y, w-18, path, stackIdx)
			if g.canEditCode() {
				inner := Rect{x + 18, y, 110, 18}
				if button(dst, inner, "+ w środku", false, colBtn, g.mouse) {
					g.codeSel = stackIdx
					g.selPath = path
					g.flash("Kliknij akcję z palety — wpadnie do środka")
				}
				y += 22
			}
		}
	}
	return y
}

func (g *Game) drawHelp(dst *ebiten.Image) {
	r := Rect{g.w/2 - 250, g.h/2 - 180, 500, 360}
	panel(dst, r, colPanel)
	drawText(dst, "Jak grać w BOXEL", r.X+20, r.Y+16, face18, colAccent)
	lines := []string{
		"1. Buduj — klocki stawiasz myszą (LPM / PPM).",
		"2. Klocki — żółte „gdy…” + akcje. +/− to liczby, kliknij wiadomość i pisz.",
		"3. Graj — skrypty działają na tym komputerze (zapis = Zapisz).",
		"4. Pokój — gospodarz wysyła klocki i świat do wszystkich.",
		"5. Znajomi — kod BX-…. Czat w pokoju, szept tylko do znajomych.",
		"6. Drugie okno: uruchom BOXEL jeszcze raz (to aplikacja, nie strona).",
	}
	yy := r.Y + 52
	for _, ln := range lines {
		drawText(dst, ln, r.X+20, yy, face12, colInk)
		yy += 24
	}
	drawText(dst, "Lokalnie: save.json. W sieci: serwer pokoju trzyma te same klocki.", r.X+20, yy+8, face10, colMuted)
	if button(dst, Rect{r.X + 170, r.Y + 310, 160, 34}, "OK, buduję!", false, colPlay, g.mouse) {
		g.help = false
	}
}
