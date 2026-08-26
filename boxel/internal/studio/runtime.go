package studio

import (
	"fmt"
	"math"
	"time"

	"boxel/internal/engine"
)

type Edit struct {
	X, Y, Z, Prev, Next int
}

type Runtime struct {
	World   engine.World
	Player  *engine.PlayerState
	Mode    engine.StudioMode
	Brush   int
	Scripts []engine.ScriptStack
	Name    string

	VM         *engine.ScriptVM
	Snapshot   *engine.World
	Undo       []Edit
	Remotes    []engine.RemoteAvatar
	JumpQueued bool
	Virtual    engine.InputState
	Overlap    map[string]struct{}
	MsgUntil   time.Time
	LastFps    int
	Sounds     []engine.SoundName
	CoinSpin   float64

	OnEdit        func(x, y, z, id int)
	OnPos         func(x, y, z, yaw float64)
	OnScriptEvent func(kind string, x, y, z, block int)
	lastPos       time.Time
}

func New(world engine.World, scripts []engine.ScriptStack, name string) *Runtime {
	rt := &Runtime{
		World:   world,
		Player:  engine.CreatePlayer(),
		Mode:    engine.ModeBuild,
		Brush:   1,
		Scripts: scripts,
		Name:    name,
		VM:      &engine.ScriptVM{},
		Overlap: map[string]struct{}{},
	}
	rt.SetScripts(scripts)
	rt.SyncSpawn()
	return rt
}

func (rt *Runtime) SetScripts(s []engine.ScriptStack) {
	rt.Scripts = engine.SanitizeScripts(s)
	rt.VM.SetScripts(rt.Scripts)
}

func (rt *Runtime) sound(name engine.SoundName) {
	rt.Sounds = append(rt.Sounds, name)
}

func (rt *Runtime) DrainSounds() []engine.SoundName {
	s := rt.Sounds
	rt.Sounds = nil
	return s
}

func (rt *Runtime) api() engine.GameAPI { return runtimeAPI{rt} }

type runtimeAPI struct{ rt *Runtime }

func (a runtimeAPI) AddScore(n int) { a.rt.Player.Score += n }
func (a runtimeAPI) SetMessage(text string) {
	a.rt.Player.Message = text
	a.rt.MsgUntil = time.Now().Add(2500 * time.Millisecond)
}
func (a runtimeAPI) DoWin() {
	if a.rt.Player.Phase != "playing" {
		return
	}
	a.rt.Player.Phase = "won"
	if a.rt.Player.Message == "" {
		a.rt.Player.Message = "WYGRANA!"
	}
	a.rt.sound(engine.SndWin)
}
func (a runtimeAPI) DoLose() {
	if a.rt.Player.Phase != "playing" {
		return
	}
	a.rt.Player.Phase = "lost"
	if a.rt.Player.Message == "" {
		a.rt.Player.Message = "Przegrana"
	}
	a.rt.sound(engine.SndLose)
}
func (a runtimeAPI) TeleportSpawn()                  { engine.TeleportToSpawn(a.rt.Player) }
func (a runtimeAPI) PlaySound(name engine.SoundName) { a.rt.sound(name) }
func (a runtimeAPI) GetScore() int                   { return a.rt.Player.Score }
func (a runtimeAPI) IsStopped() bool {
	return a.rt.Mode != engine.ModePlay || a.rt.Player.Phase != "playing"
}

func (rt *Runtime) SyncSpawn() {
	x, y, z := engine.FindSpawn(rt.World)
	rt.Player.SpawnX, rt.Player.SpawnY, rt.Player.SpawnZ = x, y, z
	if rt.Mode != engine.ModePlay {
		rt.Player.X, rt.Player.Y, rt.Player.Z = x, y, z
	}
}

func (rt *Runtime) SetMode(mode engine.StudioMode) {
	if mode == engine.ModePlay {
		rt.Play()
		return
	}
	if rt.Mode == engine.ModePlay {
		rt.Stop(mode)
		return
	}
	rt.Mode = mode
}

func (rt *Runtime) Play() {
	if rt.Mode != engine.ModePlay {
		snap := engine.CloneWorld(rt.World)
		rt.Snapshot = &snap
	} else if rt.Snapshot != nil {
		engine.CopyWorld(*rt.Snapshot, &rt.World)
	}
	rt.Mode = engine.ModePlay
	rt.VM.Reset()
	rt.Overlap = map[string]struct{}{}
	rt.Player.Lives = 3
	rt.Player.Score = 0
	rt.Player.Coins = 0
	rt.Player.Phase = "playing"
	rt.Player.VX, rt.Player.VY, rt.Player.VZ = 0, 0, 0
	rt.Player.HitCooldown = 0
	rt.SyncSpawn()
	engine.TeleportToSpawn(rt.Player)
	rt.Player.Message = ""
	rt.MsgUntil = time.Time{}
	rt.VM.SetScripts(rt.Scripts)
	rt.VM.Emit(engine.ScriptEvent{Kind: "start"}, rt.api())
	rt.VM.Tick(0.016, rt.api())
	rt.sound(engine.SndPop)
}

func (rt *Runtime) Stop(next engine.StudioMode) {
	if rt.Snapshot != nil {
		engine.CopyWorld(*rt.Snapshot, &rt.World)
		rt.Snapshot = nil
	}
	rt.Mode = next
	rt.VM.Reset()
	rt.Player.Phase = "playing"
	rt.SyncSpawn()
}

func (rt *Runtime) LoadWorld(w engine.World) {
	engine.CopyWorld(w, &rt.World)
	rt.Undo = nil
	rt.SyncSpawn()
}

func (rt *Runtime) ApplyWorldData(sizeX, sizeY, sizeZ int, blocks []int) {
	rt.World.SizeX, rt.World.SizeY, rt.World.SizeZ = sizeX, sizeY, sizeZ
	rt.World.Blocks = make([]uint8, sizeX*sizeY*sizeZ)
	n := min(len(rt.World.Blocks), len(blocks))
	for i := 0; i < n; i++ {
		rt.World.Blocks[i] = uint8(blocks[i])
	}
	rt.Undo = nil
	rt.SyncSpawn()
}

func (rt *Runtime) ApplyNetworkEdit(x, y, z, id int) {
	if !engine.InBounds(rt.World, x, y, z) {
		return
	}
	engine.SetBlock(rt.World, x, y, z, id)
	rt.SyncSpawn()
}

func (rt *Runtime) UndoLast() {
	if rt.Mode == engine.ModePlay || len(rt.Undo) == 0 {
		return
	}
	e := rt.Undo[len(rt.Undo)-1]
	rt.Undo = rt.Undo[:len(rt.Undo)-1]
	engine.SetBlock(rt.World, e.X, e.Y, e.Z, e.Prev)
}

func (rt *Runtime) ApplyEdit(x, y, z, id int) {
	if !engine.InBounds(rt.World, x, y, z) {
		return
	}
	prev := engine.GetBlock(rt.World, x, y, z)
	if prev == id {
		return
	}
	if id == int(engine.Spawn) {
		for i, b := range rt.World.Blocks {
			if b == uint8(engine.Spawn) {
				rt.World.Blocks[i] = uint8(engine.Air)
			}
		}
	}
	engine.SetBlock(rt.World, x, y, z, id)
	rt.Undo = append(rt.Undo, Edit{X: x, Y: y, Z: z, Prev: prev, Next: id})
	if len(rt.Undo) > 80 {
		rt.Undo = rt.Undo[1:]
	}
	rt.SyncSpawn()
	if rt.OnEdit != nil {
		rt.OnEdit(x, y, z, id)
	}
}

func (rt *Runtime) Tick(dt float64, input engine.InputState) {
	if rt.Mode == engine.ModePlay && rt.Player.Phase == "playing" {
		if rt.JumpQueued {
			input.Jump = true
			rt.JumpQueued = false
		}
		input.Forward = input.Forward || rt.Virtual.Forward
		input.Back = input.Back || rt.Virtual.Back
		input.Left = input.Left || rt.Virtual.Left
		input.Right = input.Right || rt.Virtual.Right
		input.Jump = input.Jump || rt.Virtual.Jump
		wasGround := rt.Player.OnGround
		var events []engine.PhysEvent
		engine.StepPlayer(rt.Player, rt.World, input, dt, &events)
		if input.Jump && wasGround && rt.Player.VY > 4 {
			rt.sound(engine.SndJump)
		}
		rt.handleEvents(events)
		rt.emitTouches()
		rt.VM.Tick(dt, rt.api())
	}
	rt.CoinSpin += dt * 2.6
	if rt.OnPos != nil && time.Since(rt.lastPos) > 70*time.Millisecond {
		rt.lastPos = time.Now()
		rt.OnPos(rt.Player.X, rt.Player.Y, rt.Player.Z, rt.Player.Yaw)
	}
	if rt.Player.Message != "" && time.Now().After(rt.MsgUntil) && rt.Player.Phase == "playing" {
		rt.Player.Message = ""
	}
}

func (rt *Runtime) handleEvents(events []engine.PhysEvent) {
	api := rt.api()
	for _, ev := range events {
		switch ev.Type {
		case "coin":
			x, y, z := int(ev.X), int(ev.Y), int(ev.Z)
			if engine.GetBlock(rt.World, x, y, z) != int(engine.Coin) {
				break
			}
			engine.SetBlock(rt.World, x, y, z, int(engine.Air))
			rt.Player.Coins++
			if rt.OnEdit != nil {
				rt.OnEdit(x, y, z, int(engine.Air))
			}
			rt.emitNetScript("coin", x, y, z, 0)
			rt.VM.Emit(engine.ScriptEvent{Kind: "coin"}, api)
		case "hurt":
			rt.Player.HitCooldown = 1
			rt.Player.Lives--
			rt.sound(engine.SndBump)
			rt.VM.Emit(engine.ScriptEvent{Kind: "touch", Block: ev.Block}, api)
			rt.emitNetScript("death", int(ev.X), int(ev.Y), int(ev.Z), ev.Block)
			rt.VM.Emit(engine.ScriptEvent{Kind: "death"}, api)
			if rt.Player.Lives <= 0 {
				api.DoLose()
			} else {
				engine.TeleportToSpawn(rt.Player)
			}
		case "fall":
			rt.Player.Lives--
			rt.sound(engine.SndBump)
			rt.emitNetScript("death", int(rt.Player.X), int(rt.Player.Y), int(rt.Player.Z), 0)
			rt.VM.Emit(engine.ScriptEvent{Kind: "death"}, api)
			if rt.Player.Lives <= 0 {
				api.DoLose()
			} else {
				engine.TeleportToSpawn(rt.Player)
			}
		case "goal":
			rt.emitNetScript("goal", int(ev.X), int(ev.Y), int(ev.Z), 0)
			rt.VM.Emit(engine.ScriptEvent{Kind: "goal"}, api)
		case "checkpoint":
			rt.Player.SpawnX, rt.Player.SpawnY, rt.Player.SpawnZ = ev.X, ev.Y, ev.Z
			rt.Player.Message = "Checkpoint!"
			rt.MsgUntil = time.Now().Add(1500 * time.Millisecond)
			rt.sound(engine.SndPop)
		case "bounce":
			rt.sound(engine.SndJump)
		}
	}
}

func (rt *Runtime) emitTouches() {
	next := map[string]struct{}{}
	api := rt.api()
	for _, cell := range engine.CollectOverlap(rt.World, rt.Player, -0.12) {
		if cell.ID == int(engine.Air) {
			continue
		}
		key := fmt.Sprintf("%d:%d:%d:%d", cell.X, cell.Y, cell.Z, cell.ID)
		next[key] = struct{}{}
		if _, ok := rt.Overlap[key]; !ok {
			rt.VM.Emit(engine.ScriptEvent{Kind: "touch", Block: cell.ID}, api)
		}
	}
	rt.Overlap = next
}

func (rt *Runtime) emitNetScript(kind string, x, y, z, block int) {
	if rt.OnScriptEvent != nil {
		rt.OnScriptEvent(kind, x, y, z, block)
	}
}

func (rt *Runtime) HUD() engine.HudState {
	phase := "build"
	if rt.Mode == engine.ModePlay {
		phase = rt.Player.Phase
	}
	debug := ""
	if rt.LastFps > 0 {
		debug = fmt.Sprintf("%d fps", rt.LastFps)
	}
	return engine.HudState{
		Mode: rt.Mode, Score: rt.Player.Score, Coins: rt.Player.Coins,
		Lives: rt.Player.Lives, Message: rt.Player.Message, Phase: phase, Debug: debug,
	}
}

func (rt *Runtime) PlaceFromRay(hit, prev *[3]int, erase bool) {
	if rt.Mode == engine.ModePlay || rt.Mode == engine.ModeCode {
		return
	}
	if erase || rt.Brush == int(engine.Air) {
		if hit != nil {
			rt.ApplyEdit(hit[0], hit[1], hit[2], int(engine.Air))
		}
		return
	}
	if prev != nil {
		rt.ApplyEdit(prev[0], prev[1], prev[2], rt.Brush)
	}
}

func OrbitFollow(player *engine.PlayerState, camX, camY, camZ, dt float64) (float64, float64, float64) {
	dist, height := 10.0, 5.5
	yaw := player.Yaw
	tx := player.X + math.Sin(yaw)*dist
	ty := player.Y + height
	tz := player.Z + math.Cos(yaw)*dist
	k := 1 - math.Pow(0.0008, dt)
	return camX + (tx-camX)*k, camY + (ty-camY)*k, camZ + (tz-camZ)*k
}
