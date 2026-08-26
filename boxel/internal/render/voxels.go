package render

import (
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"boxel/internal/engine"
)

var whiteImg *ebiten.Image

func init() {
	whiteImg = ebiten.NewImage(1, 1)
	whiteImg.Fill(color.White)
}

type face struct {
	pts        [4][3]float64
	r, g, b, a float32
	depth      float64
}

func DrawWorld(dst *ebiten.Image, vx, vy, vw, vh int, cam Camera, w engine.World, player *engine.PlayerState, remotes []engine.RemoteAvatar, ghost *[3]int, ghostID int, mode engine.StudioMode, coinSpin float64, blink bool) {
	sub := dst.SubImage(image.Rect(vx, vy, vx+vw, vy+vh)).(*ebiten.Image)
	sub.Fill(color.RGBA{0x74, 0xc7, 0xff, 0xff})

	cam.SyncEye()
	faces := collectFaces(w, cam, coinSpin)
	if !(blink && mode == engine.ModePlay && player.HitCooldown > 0) {
		addCube(&faces, cam, player.X, player.Y+0.32, player.Z, 0.28, 0.31, 0.28, engine.RGB{R: 0xff, G: 0x6b, B: 0x3d}, 1)
		addCube(&faces, cam, player.X, player.Y+0.78, player.Z, 0.24, 0.21, 0.24, engine.RGB{R: 0xff, G: 0xd3, B: 0x6a}, 1)
	}
	for _, r := range remotes {
		col := parseHex(r.Color)
		addCube(&faces, cam, r.X, r.Y+0.32, r.Z, 0.28, 0.31, 0.28, col, 1)
		addCube(&faces, cam, r.X, r.Y+0.78, r.Z, 0.24, 0.21, 0.24, engine.RGB{R: 0xff, G: 0xf6, B: 0xe8}, 1)
	}
	if ghost != nil && mode != engine.ModePlay {
		def := engine.Def(ghostID)
		addCube(&faces, cam, float64(ghost[0])+0.5, float64(ghost[1])+0.5, float64(ghost[2])+0.5, 0.5, 0.5, 0.5, def.Color, 0.38)
	}

	sort.Slice(faces, func(i, j int) bool { return faces[i].depth > faces[j].depth })

	verts := make([]ebiten.Vertex, 0, len(faces)*4)
	idx := make([]uint16, 0, len(faces)*6)
	for _, f := range faces {
		ok := true
		var s [4][2]float64
		for i := 0; i < 4; i++ {
			x, y, _, on := cam.Project(f.pts[i][0], f.pts[i][1], f.pts[i][2], vw, vh)
			if !on {
				ok = false
				break
			}
			s[i][0], s[i][1] = x, y
		}
		if !ok {
			continue
		}
		base := uint16(len(verts))
		for i := 0; i < 4; i++ {
			verts = append(verts, ebiten.Vertex{
				DstX: float32(s[i][0]), DstY: float32(s[i][1]),
				SrcX: 0, SrcY: 0,
				ColorR: f.r, ColorG: f.g, ColorB: f.b, ColorA: f.a,
			})
		}
		idx = append(idx, base, base+1, base+2, base, base+2, base+3)
	}
	if len(idx) > 0 {
		sub.DrawTriangles(verts, idx, whiteImg, nil)
	}

	drawBounds(sub, cam, w, vw, vh)
}

func collectFaces(w engine.World, cam Camera, coinSpin float64) []face {
	var out []face
	for y := 0; y < w.SizeY; y++ {
		for z := 0; z < w.SizeZ; z++ {
			for x := 0; x < w.SizeX; x++ {
				id := engine.GetBlock(w, x, y, z)
				if id == int(engine.Air) {
					continue
				}
				def := engine.Def(id)
				cx, cy, cz := float64(x)+0.5, float64(y)+0.5, float64(z)+0.5
				hs, vs := 0.5, 0.5
				if id == int(engine.Coin) {
					hs, vs = 0.23, 0.23
					cy = float64(y) + 0.55 + math.Sin(coinSpin*2+float64(x+z))*0.12
				}
				a := float32(1)
				if id == int(engine.Water) {
					a = 0.58
				}
				try := func(nx, ny, nz int, shade float32, pts [4][3]float64) {
					n := engine.GetBlock(w, x+nx, y+ny, z+nz)
					if id != int(engine.Coin) && n != int(engine.Air) {
						return
					}
					r, g, b := def.Color.Float()
					switch {
					case shade > 1:
						r, g, b = def.Light.Float()
					case shade < 0.75:
						r, g, b = def.Dark.Float()
					}
					d := dist(cam.EyeX, cam.EyeY, cam.EyeZ, cx+float64(nx)*0.5, cy+float64(ny)*0.5, cz+float64(nz)*0.5)
					out = append(out, face{pts: pts, r: r, g: g, b: b, a: a, depth: d})
				}
				// +Y top
				try(0, 1, 0, 1.2, [4][3]float64{
					{cx - hs, cy + vs, cz - hs}, {cx + hs, cy + vs, cz - hs}, {cx + hs, cy + vs, cz + hs}, {cx - hs, cy + vs, cz + hs},
				})
				// -Y
				try(0, -1, 0, 0.55, [4][3]float64{
					{cx - hs, cy - vs, cz + hs}, {cx + hs, cy - vs, cz + hs}, {cx + hs, cy - vs, cz - hs}, {cx - hs, cy - vs, cz - hs},
				})
				// +X
				try(1, 0, 0, 0.9, [4][3]float64{
					{cx + hs, cy - vs, cz - hs}, {cx + hs, cy - vs, cz + hs}, {cx + hs, cy + vs, cz + hs}, {cx + hs, cy + vs, cz - hs},
				})
				// -X
				try(-1, 0, 0, 0.7, [4][3]float64{
					{cx - hs, cy - vs, cz + hs}, {cx - hs, cy - vs, cz - hs}, {cx - hs, cy + vs, cz - hs}, {cx - hs, cy + vs, cz + hs},
				})
				// +Z
				try(0, 0, 1, 0.85, [4][3]float64{
					{cx + hs, cy - vs, cz + hs}, {cx - hs, cy - vs, cz + hs}, {cx - hs, cy + vs, cz + hs}, {cx + hs, cy + vs, cz + hs},
				})
				// -Z
				try(0, 0, -1, 0.75, [4][3]float64{
					{cx - hs, cy - vs, cz - hs}, {cx + hs, cy - vs, cz - hs}, {cx + hs, cy + vs, cz - hs}, {cx - hs, cy + vs, cz - hs},
				})
			}
		}
	}
	return out
}

func addCube(out *[]face, cam Camera, cx, cy, cz, hx, hy, hz float64, col engine.RGB, a float32) {
	r, g, b := col.Float()
	faces := []struct {
		shade      float32
		pts        [4][3]float64
		nx, ny, nz float64
	}{
		{1.15, [4][3]float64{{cx - hx, cy + hy, cz - hz}, {cx + hx, cy + hy, cz - hz}, {cx + hx, cy + hy, cz + hz}, {cx - hx, cy + hy, cz + hz}}, 0, 1, 0},
		{0.6, [4][3]float64{{cx - hx, cy - hy, cz + hz}, {cx + hx, cy - hy, cz + hz}, {cx + hx, cy - hy, cz - hz}, {cx - hx, cy - hy, cz - hz}}, 0, -1, 0},
		{0.9, [4][3]float64{{cx + hx, cy - hy, cz - hz}, {cx + hx, cy - hy, cz + hz}, {cx + hx, cy + hy, cz + hz}, {cx + hx, cy + hy, cz - hz}}, 1, 0, 0},
		{0.7, [4][3]float64{{cx - hx, cy - hy, cz + hz}, {cx - hx, cy - hy, cz - hz}, {cx - hx, cy + hy, cz - hz}, {cx - hx, cy + hy, cz + hz}}, -1, 0, 0},
		{0.85, [4][3]float64{{cx + hx, cy - hy, cz + hz}, {cx - hx, cy - hy, cz + hz}, {cx - hx, cy + hy, cz + hz}, {cx + hx, cy + hy, cz + hz}}, 0, 0, 1},
		{0.75, [4][3]float64{{cx - hx, cy - hy, cz - hz}, {cx + hx, cy - hy, cz - hz}, {cx + hx, cy + hy, cz - hz}, {cx - hx, cy + hy, cz - hz}}, 0, 0, -1},
	}
	for _, f := range faces {
		d := dist(cam.EyeX, cam.EyeY, cam.EyeZ, cx+f.nx*hx, cy+f.ny*hy, cz+f.nz*hz)
		*out = append(*out, face{pts: f.pts, r: r * f.shade, g: g * f.shade, b: b * f.shade, a: a, depth: d})
	}
}

func dist(ax, ay, az, bx, by, bz float64) float64 {
	dx, dy, dz := ax-bx, ay-by, az-bz
	return dx*dx + dy*dy + dz*dz
}

func parseHex(s string) engine.RGB {
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	var n uint32
	for _, c := range s {
		n <<= 4
		switch {
		case c >= '0' && c <= '9':
			n += uint32(c - '0')
		case c >= 'a' && c <= 'f':
			n += uint32(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			n += uint32(c - 'A' + 10)
		}
	}
	return engine.RGB{R: uint8(n >> 16), G: uint8(n >> 8), B: uint8(n)}
}

func drawBounds(dst *ebiten.Image, cam Camera, w engine.World, vw, vh int) {
	sx0, sz0 := 0.0, 0.0
	sx1, sz1 := float64(w.SizeX), float64(w.SizeZ)
	pts := [][3]float64{
		{sx0, 0, sz0}, {sx1, 0, sz0}, {sx1, 0, sz1}, {sx0, 0, sz1},
	}
	var screen [][2]float32
	for _, p := range pts {
		x, y, _, ok := cam.Project(p[0], p[1], p[2], vw, vh)
		if !ok {
			return
		}
		screen = append(screen, [2]float32{float32(x), float32(y)})
	}
	col := color.RGBA{0x1a, 0x10, 0x20, 180}
	for i := 0; i < 4; i++ {
		j := (i + 1) % 4
		vector.StrokeLine(dst, screen[i][0], screen[i][1], screen[j][0], screen[j][1], 2, col, false)
	}
}
