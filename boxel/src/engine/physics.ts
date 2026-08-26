import { BOUNCE, CHECKPOINT, COIN, GOAL, ICE, isHazard, isSolid, WATER } from './blocks.ts'
import { getBlock, inBounds } from './world.ts'
import type { BlockId, InputState, OverlapCell, PhysEvent, PlayerState, World } from './types.ts'

const GRAVITY = 24
const WATER_GRAVITY = 8
const JUMP = 8.4
const WATER_JUMP = 4.2
const SPEED = 6.2
const AIR_SPEED = 4.4
const WATER_SPEED = 3.2
const HALF_W = 0.28
const HEIGHT = 0.88

export function playerAabb(player: PlayerState): {
  minX: number
  maxX: number
  minY: number
  maxY: number
  minZ: number
  maxZ: number
} {
  return {
    minX: player.x - HALF_W,
    maxX: player.x + HALF_W,
    minY: player.y,
    maxY: player.y + HEIGHT,
    minZ: player.z - HALF_W,
    maxZ: player.z + HALF_W,
  }
}

export function collectOverlap(
  world: World,
  player: PlayerState,
  extraY = -0.12,
): OverlapCell[] {
  const cells: OverlapCell[] = []
  const box = playerAabb(player)
  const minX = Math.floor(box.minX)
  const maxX = Math.floor(box.maxX)
  const minY = Math.floor(box.minY + extraY)
  const maxY = Math.floor(box.maxY - 0.05)
  const minZ = Math.floor(box.minZ)
  const maxZ = Math.floor(box.maxZ)
  const seen = new Set<string>()
  for (let y = minY; y <= maxY; y++) {
    for (let z = minZ; z <= maxZ; z++) {
      for (let x = minX; x <= maxX; x++) {
        if (!inBounds(world, x, y, z)) continue
        const key = `${x},${y},${z}`
        if (seen.has(key)) continue
        seen.add(key)
        cells.push({ x, y, z, id: getBlock(world, x, y, z) })
      }
    }
  }
  return cells
}

function overlapsSolid(world: World, x: number, y: number, z: number): boolean {
  const minX = Math.floor(x - HALF_W)
  const maxX = Math.floor(x + HALF_W)
  const minY = Math.floor(y + 0.001)
  const maxY = Math.floor(y + HEIGHT - 0.001)
  const minZ = Math.floor(z - HALF_W)
  const maxZ = Math.floor(z + HALF_W)
  for (let by = minY; by <= maxY; by++) {
    for (let bz = minZ; bz <= maxZ; bz++) {
      for (let bx = minX; bx <= maxX; bx++) {
        if (isSolid(getBlock(world, bx, by, bz))) return true
      }
    }
  }
  return false
}

function feetBlock(world: World, player: PlayerState): number {
  return getBlock(
    world,
    Math.floor(player.x),
    Math.floor(player.y - 0.08),
    Math.floor(player.z),
  )
}

export function createPlayer(): PlayerState {
  return {
    x: 12.5,
    y: 4,
    z: 12.5,
    vx: 0,
    vy: 0,
    vz: 0,
    yaw: 0,
    onGround: false,
    inWater: false,
    lives: 3,
    score: 0,
    coins: 0,
    spawnX: 12.5,
    spawnY: 4,
    spawnZ: 12.5,
    hitCooldown: 0,
    bounceCooldown: 0,
    goalCooldown: 0,
    message: '',
    phase: 'playing',
  }
}

export function teleportToSpawn(player: PlayerState): void {
  player.x = player.spawnX
  player.y = player.spawnY
  player.z = player.spawnZ
  player.vx = 0
  player.vy = 0
  player.vz = 0
}

export function stepPlayer(
  player: PlayerState,
  world: World,
  input: InputState,
  dt: number,
  events: PhysEvent[],
): void {
  if (player.phase !== 'playing') {
    player.vx = 0
    player.vz = 0
    player.vy = 0
    return
  }

  const clampedDt = Math.min(dt, 0.05)
  player.hitCooldown = Math.max(0, player.hitCooldown - clampedDt)
  player.bounceCooldown = Math.max(0, player.bounceCooldown - clampedDt)
  player.goalCooldown = Math.max(0, player.goalCooldown - clampedDt)

  const cells = collectOverlap(world, player, 0)
  player.inWater = cells.some((c) => c.id === WATER)

  player.vy -= (player.inWater ? WATER_GRAVITY : GRAVITY) * clampedDt

  const feet = feetBlock(world, player)
  const onIce = player.onGround && feet === ICE
  const speed = player.inWater ? WATER_SPEED : player.onGround ? SPEED : AIR_SPEED

  const sin = Math.sin(player.yaw)
  const cos = Math.cos(player.yaw)
  let wishX = 0
  let wishZ = 0
  if (input.forward) {
    wishX -= sin
    wishZ -= cos
  }
  if (input.back) {
    wishX += sin
    wishZ += cos
  }
  if (input.left) {
    wishX -= cos
    wishZ += sin
  }
  if (input.right) {
    wishX += cos
    wishZ -= sin
  }
  const wishLen = Math.hypot(wishX, wishZ)
  if (wishLen > 0) {
    wishX = (wishX / wishLen) * speed
    wishZ = (wishZ / wishLen) * speed
  }

  const accel = player.onGround ? (onIce ? 4 : 18) : 6
  player.vx += (wishX - player.vx) * Math.min(1, accel * clampedDt)
  player.vz += (wishZ - player.vz) * Math.min(1, accel * clampedDt)

  if (input.jump && (player.onGround || player.inWater)) {
    player.vy = player.inWater ? WATER_JUMP : JUMP
    player.onGround = false
  }

  moveAxis(player, world, player.vx * clampedDt, 0, 0)
  moveAxis(player, world, 0, 0, player.vz * clampedDt)
  player.onGround = false
  moveAxis(player, world, 0, player.vy * clampedDt, 0)
  if (player.vy === 0 && overlapsSolid(world, player.x, player.y - 0.04, player.z)) {
    player.onGround = true
  }

  if (player.y < -2) {
    events.push({ type: 'fall' })
    return
  }

  const after = collectOverlap(world, player)
  const seenCoin = new Set<string>()
  for (const cell of after) {
    if (cell.id === COIN) {
      const key = `${cell.x},${cell.y},${cell.z}`
      if (seenCoin.has(key)) continue
      seenCoin.add(key)
      events.push({ type: 'coin', x: cell.x, y: cell.y, z: cell.z })
    } else if (cell.id === GOAL && player.goalCooldown <= 0) {
      player.goalCooldown = 1.2
      events.push({ type: 'goal' })
    } else if (cell.id === CHECKPOINT) {
      events.push({
        type: 'checkpoint',
        x: cell.x + 0.5,
        y: cell.y + 1,
        z: cell.z + 0.5,
      })
    } else if (isHazard(cell.id) && player.hitCooldown <= 0) {
      events.push({ type: 'hurt', block: cell.id as BlockId })
    }
  }

  const land = feetBlock(world, player)
  if (player.onGround && land === BOUNCE && player.bounceCooldown <= 0) {
    player.vy = 11
    player.onGround = false
    player.bounceCooldown = 0.2
    events.push({ type: 'bounce' })
  }
}

function moveAxis(
  player: PlayerState,
  world: World,
  dx: number,
  dy: number,
  dz: number,
): void {
  player.x += dx
  player.y += dy
  player.z += dz
  if (!overlapsSolid(world, player.x, player.y, player.z)) return

  if (dx !== 0) {
    player.x -= dx
    player.vx = 0
  }
  if (dz !== 0) {
    player.z -= dz
    player.vz = 0
  }
  if (dy !== 0) {
    player.y -= dy
    if (dy < 0) player.onGround = true
    player.vy = 0
  }
}
