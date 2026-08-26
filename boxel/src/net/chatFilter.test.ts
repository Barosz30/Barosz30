import { describe, expect, it } from 'vitest'
import { makeFriendCode, sanitizeChat, sanitizeName } from './chatFilter.ts'

describe('chat filter', () => {
  it('allows normal kid messages', () => {
    expect(sanitizeChat('Cześć! Chodź tu')).toBe('Cześć! Chodź tu')
  })

  it('blocks links and discord', () => {
    expect(sanitizeChat('wejdź na discord.gg/xyz')).toBeNull()
    expect(sanitizeChat('http://zly-link.pl')).toBeNull()
  })

  it('blocks banned words', () => {
    expect(sanitizeChat('kurwa lawy')).toBeNull()
  })

  it('sanitizes names', () => {
    expect(sanitizeName('Ania!!!')).toBe('Ania')
    expect(sanitizeName('x')).toBe('Gracz')
  })

  it('makes BX friend codes', () => {
    expect(makeFriendCode()).toMatch(/^BX-[A-Z0-9]{6}$/)
  })
})
