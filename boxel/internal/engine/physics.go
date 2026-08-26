package engine

import "math"

const (
	gravity      = 24.0
	waterGravity = 8.0
	jumpVel      = 8.4
	waterJump    = 4.2
	speed        = 6.2
	airSpeed     = 4.4
	waterSpeed   = 3.2
	halfW        = 0.28
	height       = 0.88
)

func PlayerAABB(p *PlayerState) (minX, maxX, minY, maxY, minZ, maxZ float64) {
	return p.X - halfW, p.X + halfW, p.Y, p.Y + height, p.Z - halfW, p.Z + halfW
}

func CollectOverlap(w World, p *PlayerState, extraY float64) []OverlapCell {
	minX, maxX, minY, maxY, minZ, maxZ := PlayerAABB(p)
	x0, x1 := int(math.Floor(minX)), int(math.Floor(maxX))
	y0, y1 := int(math.Floor(minY+extraY)), int(math.Floor(maxY-0.05))
	z0, z1 := int(math.Floor(minZ)), int(math.Floor(maxZ))
	seen := map[[3]int]struct{}{}
	var cells []OverlapCell
	for y := y0; y <= y1; y++ {
		for z := z0; z <= z1; z++ {
			for x := x0; x <= x1; x++ {
				if !InBounds(w, x, y, z) {
					continue
				}
				key := [3]int{x, y, z}
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				cells = append(cells, OverlapCell{X: x, Y: y, Z: z, ID: GetBlock(w, x, y, z)})
			}
		}
	}
	return cells
}

func overlapsSolid(w World, x, y, z float64) bool {
	minX, maxX := int(math.Floor(x-halfW)), int(math.Floor(x+halfW))
	minY, maxY := int(math.Floor(y+0.001)), int(math.Floor(y+height-0.001))
	minZ, maxZ := int(math.Floor(z-halfW)), int(math.Floor(z+halfW))
	for by := minY; by <= maxY; by++ {
		for bz := minZ; bz <= maxZ; bz++ {
			for bx := minX; bx <= maxX; bx++ {
				if IsSolid(GetBlock(w, bx, by, bz)) {
					return true
				}
			}
		}
	}
	return false
}

func feetBlock(w World, p *PlayerState) int {
	return GetBlock(w, int(math.Floor(p.X)), int(math.Floor(p.Y-0.08)), int(math.Floor(p.Z)))
}

func CreatePlayer() *PlayerState {
	return &PlayerState{
		X: 12.5, Y: 4, Z: 12.5,
		Lives: 3, SpawnX: 12.5, SpawnY: 4, SpawnZ: 12.5,
		Phase: "playing",
	}
}

func TeleportToSpawn(p *PlayerState) {
	p.X, p.Y, p.Z = p.SpawnX, p.SpawnY, p.SpawnZ
	p.VX, p.VY, p.VZ = 0, 0, 0
}

func StepPlayer(p *PlayerState, w World, in InputState, dt float64, events *[]PhysEvent) {
	if p.Phase != "playing" {
		p.VX, p.VY, p.VZ = 0, 0, 0
		return
	}
	clamped := dt
	if clamped > 0.05 {
		clamped = 0.05
	}
	p.HitCooldown = math.Max(0, p.HitCooldown-clamped)
	p.BounceCooldown = math.Max(0, p.BounceCooldown-clamped)
	p.GoalCooldown = math.Max(0, p.GoalCooldown-clamped)

	cells := CollectOverlap(w, p, 0)
	p.InWater = false
	for _, c := range cells {
		if c.ID == int(Water) {
			p.InWater = true
			break
		}
	}

	g := gravity
	if p.InWater {
		g = waterGravity
	}
	p.VY -= g * clamped

	feet := feetBlock(w, p)
	onIce := p.OnGround && feet == int(Ice)
	spd := airSpeed
	if p.InWater {
		spd = waterSpeed
	} else if p.OnGround {
		spd = speed
	}

	sin, cos := math.Sin(p.Yaw), math.Cos(p.Yaw)
	wishX, wishZ := 0.0, 0.0
	if in.Forward {
		wishX -= sin
		wishZ -= cos
	}
	if in.Back {
		wishX += sin
		wishZ += cos
	}
	if in.Left {
		wishX -= cos
		wishZ += sin
	}
	if in.Right {
		wishX += cos
		wishZ -= sin
	}
	if len := math.Hypot(wishX, wishZ); len > 0 {
		wishX = wishX / len * spd
		wishZ = wishZ / len * spd
	}

	accel := 6.0
	if p.OnGround {
		accel = 18
		if onIce {
			accel = 4
		}
	}
	p.VX += (wishX - p.VX) * math.Min(1, accel*clamped)
	p.VZ += (wishZ - p.VZ) * math.Min(1, accel*clamped)

	if in.Jump && (p.OnGround || p.InWater) {
		if p.InWater {
			p.VY = waterJump
		} else {
			p.VY = jumpVel
		}
		p.OnGround = false
	}

	moveAxis(p, w, p.VX*clamped, 0, 0)
	moveAxis(p, w, 0, 0, p.VZ*clamped)
	p.OnGround = false
	moveAxis(p, w, 0, p.VY*clamped, 0)
	if p.VY == 0 && overlapsSolid(w, p.X, p.Y-0.04, p.Z) {
		p.OnGround = true
	}

	if p.Y < -2 {
		*events = append(*events, PhysEvent{Type: "fall"})
		return
	}

	after := CollectOverlap(w, p, -0.12)
	seenCoin := map[[3]int]struct{}{}
	for _, cell := range after {
		switch {
		case cell.ID == int(Coin):
			key := [3]int{cell.X, cell.Y, cell.Z}
			if _, ok := seenCoin[key]; ok {
				continue
			}
			seenCoin[key] = struct{}{}
			*events = append(*events, PhysEvent{Type: "coin", X: float64(cell.X), Y: float64(cell.Y), Z: float64(cell.Z)})
		case cell.ID == int(Goal) && p.GoalCooldown <= 0:
			p.GoalCooldown = 1.2
			*events = append(*events, PhysEvent{Type: "goal"})
		case cell.ID == int(Checkpoint):
			*events = append(*events, PhysEvent{
				Type: "checkpoint",
				X:    float64(cell.X) + 0.5,
				Y:    float64(cell.Y) + 1,
				Z:    float64(cell.Z) + 0.5,
			})
		case IsHazard(cell.ID) && p.HitCooldown <= 0:
			*events = append(*events, PhysEvent{Type: "hurt", Block: cell.ID})
		}
	}

	land := feetBlock(w, p)
	if p.OnGround && land == int(Bounce) && p.BounceCooldown <= 0 {
		p.VY = 11
		p.OnGround = false
		p.BounceCooldown = 0.2
		*events = append(*events, PhysEvent{Type: "bounce"})
	}
}

func moveAxis(p *PlayerState, w World, dx, dy, dz float64) {
	p.X += dx
	p.Y += dy
	p.Z += dz
	if !overlapsSolid(w, p.X, p.Y, p.Z) {
		return
	}
	if dx != 0 {
		p.X -= dx
		p.VX = 0
	}
	if dz != 0 {
		p.Z -= dz
		p.VZ = 0
	}
	if dy != 0 {
		p.Y -= dy
		if dy < 0 {
			p.OnGround = true
		}
		p.VY = 0
	}
}
