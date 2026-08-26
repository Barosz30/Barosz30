import { BOUNCE, BRICK, CHECKPOINT, COIN, GOAL, GRASS, ICE, LAVA, SPAWN, SPIKE, STONE, WOOD } from './blocks.ts'
import type { ScriptStack, World } from './types.ts'
import { createWorld, fillBox, makeNode, setBlock } from './world.ts'

export interface Template {
  id: string
  name: string
  blurb: string
  build: () => { world: World; scripts: ScriptStack[]; name: string }
}

function parkourScripts(): ScriptStack[] {
  return [
    {
      id: crypto.randomUUID(),
      blocks: [
        makeNode('hatStart'),
        makeNode('setMessage', { text: 'Zbierz 3 monety i dobiegnij do złotej flagi!' }),
        makeNode('sound', { sound: 'pop' }),
      ],
    },
    {
      id: crypto.randomUUID(),
      blocks: [
        makeNode('hatCoin'),
        makeNode('addScore', { amount: 10 }),
        makeNode('sound', { sound: 'coin' }),
        makeNode('setMessage', { text: 'Moneta! +10' }),
      ],
    },
    {
      id: crypto.randomUUID(),
      blocks: [
        makeNode('hatTouch', { block: LAVA }),
        makeNode('setMessage', { text: 'Auć! Lawa parzy.' }),
        makeNode('sound', { sound: 'bump' }),
      ],
    },
    {
      id: crypto.randomUUID(),
      blocks: [
        makeNode('hatTouch', { block: SPIKE }),
        makeNode('setMessage', { text: 'Ostrożnie, kolce!' }),
        makeNode('sound', { sound: 'bump' }),
      ],
    },
    {
      id: crypto.randomUUID(),
      blocks: [
        makeNode('hatDeath'),
        makeNode('setMessage', { text: 'Spróbuj jeszcze raz!' }),
        makeNode('teleport'),
      ],
    },
    {
      id: crypto.randomUUID(),
      blocks: [
        makeNode('hatGoal'),
        makeNode('ifScore', {
          gte: 30,
          body: [
            makeNode('sound', { sound: 'win' }),
            makeNode('setMessage', { text: 'WYGRANA!' }),
            makeNode('win'),
          ],
        }),
        makeNode('setMessage', { text: 'Potrzeba 30 punktów (3 monety).' }),
      ],
    },
  ]
}

function buildSandbox(): { world: World; scripts: ScriptStack[]; name: string } {
  const world = createWorld()
  fillBox(world, 0, 0, 0, 23, 0, 23, GRASS)
  setBlock(world, 12, 1, 12, SPAWN)
  return {
    world,
    name: 'Piaskownica',
    scripts: [
      {
        id: crypto.randomUUID(),
        blocks: [
          makeNode('hatStart'),
          makeNode('setMessage', { text: 'Buduj klockami, potem naciśnij GRAJ!' }),
        ],
      },
      {
        id: crypto.randomUUID(),
        blocks: [
          makeNode('hatCoin'),
          makeNode('addScore', { amount: 10 }),
          makeNode('sound', { sound: 'coin' }),
        ],
      },
      {
        id: crypto.randomUUID(),
        blocks: [
          makeNode('hatGoal'),
          makeNode('win'),
        ],
      },
    ],
  }
}

function buildParkour(): { world: World; scripts: ScriptStack[]; name: string } {
  const world = createWorld()
  fillBox(world, 0, 0, 0, 23, 0, 23, LAVA)

  // start island
  fillBox(world, 1, 1, 1, 5, 1, 5, GRASS)
  setBlock(world, 3, 2, 3, SPAWN)

  // stone steps
  fillBox(world, 4, 2, 8, 6, 2, 10, STONE)
  setBlock(world, 5, 4, 9, COIN)

  fillBox(world, 8, 3, 12, 11, 3, 14, BRICK)
  setBlock(world, 10, 3, 13, BOUNCE)
  setBlock(world, 9, 5, 13, COIN)

  // ice run
  fillBox(world, 12, 4, 14, 18, 4, 15, ICE)
  setBlock(world, 16, 6, 14, COIN)

  // checkpoint ledge
  fillBox(world, 18, 5, 17, 21, 5, 21, WOOD)
  setBlock(world, 20, 6, 19, CHECKPOINT)
  setBlock(world, 19, 7, 20, COIN)

  // spikes as a trap on the side
  setBlock(world, 18, 6, 17, SPIKE)
  setBlock(world, 21, 6, 21, SPIKE)

  // upper bricks toward goal
  fillBox(world, 12, 7, 20, 16, 7, 22, BRICK)
  fillBox(world, 6, 8, 20, 10, 8, 22, STONE)
  fillBox(world, 2, 9, 19, 5, 9, 22, WOOD)
  setBlock(world, 3, 10, 20, GOAL)

  return {
    world,
    name: 'Parkour monetkowy',
    scripts: parkourScripts(),
  }
}

function buildArena(): { world: World; scripts: ScriptStack[]; name: string } {
  const world = createWorld()
  fillBox(world, 0, 0, 0, 23, 0, 23, STONE)
  fillBox(world, 1, 1, 1, 22, 1, 22, GRASS)
  fillBox(world, 8, 1, 8, 15, 1, 15, WOOD)
  setBlock(world, 12, 2, 12, SPAWN)
  setBlock(world, 4, 2, 4, COIN)
  setBlock(world, 19, 2, 4, COIN)
  setBlock(world, 4, 2, 19, COIN)
  setBlock(world, 19, 2, 19, COIN)
  setBlock(world, 12, 2, 20, GOAL)
  fillBox(world, 10, 2, 10, 14, 4, 10, BRICK)
  return {
    world,
    name: 'Arena monet',
    scripts: [
      {
        id: crypto.randomUUID(),
        blocks: [
          makeNode('hatStart'),
          makeNode('setMessage', { text: 'Zbierz monety w arenie!' }),
        ],
      },
      {
        id: crypto.randomUUID(),
        blocks: [
          makeNode('hatCoin'),
          makeNode('addScore', { amount: 10 }),
          makeNode('sound', { sound: 'coin' }),
        ],
      },
      {
        id: crypto.randomUUID(),
        blocks: [
          makeNode('hatGoal'),
          makeNode('ifScore', {
            gte: 40,
            body: [makeNode('win'), makeNode('sound', { sound: 'win' })],
          }),
          makeNode('setMessage', { text: 'Zbierz wszystkie 4 monety (40 pkt).' }),
        ],
      },
    ],
  }
}

export const TEMPLATES: Template[] = [
  {
    id: 'parkour',
    name: 'Parkour monetkowy',
    blurb: 'Skacz, zbieraj monety, unikaj lawy.',
    build: buildParkour,
  },
  {
    id: 'sandbox',
    name: 'Piaskownica',
    blurb: 'Płaska trawa — buduj od zera.',
    build: buildSandbox,
  },
  {
    id: 'arena',
    name: 'Arena monet',
    blurb: 'Mały dziedziniec z monetami.',
    build: buildArena,
  },
]

export function loadTemplate(id: string) {
  const t = TEMPLATES.find((x) => x.id === id) ?? TEMPLATES[0]
  return t.build()
}
