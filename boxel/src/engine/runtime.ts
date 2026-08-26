import {
  Box3,
  Box3Helper,
  BoxGeometry,
  Color,
  DirectionalLight,
  DynamicDrawUsage,
  Fog,
  HemisphereLight,
  InstancedMesh,
  Mesh,
  MeshLambertMaterial,
  Object3D,
  PerspectiveCamera,
  Scene,
  Vector2,
  Vector3,
  WebGLRenderer,
} from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'
import { playSound, unlockAudio } from './audio.ts'
import { AIR, blockDef, SPAWN } from './blocks.ts'
import { ScriptVM } from './interpreter.ts'
import {
  collectOverlap,
  createPlayer,
  stepPlayer,
  teleportToSpawn,
} from './physics.ts'
import { voxelRaycast } from './raycast.ts'
import { disposeTextures, textureFor } from './textures.ts'
import type {
  GameApi,
  HudState,
  InputState,
  PhysEvent,
  PlayerState,
  ScriptStack,
  StudioMode,
  World,
} from './types.ts'
import {
  cloneWorld,
  copyWorld,
  findSpawn,
  getBlock,
  inBounds,
  setBlock,
} from './world.ts'

const MAX_INSTANCES = 24 * 16 * 24
const DUMMY = new Object3D()

export interface RuntimeCallbacks {
  onHud: (hud: HudState) => void
  onBrush?: (id: number) => void
}

interface Edit {
  x: number
  y: number
  z: number
  prev: number
  next: number
}

export class BoxelRuntime {
  readonly world: World
  readonly player: PlayerState
  mode: StudioMode = 'build'
  brush: number = 1

  private canvas: HTMLCanvasElement
  private renderer: WebGLRenderer
  private scene: Scene
  private camera: PerspectiveCamera
  private controls: OrbitControls
  private meshes = new Map<number, InstancedMesh>()
  private ghost: Mesh
  private playerMesh: Mesh
  private eyeL: Mesh
  private eyeR: Mesh
  private vm = new ScriptVM()
  private snapshot: World | null = null
  private raf = 0
  private lastT = 0
  private keys = new Set<string>()
  private virtual: InputState = emptyInput()
  private callbacks: RuntimeCallbacks
  private undoStack: Edit[] = []
  private pointer = { x: 0, y: 0, downX: 0, downY: 0, dragging: false, button: -1 }
  private overlapKeys = new Set<string>()
  private coinSpin = 0
  private coinCells: { x: number; y: number; z: number }[] = []
  private disposed = false
  private hudMessageUntil = 0
  private lastHud = ''
  private ro: ResizeObserver
  private jumpQueued = false

  constructor(canvas: HTMLCanvasElement, world: World, callbacks: RuntimeCallbacks) {
    this.canvas = canvas
    this.world = world
    this.callbacks = callbacks
    this.player = createPlayer()

    this.scene = new Scene()
    this.scene.background = new Color(0x74c7ff)
    this.scene.fog = new Fog(0x74c7ff, 28, 64)

    this.camera = new PerspectiveCamera(55, 1, 0.1, 120)
    this.camera.position.set(18, 16, 28)

    this.renderer = new WebGLRenderer({ canvas, antialias: false, alpha: false })
    this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5))
    this.renderer.setClearColor(0x74c7ff)

    this.scene.add(new HemisphereLight(0xfff0d8, 0x3a4a6a, 1.15))
    const sun = new DirectionalLight(0xffffff, 1.05)
    sun.position.set(20, 30, 12)
    this.scene.add(sun)

    const bounds = new Box3(new Vector3(0, 0, 0), new Vector3(world.sizeX, world.sizeY, world.sizeZ))
    this.scene.add(new Box3Helper(bounds, 0x1a1020))

    const geo = new BoxGeometry(1, 1, 1)
    for (let id = 1; id <= 15; id++) {
      const mat = new MeshLambertMaterial({
        map: textureFor(id),
      })
      if (id === 7) {
        mat.transparent = true
        mat.opacity = 0.58
        mat.depthWrite = false
      }
      const mesh = new InstancedMesh(geo, mat, MAX_INSTANCES)
      mesh.instanceMatrix.setUsage(DynamicDrawUsage)
      mesh.count = 0
      mesh.frustumCulled = false
      this.meshes.set(id, mesh)
      this.scene.add(mesh)
    }

    this.ghost = new Mesh(
      geo,
      new MeshLambertMaterial({ color: 0xffffff, transparent: true, opacity: 0.35 }),
    )
    this.ghost.visible = false
    this.scene.add(this.ghost)

    this.playerMesh = new Mesh(
      new BoxGeometry(0.55, 0.62, 0.55),
      new MeshLambertMaterial({ color: 0xff6b3d }),
    )
    const head = new Mesh(
      new BoxGeometry(0.48, 0.42, 0.48),
      new MeshLambertMaterial({ color: 0xffd36a }),
    )
    head.position.y = 0.52
    this.eyeL = new Mesh(new BoxGeometry(0.1, 0.1, 0.08), new MeshLambertMaterial({ color: 0x1a1020 }))
    this.eyeR = new Mesh(new BoxGeometry(0.1, 0.1, 0.08), new MeshLambertMaterial({ color: 0x1a1020 }))
    this.eyeL.position.set(-0.12, 0.08, -0.22)
    this.eyeR.position.set(0.12, 0.08, -0.22)
    head.add(this.eyeL, this.eyeR)
    this.playerMesh.add(head)
    this.scene.add(this.playerMesh)

    this.controls = new OrbitControls(this.camera, canvas)
    this.controls.enableDamping = true
    this.controls.dampingFactor = 0.12
    this.controls.maxPolarAngle = Math.PI * 0.49
    this.controls.minDistance = 6
    this.controls.maxDistance = 50
    this.controls.target.set(world.sizeX / 2, 3, world.sizeZ / 2)

    this.bind()
    this.resize()
    this.ro = new ResizeObserver(() => this.resize())
    this.ro.observe(canvas.parentElement ?? canvas)

    this.rebuild()
    this.syncSpawnFromWorld()
    this.lastT = performance.now()
    this.raf = requestAnimationFrame(this.loop)
    this.pushHud()
  }

  setScripts(scripts: ScriptStack[]): void {
    this.vm.setScripts(scripts)
  }

  setBrush(id: number): void {
    this.brush = id
  }

  setMode(mode: StudioMode): void {
    if (mode === 'play') this.play()
    else if (this.mode === 'play') this.stop(mode)
    else {
      this.mode = mode
      this.controls.enabled = true
      this.pushHud()
    }
  }

  setVirtualInput(partial: Partial<InputState>): void {
    Object.assign(this.virtual, partial)
  }

  queueJump(): void {
    this.jumpQueued = true
  }

  loadWorld(world: World): void {
    copyWorld(world, this.world)
    this.undoStack = []
    this.rebuild()
    this.syncSpawnFromWorld()
    this.pushHud()
  }

  undo(): void {
    if (this.mode === 'play') return
    const edit = this.undoStack.pop()
    if (!edit) return
    setBlock(this.world, edit.x, edit.y, edit.z, edit.prev)
    this.rebuild()
  }

  play(): void {
    unlockAudio()
    if (this.mode !== 'play') {
      this.snapshot = cloneWorld(this.world)
    } else if (this.snapshot) {
      copyWorld(this.snapshot, this.world)
      this.rebuild()
    }
    this.mode = 'play'
    this.controls.enabled = false
    this.vm.reset()
    this.overlapKeys.clear()
    this.player.lives = 3
    this.player.score = 0
    this.player.coins = 0
    this.player.phase = 'playing'
    this.player.vx = 0
    this.player.vy = 0
    this.player.vz = 0
    this.player.hitCooldown = 0
    this.syncSpawnFromWorld()
    teleportToSpawn(this.player)
    this.player.message = ''
    this.hudMessageUntil = 0
    this.vm.emit({ kind: 'start' }, this.api())
    playSound('pop')
    this.pushHud()
  }

  stop(next: StudioMode = 'build'): void {
    if (this.snapshot) {
      copyWorld(this.snapshot, this.world)
      this.snapshot = null
      this.rebuild()
    }
    this.mode = next
    this.controls.enabled = true
    this.vm.reset()
    this.player.phase = 'playing'
    this.syncSpawnFromWorld()
    this.pushHud()
  }

  dispose(): void {
    this.disposed = true
    cancelAnimationFrame(this.raf)
    this.ro.disconnect()
    this.unbind()
    this.controls.dispose()
    this.renderer.dispose()
    disposeTextures()
  }

  private api(): GameApi {
    return {
      addScore: (n) => {
        this.player.score += n
        this.pushHud()
      },
      setMessage: (text) => {
        this.player.message = text
        this.hudMessageUntil = performance.now() + 2500
        this.pushHud()
      },
      win: () => {
        if (this.player.phase !== 'playing') return
        this.player.phase = 'won'
        this.player.message = this.player.message || 'WYGRANA!'
        playSound('win')
        this.pushHud()
      },
      lose: () => {
        if (this.player.phase !== 'playing') return
        this.player.phase = 'lost'
        this.player.message = this.player.message || 'Przegrana'
        playSound('lose')
        this.pushHud()
      },
      teleportSpawn: () => teleportToSpawn(this.player),
      playSound,
      getScore: () => this.player.score,
      isStopped: () => this.mode !== 'play' || this.player.phase !== 'playing',
    }
  }

  private syncSpawnFromWorld(): void {
    const s = findSpawn(this.world)
    this.player.spawnX = s.x
    this.player.spawnY = s.y
    this.player.spawnZ = s.z
    if (this.mode !== 'play') {
      this.player.x = s.x
      this.player.y = s.y
      this.player.z = s.z
    }
  }

  private rebuild(): void {
    this.coinCells = []
    const buckets = new Map<number, { x: number; y: number; z: number }[]>()
    const { sizeX, sizeY, sizeZ, blocks } = this.world
    for (let y = 0; y < sizeY; y++) {
      for (let z = 0; z < sizeZ; z++) {
        for (let x = 0; x < sizeX; x++) {
          const id = blocks[x + sizeX * (z + sizeZ * y)] ?? 0
          if (id === AIR) continue
          let list = buckets.get(id)
          if (!list) {
            list = []
            buckets.set(id, list)
          }
          list.push({ x, y, z })
        }
      }
    }

    for (let id = 1; id <= 15; id++) {
      const mesh = this.meshes.get(id)
      if (!mesh) continue
      const list = buckets.get(id) ?? []
      mesh.count = list.length
      for (let i = 0; i < list.length; i++) {
        const p = list[i]
        const scale = id === 12 ? 0.46 : 1
        const yOff = id === 12 ? 0.12 : 0
        DUMMY.position.set(p.x + 0.5, p.y + 0.5 + yOff, p.z + 0.5)
        DUMMY.rotation.set(0, 0, 0)
        DUMMY.scale.setScalar(scale)
        DUMMY.updateMatrix()
        mesh.setMatrixAt(i, DUMMY.matrix)
      }
      mesh.instanceMatrix.needsUpdate = true
      if (id === 12) this.coinCells = list
    }
  }

  private loop = (now: number): void => {
    if (this.disposed) return
    const dt = Math.min(0.05, (now - this.lastT) / 1000)
    this.lastT = now

    if (this.mode === 'play' && this.player.phase === 'playing') {
      const input = this.readInput()
      if (this.jumpQueued) {
        input.jump = true
        this.jumpQueued = false
      }
      const events: PhysEvent[] = []
      const wasGround = this.player.onGround
      stepPlayer(this.player, this.world, input, dt, events)
      if (input.jump && wasGround && this.player.vy > 4) {
        playSound('jump')
      }
      this.handleEvents(events)
      this.emitTouches()
      this.vm.tick(dt, this.api())
      this.followCamera(dt)
    } else {
      this.controls.update()
    }

    this.updatePlayerMesh()
    this.animateCoins(dt)

    if (this.player.message && now > this.hudMessageUntil && this.player.phase === 'playing') {
      this.player.message = ''
      this.pushHud()
    }

    this.renderer.render(this.scene, this.camera)
    this.raf = requestAnimationFrame(this.loop)
  }

  private readInput(): InputState {
    const k = this.keys
    return {
      forward: k.has('w') || k.has('arrowup') || this.virtual.forward,
      back: k.has('s') || k.has('arrowdown') || this.virtual.back,
      left: k.has('a') || k.has('arrowleft') || this.virtual.left,
      right: k.has('d') || k.has('arrowright') || this.virtual.right,
      jump: k.has(' ') || this.virtual.jump,
    }
  }

  private handleEvents(events: PhysEvent[]): void {
    const api = this.api()
    for (const ev of events) {
      switch (ev.type) {
        case 'coin':
          if (getBlock(this.world, ev.x, ev.y, ev.z) !== 12) break
          setBlock(this.world, ev.x, ev.y, ev.z, AIR)
          this.player.coins += 1
          this.rebuild()
          this.vm.emit({ kind: 'coin' }, api)
          this.pushHud()
          break
        case 'hurt':
          this.player.hitCooldown = 1
          this.player.lives -= 1
          playSound('bump')
          this.vm.emit({ kind: 'touch', block: ev.block }, api)
          this.vm.emit({ kind: 'death' }, api)
          if (this.player.lives <= 0) {
            api.lose()
          } else {
            teleportToSpawn(this.player)
          }
          this.pushHud()
          break
        case 'fall':
          this.player.lives -= 1
          playSound('bump')
          this.vm.emit({ kind: 'death' }, api)
          if (this.player.lives <= 0) api.lose()
          else teleportToSpawn(this.player)
          this.pushHud()
          break
        case 'goal':
          this.vm.emit({ kind: 'goal' }, api)
          this.pushHud()
          break
        case 'checkpoint':
          this.player.spawnX = ev.x
          this.player.spawnY = ev.y
          this.player.spawnZ = ev.z
          this.player.message = 'Checkpoint!'
          this.hudMessageUntil = performance.now() + 1500
          playSound('pop')
          this.pushHud()
          break
        case 'bounce':
          playSound('jump')
          break
        default:
          break
      }
    }
  }

  private emitTouches(): void {
    const next = new Set<string>()
    const api = this.api()
    for (const cell of collectOverlap(this.world, this.player)) {
      if (cell.id === AIR) continue
      const key = `${cell.x}:${cell.y}:${cell.z}:${cell.id}`
      next.add(key)
      if (!this.overlapKeys.has(key)) {
        this.vm.emit({ kind: 'touch', block: cell.id }, api)
      }
    }
    this.overlapKeys = next
  }

  private followCamera(dt: number): void {
    const dist = 10
    const height = 5.5
    const yaw = this.player.yaw
    const tx = this.player.x + Math.sin(yaw) * dist
    const ty = this.player.y + height
    const tz = this.player.z + Math.cos(yaw) * dist
    const k = 1 - Math.pow(0.0008, dt)
    this.camera.position.x += (tx - this.camera.position.x) * k
    this.camera.position.y += (ty - this.camera.position.y) * k
    this.camera.position.z += (tz - this.camera.position.z) * k
    this.camera.lookAt(this.player.x, this.player.y + 0.7, this.player.z)
  }

  private updatePlayerMesh(): void {
    this.playerMesh.visible = true
    this.playerMesh.position.set(this.player.x, this.player.y + 0.4, this.player.z)
    this.playerMesh.rotation.y = this.player.yaw
    const mat = this.playerMesh.material as MeshLambertMaterial
    mat.opacity = this.mode === 'play' ? 1 : 0.55
    mat.transparent = this.mode !== 'play'
    if (this.player.hitCooldown > 0 && this.mode === 'play') {
      this.playerMesh.visible = Math.floor(performance.now() / 80) % 2 === 0
    }
  }

  private animateCoins(dt: number): void {
    const mesh = this.meshes.get(12)
    if (!mesh || this.coinCells.length === 0) return
    this.coinSpin += dt * 2.6
    for (let i = 0; i < this.coinCells.length; i++) {
      const p = this.coinCells[i]
      DUMMY.position.set(
        p.x + 0.5,
        p.y + 0.55 + Math.sin(this.coinSpin * 2 + i) * 0.12,
        p.z + 0.5,
      )
      DUMMY.rotation.set(0, this.coinSpin + i, 0)
      DUMMY.scale.setScalar(0.46)
      DUMMY.updateMatrix()
      mesh.setMatrixAt(i, DUMMY.matrix)
    }
    mesh.instanceMatrix.needsUpdate = true
  }

  private applyEdit(x: number, y: number, z: number, id: number): void {
    if (!inBounds(this.world, x, y, z)) return
    const prev = getBlock(this.world, x, y, z)
    if (prev === id) return
    if (id === SPAWN) {
      for (let i = 0; i < this.world.blocks.length; i++) {
        if (this.world.blocks[i] === SPAWN) this.world.blocks[i] = AIR
      }
    }
    setBlock(this.world, x, y, z, id)
    this.undoStack.push({ x, y, z, prev, next: id })
    if (this.undoStack.length > 80) this.undoStack.shift()
    this.rebuild()
    this.syncSpawnFromWorld()
  }

  private pick(): { hit: [number, number, number] | null; prev: [number, number, number] | null } {
    const rect = this.canvas.getBoundingClientRect()
    const ndc = new Vector2(
      ((this.pointer.x - rect.left) / rect.width) * 2 - 1,
      -((this.pointer.y - rect.top) / rect.height) * 2 + 1,
    )
    const origin = new Vector3()
    const dir = new Vector3()
    origin.setFromMatrixPosition(this.camera.matrixWorld)
    dir.set(ndc.x, ndc.y, 0.5).unproject(this.camera).sub(origin).normalize()
    return voxelRaycast(this.world, origin.x, origin.y, origin.z, dir.x, dir.y, dir.z)
  }

  private updateGhost(): void {
    if (this.mode === 'play') {
      this.ghost.visible = false
      return
    }
    const { hit, prev } = this.pick()
    const cell = this.brush === AIR ? hit : prev
    if (!cell || !inBounds(this.world, cell[0], cell[1], cell[2])) {
      this.ghost.visible = false
      return
    }
    this.ghost.visible = true
    this.ghost.position.set(cell[0] + 0.5, cell[1] + 0.5, cell[2] + 0.5)
    const def = blockDef(this.brush === AIR ? 0 : this.brush)
    const mat = this.ghost.material as MeshLambertMaterial
    mat.color.set(def.color)
  }

  private onPointerMove = (e: PointerEvent): void => {
    this.pointer.x = e.clientX
    this.pointer.y = e.clientY
    if (this.pointer.button >= 0) {
      const dx = e.clientX - this.pointer.downX
      const dy = e.clientY - this.pointer.downY
      if (Math.hypot(dx, dy) > 6) this.pointer.dragging = true
    }
    if (this.mode === 'play' && this.pointer.button === 0 && this.pointer.dragging) {
      this.player.yaw -= e.movementX * 0.008
    }
    if (this.mode !== 'play') this.updateGhost()
  }

  private onPointerDown = (e: PointerEvent): void => {
    this.pointer.button = e.button
    this.pointer.downX = e.clientX
    this.pointer.downY = e.clientY
    this.pointer.dragging = false
    this.pointer.x = e.clientX
    this.pointer.y = e.clientY
    if (e.button === 2) e.preventDefault()
  }

  private onPointerUp = (e: PointerEvent): void => {
    const dragged = this.pointer.dragging
    this.pointer.button = -1
    this.pointer.dragging = false
    if (this.mode === 'play' || dragged) return
    if (this.mode === 'code') return
    const { hit, prev } = this.pick()
    if (e.button === 2 || this.brush === AIR) {
      if (hit) this.applyEdit(hit[0], hit[1], hit[2], AIR)
      return
    }
    if (e.button === 0 && prev) {
      this.applyEdit(prev[0], prev[1], prev[2], this.brush)
    }
  }

  private onKeyDown = (e: KeyboardEvent): void => {
    const key = e.key.length === 1 ? e.key.toLowerCase() : e.key.toLowerCase()
    this.keys.add(key)
    if ([' ', 'arrowup', 'arrowdown', 'arrowleft', 'arrowright'].includes(key)) {
      e.preventDefault()
    }
    if (e.ctrlKey && key === 'z') {
      e.preventDefault()
      this.undo()
    }
    if (key >= '1' && key <= '9') {
      this.brush = Number(key)
      this.callbacks.onBrush?.(this.brush)
    }
    if (key === '0' || key === 'x') {
      this.brush = AIR
      this.callbacks.onBrush?.(AIR)
    }
  }

  private onKeyUp = (e: KeyboardEvent): void => {
    this.keys.delete(e.key.length === 1 ? e.key.toLowerCase() : e.key.toLowerCase())
  }

  private onContext = (e: Event): void => {
    e.preventDefault()
  }

  private bind(): void {
    this.canvas.addEventListener('pointermove', this.onPointerMove)
    this.canvas.addEventListener('pointerdown', this.onPointerDown)
    this.canvas.addEventListener('pointerup', this.onPointerUp)
    this.canvas.addEventListener('contextmenu', this.onContext)
    window.addEventListener('keydown', this.onKeyDown)
    window.addEventListener('keyup', this.onKeyUp)
  }

  private unbind(): void {
    this.canvas.removeEventListener('pointermove', this.onPointerMove)
    this.canvas.removeEventListener('pointerdown', this.onPointerDown)
    this.canvas.removeEventListener('pointerup', this.onPointerUp)
    this.canvas.removeEventListener('contextmenu', this.onContext)
    window.removeEventListener('keydown', this.onKeyDown)
    window.removeEventListener('keyup', this.onKeyUp)
  }

  private resize(): void {
    const parent = this.canvas.parentElement ?? this.canvas
    const w = Math.max(1, parent.clientWidth)
    const h = Math.max(1, parent.clientHeight)
    this.camera.aspect = w / h
    this.camera.updateProjectionMatrix()
    this.renderer.setSize(w, h, false)
  }

  private pushHud(): void {
    const hud: HudState = {
      mode: this.mode,
      score: this.player.score,
      coins: this.player.coins,
      lives: this.player.lives,
      message: this.player.message,
      phase: this.mode === 'play' ? this.player.phase : 'build',
    }
    const s = JSON.stringify(hud)
    if (s === this.lastHud) return
    this.lastHud = s
    this.callbacks.onHud(hud)
  }
}

function emptyInput(): InputState {
  return { forward: false, back: false, left: false, right: false, jump: false }
}
