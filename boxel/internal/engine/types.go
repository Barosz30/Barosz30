package engine

type StudioMode string

const (
	ModeBuild StudioMode = "build"
	ModePlay  StudioMode = "play"
	ModeCode  StudioMode = "code"
)

type SoundName string

const (
	SndCoin SoundName = "coin"
	SndJump SoundName = "jump"
	SndWin  SoundName = "win"
	SndBump SoundName = "bump"
	SndLose SoundName = "lose"
	SndPop  SoundName = "pop"
)

type World struct {
	SizeX, SizeY, SizeZ int
	Blocks              []uint8
}

type CodeKind string

const (
	HatStart   CodeKind = "hatStart"
	HatCoin    CodeKind = "hatCoin"
	HatTouch   CodeKind = "hatTouch"
	HatGoal    CodeKind = "hatGoal"
	HatDeath   CodeKind = "hatDeath"
	HatAlways  CodeKind = "hatAlways"
	AddScore   CodeKind = "addScore"
	SetMessage CodeKind = "setMessage"
	Win        CodeKind = "win"
	Lose       CodeKind = "lose"
	Teleport   CodeKind = "teleport"
	Sound      CodeKind = "sound"
	Wait       CodeKind = "wait"
	IfScore    CodeKind = "ifScore"
	Repeat     CodeKind = "repeat"
)

type CodeNode struct {
	ID      string     `json:"id"`
	Kind    CodeKind   `json:"kind"`
	Amount  int        `json:"amount,omitempty"`
	Text    string     `json:"text,omitempty"`
	Block   int        `json:"block,omitempty"`
	Sound   SoundName  `json:"sound,omitempty"`
	Seconds float64    `json:"seconds,omitempty"`
	Gte     int        `json:"gte,omitempty"`
	Times   int        `json:"times,omitempty"`
	Body    []CodeNode `json:"body,omitempty"`
}

type ScriptStack struct {
	ID     string     `json:"id"`
	Blocks []CodeNode `json:"blocks"`
}

type SavedGame struct {
	Version int           `json:"version"`
	Name    string        `json:"name"`
	SizeX   int           `json:"sizeX"`
	SizeY   int           `json:"sizeY"`
	SizeZ   int           `json:"sizeZ"`
	Blocks  []int         `json:"blocks"`
	Scripts []ScriptStack `json:"scripts"`
}

type HudState struct {
	Mode    StudioMode
	Score   int
	Coins   int
	Lives   int
	Message string
	Phase   string
	Debug   string
}

type PlayerState struct {
	X, Y, Z                     float64
	VX, VY, VZ                  float64
	Yaw                         float64
	OnGround, InWater           bool
	Lives, Score, Coins         int
	SpawnX, SpawnY, SpawnZ      float64
	HitCooldown, BounceCooldown float64
	GoalCooldown                float64
	Message                     string
	Phase                       string
}

type InputState struct {
	Forward, Back, Left, Right, Jump bool
}

type PhysEvent struct {
	Type    string
	X, Y, Z float64
	Block   int
}

type OverlapCell struct {
	X, Y, Z int
	ID      int
}

type GameAPI interface {
	AddScore(n int)
	SetMessage(text string)
	DoWin()
	DoLose()
	TeleportSpawn()
	PlaySound(name SoundName)
	GetScore() int
	IsStopped() bool
}

type RemoteAvatar struct {
	ID, Name, Color string
	X, Y, Z, Yaw    float64
	Score           int
}
