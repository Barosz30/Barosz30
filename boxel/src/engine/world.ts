import { AIR, SPAWN } from './blocks.ts'
import type { CodeKind, CodeNode, SavedGame, ScriptStack, World } from './types.ts'

export function makeNode(
  kind: CodeKind,
  extra: Partial<CodeNode> = {},
): CodeNode {
  return { id: newId(), kind, ...extra }
}

export const SIZE_X = 24
export const SIZE_Y = 16
export const SIZE_Z = 24

export function createWorld(
  sizeX = SIZE_X,
  sizeY = SIZE_Y,
  sizeZ = SIZE_Z,
): World {
  return {
    sizeX,
    sizeY,
    sizeZ,
    blocks: new Uint8Array(sizeX * sizeY * sizeZ),
  }
}

export function worldIndex(world: World, x: number, y: number, z: number): number {
  return x + world.sizeX * (z + world.sizeZ * y)
}

export function inBounds(world: World, x: number, y: number, z: number): boolean {
  return (
    x >= 0 &&
    y >= 0 &&
    z >= 0 &&
    x < world.sizeX &&
    y < world.sizeY &&
    z < world.sizeZ
  )
}

export function getBlock(world: World, x: number, y: number, z: number): number {
  if (!inBounds(world, x, y, z)) return AIR
  return world.blocks[worldIndex(world, x, y, z)] ?? AIR
}

export function setBlock(
  world: World,
  x: number,
  y: number,
  z: number,
  id: number,
): boolean {
  if (!inBounds(world, x, y, z)) return false
  world.blocks[worldIndex(world, x, y, z)] = id
  return true
}

export function fillBox(
  world: World,
  x1: number,
  y1: number,
  z1: number,
  x2: number,
  y2: number,
  z2: number,
  id: number,
): void {
  const xa = Math.min(x1, x2)
  const xb = Math.max(x1, x2)
  const ya = Math.min(y1, y2)
  const yb = Math.max(y1, y2)
  const za = Math.min(z1, z2)
  const zb = Math.max(z1, z2)
  for (let y = ya; y <= yb; y++) {
    for (let z = za; z <= zb; z++) {
      for (let x = xa; x <= xb; x++) {
        setBlock(world, x, y, z, id)
      }
    }
  }
}

export function cloneWorld(world: World): World {
  return {
    sizeX: world.sizeX,
    sizeY: world.sizeY,
    sizeZ: world.sizeZ,
    blocks: new Uint8Array(world.blocks),
  }
}

export function copyWorld(from: World, to: World): void {
  to.sizeX = from.sizeX
  to.sizeY = from.sizeY
  to.sizeZ = from.sizeZ
  to.blocks = new Uint8Array(from.blocks)
}

export function findSpawn(world: World): { x: number; y: number; z: number } {
  for (let y = 0; y < world.sizeY; y++) {
    for (let z = 0; z < world.sizeZ; z++) {
      for (let x = 0; x < world.sizeX; x++) {
        if (getBlock(world, x, y, z) === SPAWN) {
          return { x: x + 0.5, y: y + 1, z: z + 0.5 }
        }
      }
    }
  }
  for (let y = world.sizeY - 2; y >= 0; y--) {
    for (let z = 0; z < world.sizeZ; z++) {
      for (let x = 0; x < world.sizeX; x++) {
        if (getBlock(world, x, y, z) !== AIR && getBlock(world, x, y + 1, z) === AIR) {
          return { x: x + 0.5, y: y + 1, z: z + 0.5 }
        }
      }
    }
  }
  return {
    x: world.sizeX / 2,
    y: 4,
    z: world.sizeZ / 2,
  }
}

export function serializeGame(
  name: string,
  world: World,
  scripts: ScriptStack[],
): SavedGame {
  return {
    version: 1,
    name,
    sizeX: world.sizeX,
    sizeY: world.sizeY,
    sizeZ: world.sizeZ,
    blocks: Array.from(world.blocks),
    scripts: cloneScripts(scripts),
  }
}

export function deserializeGame(data: SavedGame): { world: World; scripts: ScriptStack[]; name: string } {
  const world = createWorld(data.sizeX, data.sizeY, data.sizeZ)
  world.blocks.set(data.blocks)
  return {
    world,
    scripts: cloneScripts(data.scripts ?? []),
    name: data.name || 'Moja gra',
  }
}

export function cloneScripts(scripts: ScriptStack[]): ScriptStack[] {
  return JSON.parse(JSON.stringify(scripts)) as ScriptStack[]
}

export function newId(): string {
  return crypto.randomUUID()
}
