import { createServer } from 'node:http'
import { WebSocketServer } from 'ws'
import { GameHub } from '../src/net/hub.ts'
import type { ClientMsg, ServerMsg } from '../src/net/protocol.ts'

const port = Number(process.env.PORT || 8787)
const host = process.env.HOST || '0.0.0.0'

const httpServer = createServer((_req, res) => {
  res.writeHead(200, { 'content-type': 'text/plain' })
  res.end('BOXEL play server\n')
})

const wss = new WebSocketServer({ server: httpServer })
const hub = new GameHub()

wss.on('connection', (socket) => {
  let player: ReturnType<GameHub['connect']> | null = null
  const sink = {
    send: (msg: ServerMsg) => {
      if (socket.readyState === socket.OPEN) socket.send(JSON.stringify(msg))
    },
  }

  socket.on('message', (raw) => {
    let msg: ClientMsg
    try {
      msg = JSON.parse(String(raw)) as ClientMsg
    } catch {
      return
    }
    if (!player) {
      if (msg.type !== 'hello') return
      player = hub.connect(sink, msg)
      return
    }
    hub.handle(player, msg)
  })

  socket.on('close', () => {
    if (player) hub.disconnect(player)
  })
})

httpServer.listen(port, host, () => {
  console.log(`BOXEL server ws://${host}:${port}`)
})
