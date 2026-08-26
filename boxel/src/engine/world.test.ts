import { describe, expect, it } from 'vitest'
import { GRASS, SPAWN } from './blocks.ts'
import {
  createWorld,
  deserializeGame,
  fillBox,
  getBlock,
  serializeGame,
  setBlock,
  worldIndex,
} from './world.ts'

describe('world', () => {
  it('sets and reads blocks', () => {
    const world = createWorld(8, 4, 8)
    expect(getBlock(world, 1, 1, 1)).toBe(0)
    setBlock(world, 1, 1, 1, GRASS)
    expect(getBlock(world, 1, 1, 1)).toBe(GRASS)
    expect(world.blocks[worldIndex(world, 1, 1, 1)]).toBe(GRASS)
  })

  it('ignores out of bounds writes', () => {
    const world = createWorld(4, 4, 4)
    expect(setBlock(world, -1, 0, 0, GRASS)).toBe(false)
    expect(getBlock(world, -1, 0, 0)).toBe(0)
  })

  it('fills a box inclusively', () => {
    const world = createWorld(8, 4, 8)
    fillBox(world, 1, 0, 1, 3, 0, 2, GRASS)
    expect(getBlock(world, 1, 0, 1)).toBe(GRASS)
    expect(getBlock(world, 3, 0, 2)).toBe(GRASS)
    expect(getBlock(world, 4, 0, 2)).toBe(0)
  })

  it('roundtrips a save', () => {
    const world = createWorld(4, 4, 4)
    setBlock(world, 0, 1, 0, SPAWN)
    const saved = serializeGame('Test', world, [])
    const loaded = deserializeGame(saved)
    expect(loaded.name).toBe('Test')
    expect(getBlock(loaded.world, 0, 1, 0)).toBe(SPAWN)
  })
})
