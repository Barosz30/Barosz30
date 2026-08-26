export type StudioMode = 'build' | 'play' | 'code'

export type BlockId =
  | 0
  | 1
  | 2
  | 3
  | 4
  | 5
  | 6
  | 7
  | 8
  | 9
  | 10
  | 11
  | 12
  | 13
  | 14
  | 15

export type SoundName = 'coin' | 'jump' | 'win' | 'bump' | 'lose' | 'pop'

export interface World {
  sizeX: number
  sizeY: number
  sizeZ: number
  blocks: Uint8Array
}

export type CodeKind =
  | 'hatStart'
  | 'hatCoin'
  | 'hatTouch'
  | 'hatGoal'
  | 'hatDeath'
  | 'hatAlways'
  | 'addScore'
  | 'setMessage'
  | 'win'
  | 'lose'
  | 'teleport'
  | 'sound'
  | 'wait'
  | 'ifScore'
  | 'repeat'

export interface CodeNode {
  id: string
  kind: CodeKind
  amount?: number
  text?: string
  block?: BlockId
  sound?: SoundName
  seconds?: number
  gte?: number
  times?: number
  body?: CodeNode[]
}

export interface ScriptStack {
  id: string
  blocks: CodeNode[]
}

export interface SavedGame {
  version: 1
  name: string
  sizeX: number
  sizeY: number
  sizeZ: number
  blocks: number[]
  scripts: ScriptStack[]
}

export interface HudState {
  mode: StudioMode
  score: number
  coins: number
  lives: number
  message: string
  phase: 'playing' | 'won' | 'lost' | 'build'
  debug?: string
}

export interface PlayerState {
  x: number
  y: number
  z: number
  vx: number
  vy: number
  vz: number
  yaw: number
  onGround: boolean
  inWater: boolean
  lives: number
  score: number
  coins: number
  spawnX: number
  spawnY: number
  spawnZ: number
  hitCooldown: number
  bounceCooldown: number
  goalCooldown: number
  message: string
  phase: 'playing' | 'won' | 'lost'
}

export interface InputState {
  forward: boolean
  back: boolean
  left: boolean
  right: boolean
  jump: boolean
}

export type PhysEvent =
  | { type: 'coin'; x: number; y: number; z: number }
  | { type: 'hurt'; block: BlockId }
  | { type: 'goal' }
  | { type: 'checkpoint'; x: number; y: number; z: number }
  | { type: 'fall' }
  | { type: 'bounce' }

export interface OverlapCell {
  x: number
  y: number
  z: number
  id: number
}

export interface GameApi {
  addScore: (n: number) => void
  setMessage: (text: string) => void
  win: () => void
  lose: () => void
  teleportSpawn: () => void
  playSound: (name: SoundName) => void
  getScore: () => number
  isStopped: () => boolean
}
