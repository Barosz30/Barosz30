import {
  CanvasTexture,
  NearestFilter,
  SRGBColorSpace,
} from 'three'
import { BLOCKS, type BlockDef } from './blocks.ts'

function paint(def: BlockDef): HTMLCanvasElement {
  const c = document.createElement('canvas')
  c.width = 16
  c.height = 16
  const ctx = c.getContext('2d')!
  ctx.fillStyle = def.color
  ctx.fillRect(0, 0, 16, 16)

  // OG pixel noise
  for (let y = 0; y < 16; y++) {
    for (let x = 0; x < 16; x++) {
      const n = ((x * 13 + y * 7 + def.id * 31) % 5) - 2
      if (n === 0) continue
      ctx.fillStyle = n > 0 ? def.light : def.dark
      ctx.globalAlpha = 0.22
      ctx.fillRect(x, y, 1, 1)
    }
  }
  ctx.globalAlpha = 1

  ctx.fillStyle = def.light
  ctx.fillRect(1, 1, 14, 2)
  ctx.fillStyle = def.dark
  ctx.fillRect(1, 14, 14, 1)
  ctx.fillRect(14, 2, 1, 12)

  ctx.strokeStyle = '#140c1c'
  ctx.lineWidth = 1
  ctx.strokeRect(0.5, 0.5, 15, 15)

  if (def.id === 12) {
    ctx.fillStyle = def.color
    ctx.fillRect(0, 0, 16, 16)
    ctx.fillStyle = def.light
    ctx.beginPath()
    ctx.arc(8, 8, 5, 0, Math.PI * 2)
    ctx.fill()
    ctx.fillStyle = def.dark
    ctx.fillRect(7, 4, 2, 8)
    ctx.strokeStyle = '#140c1c'
    ctx.strokeRect(0.5, 0.5, 15, 15)
  }

  if (def.id === 11) {
    ctx.fillStyle = def.dark
    ctx.fillRect(0, 8, 16, 8)
    ctx.fillStyle = '#d8d0e0'
    for (let i = 0; i < 4; i++) {
      ctx.beginPath()
      ctx.moveTo(2 + i * 4, 14)
      ctx.lineTo(4 + i * 4, 2)
      ctx.lineTo(6 + i * 4, 14)
      ctx.fill()
    }
  }

  return c
}

const cache = new Map<number, CanvasTexture>()

export function textureFor(id: number): CanvasTexture {
  const hit = cache.get(id)
  if (hit) return hit
  const def = BLOCKS.find((b) => b.id === id) ?? BLOCKS[1]
  const tex = new CanvasTexture(paint(def))
  tex.magFilter = NearestFilter
  tex.minFilter = NearestFilter
  tex.colorSpace = SRGBColorSpace
  tex.needsUpdate = true
  cache.set(id, tex)
  return tex
}

export function disposeTextures(): void {
  for (const tex of cache.values()) tex.dispose()
  cache.clear()
}
