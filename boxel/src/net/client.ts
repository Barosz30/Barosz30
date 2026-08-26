import { makeFriendCode, makeId, sanitizeName } from './chatFilter.ts'
import type { ClientMsg, Identity, ServerMsg } from './protocol.ts'

const KEY = 'boxel.profile.v1'

export function loadProfile(): Identity {
  try {
    const raw = localStorage.getItem(KEY)
    if (raw) {
      const data = JSON.parse(raw) as Identity
      if (data.id && data.friendCode) {
        return { ...data, name: sanitizeName(data.name || 'Gracz') }
      }
    }
  } catch {
    /* ignore */
  }
  const profile: Identity = {
    id: makeId(),
    name: `Gracz${Math.floor(10 + Math.random() * 89)}`,
    friendCode: makeFriendCode(),
  }
  localStorage.setItem(KEY, JSON.stringify(profile))
  return profile
}

export function saveProfile(profile: Identity): void {
  localStorage.setItem(KEY, JSON.stringify(profile))
}

export function serverUrl(): string {
  const injected = (window as unknown as { boxelApp?: { serverUrl?: string } }).boxelApp?.serverUrl
  if (injected) return injected.replace(/\/$/, '')
  const params = new URLSearchParams(window.location.search)
  if (params.get('ws')) return String(params.get('ws')).replace(/\/$/, '')
  const host = window.location.hostname || '127.0.0.1'
  return `ws://${host}:8787`
}

type Handler = (msg: ServerMsg) => void

export class BoxelNet {
  ws: WebSocket | null = null
  you: Identity | null = null
  color = '#ff6b3d'
  private handlers = new Set<Handler>()
  private queue: ClientMsg[] = []
  private profile: Identity
  private ping: number | null = null

  constructor(profile: Identity) {
    this.profile = profile
  }

  on(handler: Handler): () => void {
    this.handlers.add(handler)
    return () => this.handlers.delete(handler)
  }

  connect(): void {
    this.disconnect()
    const ws = new WebSocket(serverUrl())
    this.ws = ws
    ws.onopen = () => {
      this.send({
        type: 'hello',
        id: this.profile.id,
        name: this.profile.name,
        friendCode: this.profile.friendCode,
      })
      for (const msg of this.queue) ws.send(JSON.stringify(msg))
      this.queue = []
    }
    ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(String(ev.data)) as ServerMsg
        if (msg.type === 'helloOk') {
          this.you = msg.you
          this.color = msg.color
          this.profile = msg.you
          saveProfile(msg.you)
        }
        for (const h of this.handlers) h(msg)
      } catch {
        /* ignore */
      }
    }
    ws.onclose = () => {
      window.setTimeout(() => this.connect(), 1200)
    }
  }

  disconnect(): void {
    if (this.ping) window.clearInterval(this.ping)
    this.ping = null
    if (this.ws) {
      this.ws.onclose = null
      this.ws.close()
      this.ws = null
    }
  }

  send(msg: ClientMsg): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg))
      return
    }
    this.queue.push(msg)
  }
}
