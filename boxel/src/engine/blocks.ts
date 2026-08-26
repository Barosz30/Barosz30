import type { BlockId } from './types.ts'

export const AIR = 0 as const
export const GRASS = 1 as const
export const DIRT = 2 as const
export const STONE = 3 as const
export const BRICK = 4 as const
export const WOOD = 5 as const
export const SAND = 6 as const
export const WATER = 7 as const
export const LAVA = 8 as const
export const ICE = 9 as const
export const BOUNCE = 10 as const
export const SPIKE = 11 as const
export const COIN = 12 as const
export const CHECKPOINT = 13 as const
export const SPAWN = 14 as const
export const GOAL = 15 as const

export interface BlockDef {
  id: BlockId
  name: string
  hint: string
  color: string
  dark: string
  light: string
  solid: boolean
  hotbar: boolean
}

export const BLOCKS: BlockDef[] = [
  {
    id: 0,
    name: 'Gumka',
    hint: 'Usuwa klocki',
    color: '#3d3354',
    dark: '#241b36',
    light: '#6a5a86',
    solid: false,
    hotbar: true,
  },
  {
    id: 1,
    name: 'Trawa',
    hint: 'Zwykły klocek',
    color: '#5dad3e',
    dark: '#3e7a28',
    light: '#8ad45e',
    solid: true,
    hotbar: true,
  },
  {
    id: 2,
    name: 'Ziemia',
    hint: 'Zwykły klocek',
    color: '#8b5a2b',
    dark: '#5e3b1b',
    light: '#c4894f',
    solid: true,
    hotbar: true,
  },
  {
    id: 3,
    name: 'Kamień',
    hint: 'Zwykły klocek',
    color: '#8a8f99',
    dark: '#5c616b',
    light: '#c5cad3',
    solid: true,
    hotbar: true,
  },
  {
    id: 4,
    name: 'Cegła',
    hint: 'Zwykły klocek',
    color: '#c45c4a',
    dark: '#8a3328',
    light: '#e08978',
    solid: true,
    hotbar: true,
  },
  {
    id: 5,
    name: 'Drewno',
    hint: 'Zwykły klocek',
    color: '#c48a3a',
    dark: '#8a5a1e',
    light: '#e0b56a',
    solid: true,
    hotbar: true,
  },
  {
    id: 6,
    name: 'Piasek',
    hint: 'Zwykły klocek',
    color: '#e1c16e',
    dark: '#b3903e',
    light: '#ffe9a8',
    solid: true,
    hotbar: true,
  },
  {
    id: 7,
    name: 'Woda',
    hint: 'Spowalnia',
    color: '#3a9ad9',
    dark: '#1d6ea3',
    light: '#7ecfff',
    solid: false,
    hotbar: true,
  },
  {
    id: 8,
    name: 'Lawa',
    hint: 'Odejmuje życie',
    color: '#ff5a18',
    dark: '#b32600',
    light: '#ffb347',
    solid: true,
    hotbar: true,
  },
  {
    id: 9,
    name: 'Lód',
    hint: 'Ślisko',
    color: '#9fe7ff',
    dark: '#4fa9c8',
    light: '#e8fbff',
    solid: true,
    hotbar: true,
  },
  {
    id: 10,
    name: 'Sprężyna',
    hint: 'Wyrzuca w górę',
    color: '#ff4fa3',
    dark: '#b21f6c',
    light: '#ff9fd0',
    solid: true,
    hotbar: true,
  },
  {
    id: 11,
    name: 'Kolce',
    hint: 'Odejmuje życie',
    color: '#3a3148',
    dark: '#1a1424',
    light: '#6a5a80',
    solid: true,
    hotbar: true,
  },
  {
    id: 12,
    name: 'Moneta',
    hint: 'Do zbierania',
    color: '#ffd24a',
    dark: '#c79212',
    light: '#fff1a3',
    solid: false,
    hotbar: true,
  },
  {
    id: 13,
    name: 'Checkpoint',
    hint: 'Nowy start po upadku',
    color: '#7cff6b',
    dark: '#2f9e2a',
    light: '#c8ffb8',
    solid: true,
    hotbar: true,
  },
  {
    id: 14,
    name: 'Start',
    hint: 'Tu pojawia się gracz',
    color: '#5ec8ff',
    dark: '#2178aa',
    light: '#b7e9ff',
    solid: true,
    hotbar: true,
  },
  {
    id: 15,
    name: 'Meta',
    hint: 'Koniec poziomu',
    color: '#fff36a',
    dark: '#c8b020',
    light: '#fff7b0',
    solid: true,
    hotbar: true,
  },
]

const BY_ID = new Map<BlockId, BlockDef>(BLOCKS.map((b) => [b.id, b]))

export function blockDef(id: number): BlockDef {
  return BY_ID.get(id as BlockId) ?? BLOCKS[0]
}

export function isSolid(id: number): boolean {
  if (id === AIR) return false
  return blockDef(id).solid
}

export function isHazard(id: number): boolean {
  return id === LAVA || id === SPIKE
}

export const HOTBAR: BlockDef[] = BLOCKS.filter((b) => b.hotbar)
