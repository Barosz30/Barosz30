import type { CodeNode, GameApi, ScriptStack } from './types.ts'

const MAX_OPS_PER_TICK = 80
const MAX_FIBERS = 16
const MAX_REPEAT = 20
const MAX_WAIT = 5

export type ScriptEvent =
  | { kind: 'start' }
  | { kind: 'coin' }
  | { kind: 'touch'; block: number }
  | { kind: 'goal' }
  | { kind: 'death' }
  | { kind: 'always' }

interface Frame {
  blocks: CodeNode[]
  index: number
  remaining?: number
}

interface Fiber {
  stack: Frame[]
  wait: number
  fromAlways: boolean
}

function hatMatches(node: CodeNode | undefined, event: ScriptEvent): boolean {
  if (!node) return false
  switch (event.kind) {
    case 'start':
      return node.kind === 'hatStart'
    case 'coin':
      return node.kind === 'hatCoin'
    case 'goal':
      return node.kind === 'hatGoal'
    case 'death':
      return node.kind === 'hatDeath'
    case 'always':
      return node.kind === 'hatAlways'
    case 'touch':
      return node.kind === 'hatTouch' && node.block === event.block
    default:
      return false
  }
}

export class ScriptVM {
  scripts: ScriptStack[] = []
  fibers: Fiber[] = []
  private alwaysBusy = false

  reset(): void {
    this.fibers = []
    this.alwaysBusy = false
  }

  setScripts(scripts: ScriptStack[]): void {
    this.scripts = scripts
  }

  emit(event: ScriptEvent, api: GameApi): void {
    if (api.isStopped() && event.kind !== 'start') return
    if (this.fibers.length >= MAX_FIBERS) return
    for (const stack of this.scripts) {
      const hat = stack.blocks[0]
      if (!hatMatches(hat, event)) continue
      if (this.fibers.length >= MAX_FIBERS) return
      this.fibers.push({
        stack: [{ blocks: stack.blocks, index: 1 }],
        wait: 0,
        fromAlways: event.kind === 'always',
      })
    }
  }

  tick(dt: number, api: GameApi): void {
    if (api.isStopped()) {
      this.reset()
      return
    }

    if (!this.alwaysBusy) {
      this.emit({ kind: 'always' }, api)
      this.alwaysBusy = this.fibers.some((f) => f.fromAlways)
    }

    let ops = 0
    for (const fiber of this.fibers) {
      if (fiber.wait > 0) {
        fiber.wait = Math.max(0, fiber.wait - dt)
        if (fiber.wait > 0) continue
      }
      while (
        ops < MAX_OPS_PER_TICK &&
        fiber.wait <= 0 &&
        fiber.stack.length > 0 &&
        !api.isStopped()
      ) {
        ops += 1
        stepFiber(fiber, api)
      }
    }

    this.fibers = this.fibers.filter((f) => f.stack.length > 0)
    this.alwaysBusy = this.fibers.some((f) => f.fromAlways)
  }
}

function stepFiber(fiber: Fiber, api: GameApi): void {
  const frame = fiber.stack[fiber.stack.length - 1]
  if (!frame) return

  if (frame.index >= frame.blocks.length) {
    if (frame.remaining !== undefined && frame.remaining > 1) {
      frame.remaining -= 1
      frame.index = 0
      return
    }
    fiber.stack.pop()
    return
  }

  const node = frame.blocks[frame.index]
  frame.index += 1
  if (!node) return

  switch (node.kind) {
    case 'hatStart':
    case 'hatCoin':
    case 'hatTouch':
    case 'hatGoal':
    case 'hatDeath':
    case 'hatAlways':
      break
    case 'addScore':
      api.addScore(node.amount ?? 1)
      break
    case 'setMessage':
      api.setMessage(node.text ?? '')
      break
    case 'win':
      api.win()
      break
    case 'lose':
      api.lose()
      break
    case 'teleport':
      api.teleportSpawn()
      break
    case 'sound':
      api.playSound(node.sound ?? 'pop')
      break
    case 'wait': {
      if (fiber.fromAlways) break
      const seconds = Math.min(MAX_WAIT, Math.max(0, node.seconds ?? 1))
      fiber.wait = seconds
      break
    }
    case 'ifScore': {
      const need = node.gte ?? 0
      const body = node.body ?? []
      if (api.getScore() >= need && body.length > 0) {
        fiber.stack.push({ blocks: body, index: 0 })
      }
      break
    }
    case 'repeat': {
      const times = Math.min(MAX_REPEAT, Math.max(0, Math.floor(node.times ?? 1)))
      const body = node.body ?? []
      if (times > 0 && body.length > 0) {
        fiber.stack.push({ blocks: body, index: 0, remaining: times })
      }
      break
    }
    default:
      break
  }
}

export function countOpsForTest(scripts: ScriptStack[], event: ScriptEvent, api: GameApi, ticks = 3): number {
  const vm = new ScriptVM()
  vm.setScripts(scripts)
  vm.emit(event, api)
  for (let i = 0; i < ticks; i++) vm.tick(0.05, api)
  return api.getScore()
}
