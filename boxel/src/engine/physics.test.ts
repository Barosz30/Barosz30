import { describe, expect, it } from 'vitest'
import { COIN, GRASS, LAVA } from './blocks.ts'
import { createPlayer, stepPlayer } from './physics.ts'
import { loadTemplate } from './templates.ts'
import type { InputState } from './types.ts'
import { createWorld, fillBox, getBlock, setBlock } from './world.ts'

const still: InputState = {
  forward: false,
  back: false,
  left: false,
  right: false,
  jump: false,
}

describe('physics', () => {
  it('stands on grass instead of falling through', () => {
    const world = createWorld()
    fillBox(world, 0, 0, 0, 10, 0, 10, GRASS)
    const player = createPlayer()
    player.x = 5.5
    player.y = 1
    player.z = 5.5
    for (let i = 0; i < 40; i++) {
      stepPlayer(player, world, still, 0.016, [])
    }
    expect(player.y).toBeGreaterThan(0.95)
    expect(player.y).toBeLessThan(1.2)
    expect(player.onGround).toBe(true)
  })

  it('emits a coin pickup when overlapping a coin', () => {
    const world = createWorld()
    fillBox(world, 0, 0, 0, 10, 0, 10, GRASS)
    setBlock(world, 5, 1, 5, COIN)
    const player = createPlayer()
    player.x = 5.5
    player.y = 1
    player.z = 5.5
    const events: { type: string }[] = []
    stepPlayer(player, world, still, 0.016, events)
    expect(events.some((e) => e.type === 'coin')).toBe(true)
  })
})

describe('templates', () => {
  it('parkour has spawn, lava, coins and a goal', () => {
    const { world, scripts, name } = loadTemplate('parkour')
    expect(name).toContain('Parkour')
    const counts = { spawn: 0, lava: 0, coin: 0, goal: 0 }
    for (let i = 0; i < world.blocks.length; i++) {
      const id = world.blocks[i]
      if (id === 14) counts.spawn += 1
      if (id === LAVA) counts.lava += 1
      if (id === COIN) counts.coin += 1
      if (id === 15) counts.goal += 1
    }
    expect(counts.spawn).toBeGreaterThan(0)
    expect(counts.lava).toBeGreaterThan(10)
    expect(counts.coin).toBeGreaterThanOrEqual(3)
    expect(counts.goal).toBe(1)
    expect(getBlock(world, 3, 2, 3)).toBe(14)
    expect(scripts[0]?.blocks[0]?.kind).toBe('hatStart')
  })
})
