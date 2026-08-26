import { describe, expect, it } from 'vitest'
import { makeFriendCode, sanitizeChat, sanitizeName } from './chatFilter.ts'
import { GameHub } from './hub.ts'
import type { ServerMsg } from './protocol.ts'

function mockSink() {
  const messages: ServerMsg[] = []
  return {
    messages,
    sink: { send: (msg: ServerMsg) => messages.push(msg) },
  }
}

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

describe('hub friends and rooms', () => {
  it('adds friends by code and lets them chat in a hosted room', () => {
    const hub = new GameHub()
    const a = mockSink()
    const b = mockSink()
    const pa = hub.connect(a.sink, { type: 'hello', name: 'Ania' })
    const pb = hub.connect(b.sink, { type: 'hello', name: 'Kuba' })
    hub.handle(pa, { type: 'addFriend', code: hub.identityOf(pb).friendCode })

    const friends = a.messages.filter((m) => m.type === 'friends')
    const last = friends[friends.length - 1]
    expect(last && last.type === 'friends' && last.friends.some((f) => f.name === 'Kuba')).toBe(true)

    hub.handle(pa, {
      type: 'host',
      name: 'Parkour',
      sizeX: 2,
      sizeY: 2,
      sizeZ: 2,
      blocks: [0, 0, 0, 0, 0, 0, 0, 0],
    })
    const room = a.messages.find((m) => m.type === 'room')
    expect(room && room.type === 'room' && room.code).toMatch(/^\d{4}$/)
    const code = room && room.type === 'room' ? room.code : ''

    hub.handle(pb, { type: 'join', code })
    hub.handle(pa, { type: 'chat', text: 'Cześć!' })
    expect(b.messages.some((m) => m.type === 'chat' && m.line.text === 'Cześć!')).toBe(true)

    hub.handle(pa, { type: 'place', x: 1, y: 0, z: 0, id: 4 })
    expect(b.messages.some((m) => m.type === 'place' && m.id === 4)).toBe(true)
  })

  it('rejects whispers to non-friends', () => {
    const hub = new GameHub()
    const a = mockSink()
    const b = mockSink()
    const pa = hub.connect(a.sink, { type: 'hello', name: 'Ania' })
    const pb = hub.connect(b.sink, { type: 'hello', name: 'Kuba' })
    hub.handle(pa, { type: 'chat', text: 'psst', toId: pb.id })
    expect(a.messages.some((m) => m.type === 'error')).toBe(true)
    expect(b.messages.some((m) => m.type === 'chat')).toBe(false)
  })
})
