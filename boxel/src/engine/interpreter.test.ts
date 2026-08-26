import { describe, expect, it } from 'vitest'
import { ScriptVM } from './interpreter.ts'
import { makeNode } from './world.ts'
import type { GameApi, ScriptStack } from './types.ts'

function mockApi() {
  let score = 0
  let stopped = false
  let message = ''
  let won = false
  let lost = false
  let teleports = 0
  const sounds: string[] = []
  const api: GameApi & { snapshot: () => object } = {
    addScore: (n) => {
      score += n
    },
    getScore: () => score,
    setMessage: (text) => {
      message = text
    },
    win: () => {
      won = true
      stopped = true
    },
    lose: () => {
      lost = true
      stopped = true
    },
    teleportSpawn: () => {
      teleports += 1
    },
    playSound: (name) => {
      sounds.push(name)
    },
    isStopped: () => stopped,
    snapshot: () => ({ score, stopped, message, won, lost, teleports, sounds }),
  }
  return api
}

function stack(...nodes: ScriptStack['blocks']): ScriptStack[] {
  return [{ id: 's1', blocks: nodes }]
}

describe('interpreter', () => {
  it('adds score on start', () => {
    const api = mockApi()
    const vm = new ScriptVM()
    vm.setScripts(stack(makeNode('hatStart'), makeNode('addScore', { amount: 10 })))
    vm.emit({ kind: 'start' }, api)
    vm.tick(0.016, api)
    expect(api.getScore()).toBe(10)
  })

  it('runs ifScore body only when threshold is met', () => {
    const api = mockApi()
    const vm = new ScriptVM()
    vm.setScripts(
      stack(
        makeNode('hatGoal'),
        makeNode('ifScore', {
          gte: 30,
          body: [makeNode('win')],
        }),
        makeNode('setMessage', { text: 'za mało' }),
      ),
    )
    vm.emit({ kind: 'goal' }, api)
    vm.tick(0.016, api)
    const first = api.snapshot() as { won: boolean; message: string }
    expect(first.won).toBe(false)
    expect(first.message).toBe('za mało')

    api.addScore(30)
    vm.emit({ kind: 'goal' }, api)
    vm.tick(0.016, api)
    expect((api.snapshot() as { won: boolean }).won).toBe(true)
  })

  it('caps repeat loops', () => {
    const api = mockApi()
    const vm = new ScriptVM()
    vm.setScripts(
      stack(
        makeNode('hatStart'),
        makeNode('repeat', {
          times: 50,
          body: [makeNode('addScore', { amount: 1 })],
        }),
      ),
    )
    vm.emit({ kind: 'start' }, api)
    vm.tick(0.016, api)
    expect(api.getScore()).toBe(20)
  })

  it('respects instruction budget instead of hanging', () => {
    const api = mockApi()
    const vm = new ScriptVM()
    vm.setScripts(
      stack(
        makeNode('hatAlways'),
        makeNode('repeat', { times: 20, body: [makeNode('addScore', { amount: 1 })] }),
      ),
    )
    vm.tick(0.016, api)
    expect(api.getScore()).toBeGreaterThan(0)
    expect(api.getScore()).toBeLessThanOrEqual(80)
  })

  it('matches touch events by block id', () => {
    const api = mockApi()
    const vm = new ScriptVM()
    vm.setScripts(stack(makeNode('hatTouch', { block: 8 }), makeNode('addScore', { amount: 5 })))
    vm.emit({ kind: 'touch', block: 3 }, api)
    vm.tick(0.016, api)
    expect(api.getScore()).toBe(0)
    vm.emit({ kind: 'touch', block: 8 }, api)
    vm.tick(0.016, api)
    expect(api.getScore()).toBe(5)
  })
})
