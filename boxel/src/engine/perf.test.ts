import { describe, expect, it } from 'vitest'
import { GRASS } from './blocks.ts'
import { createPlayer, stepPlayer } from './physics.ts'
import type { InputState } from './types.ts'
import { createWorld, fillBox } from './world.ts'

describe('physics speed', () => {
  it('steps 2000 frames of a standing player quickly', () => {
    const world = createWorld()
    fillBox(world, 0, 0, 0, 12, 0, 12, GRASS)
    const player = createPlayer()
    player.x = 6.5
    player.y = 1
    player.z = 6.5
    const still: InputState = {
      forward: false,
      back: false,
      left: false,
      right: false,
      jump: false,
    }
    const t0 = performance.now()
    for (let i = 0; i < 2000; i++) stepPlayer(player, world, still, 0.016, [])
    const ms = performance.now() - t0
    expect(ms).toBeLessThan(250)
    expect(player.onGround).toBe(true)
  })
})
