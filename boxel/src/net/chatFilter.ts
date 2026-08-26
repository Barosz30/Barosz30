import { MAX_CHAT } from './protocol.ts'

const URL_RE = /(https?:\/\/|www\.|discord\.gg|discord\.com\/invite|t\.me\/|\.ru\/|\.onion)/i
const PHONE_RE = /\+?\d[\d\s-]{7,}\d/
const BANNED = [
  'kurwa',
  'chuj',
  'cipa',
  'fuck',
  'shit',
  'nigger',
  'hitler',
  'pedo',
  'nazi',
  'zginiesz',
  'zabije',
  'zabiję',
]

export const PRESETS = ['Cześć!', 'Chodź tu!', 'Pomocy!', 'Super!', 'Czekaj!', 'Gramy!'] as const

export function sanitizeName(raw: string): string {
  const name = raw.replace(/[^\p{L}\p{N} _-]/gu, '').trim().slice(0, 16)
  return name.length >= 2 ? name : 'Gracz'
}

export function sanitizeChat(raw: string): string | null {
  const text = raw.replace(/\s+/g, ' ').trim().slice(0, MAX_CHAT)
  if (text.length < 1) return null
  if (URL_RE.test(text) || PHONE_RE.test(text)) return null
  const lower = text.toLowerCase()
  if (BANNED.some((w) => lower.includes(w))) return null
  return text
}

export function makeFriendCode(): string {
  const alphabet = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'
  let out = 'BX-'
  for (let i = 0; i < 6; i++) out += alphabet[Math.floor(Math.random() * alphabet.length)]
  return out
}

export function makeRoomCode(): string {
  return String(1000 + Math.floor(Math.random() * 9000))
}

export function makeId(): string {
  return crypto.randomUUID()
}

export function colorForId(id: string): string {
  const palette = ['#ff6b3d', '#5ec8ff', '#7cff6b', '#ffd24a', '#ff4fa3', '#c45c4a', '#9fe7ff', '#c48a3a']
  let h = 0
  for (let i = 0; i < id.length; i++) h = (h * 31 + id.charCodeAt(i)) >>> 0
  return palette[h % palette.length] ?? '#ff6b3d'
}
