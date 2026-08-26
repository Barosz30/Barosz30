import type { SavedGame } from './types.ts'

const KEY = 'boxel.game.v1'

export function saveGameLocal(game: SavedGame): void {
  localStorage.setItem(KEY, JSON.stringify(game))
}

export function loadGameLocal(): SavedGame | null {
  const raw = localStorage.getItem(KEY)
  if (!raw) return null
  try {
    const data = JSON.parse(raw) as SavedGame
    if (data.version !== 1 || !Array.isArray(data.blocks)) return null
    return data
  } catch {
    return null
  }
}

export function downloadGame(game: SavedGame): void {
  const blob = new Blob([JSON.stringify(game)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${game.name.replace(/\s+/g, '-').toLowerCase()}.boxel.json`
  a.click()
  URL.revokeObjectURL(url)
}

export async function readGameFile(file: File): Promise<SavedGame> {
  const text = await file.text()
  const data = JSON.parse(text) as SavedGame
  if (data.version !== 1 || !Array.isArray(data.blocks)) {
    throw new Error('Niepoprawny plik BOXEL')
  }
  return data
}
