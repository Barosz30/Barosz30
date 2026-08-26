import { AIR } from './blocks.ts'
import { getBlock, inBounds } from './world.ts'
import type { World } from './types.ts'

export interface VoxelHit {
  hit: [number, number, number] | null
  prev: [number, number, number] | null
}

export function voxelRaycast(
  world: World,
  ox: number,
  oy: number,
  oz: number,
  dx: number,
  dy: number,
  dz: number,
  maxDist = 40,
): VoxelHit {
  let x = Math.floor(ox)
  let y = Math.floor(oy)
  let z = Math.floor(oz)

  const stepX = dx > 0 ? 1 : dx < 0 ? -1 : 0
  const stepY = dy > 0 ? 1 : dy < 0 ? -1 : 0
  const stepZ = dz > 0 ? 1 : dz < 0 ? -1 : 0

  const tDeltaX = dx === 0 ? Infinity : Math.abs(1 / dx)
  const tDeltaY = dy === 0 ? Infinity : Math.abs(1 / dy)
  const tDeltaZ = dz === 0 ? Infinity : Math.abs(1 / dz)

  const frac = (v: number, step: number) => {
    const f = v - Math.floor(v)
    return step > 0 ? 1 - f : f
  }

  let tMaxX = stepX === 0 ? Infinity : (frac(ox, stepX) * tDeltaX)
  let tMaxY = stepY === 0 ? Infinity : (frac(oy, stepY) * tDeltaY)
  let tMaxZ = stepZ === 0 ? Infinity : (frac(oz, stepZ) * tDeltaZ)

  let prev: [number, number, number] | null = null
  let t = 0

  for (let i = 0; i < 96 && t <= maxDist; i++) {
    if (inBounds(world, x, y, z) && getBlock(world, x, y, z) !== AIR) {
      return { hit: [x, y, z], prev }
    }
    prev = [x, y, z]
    if (tMaxX < tMaxY && tMaxX < tMaxZ) {
      t = tMaxX
      x += stepX
      tMaxX += tDeltaX
    } else if (tMaxY < tMaxZ) {
      t = tMaxY
      y += stepY
      tMaxY += tDeltaY
    } else {
      t = tMaxZ
      z += stepZ
      tMaxZ += tDeltaZ
    }
  }

  return { hit: null, prev: null }
}
