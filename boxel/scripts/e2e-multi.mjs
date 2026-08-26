import { WebSocket } from 'ws'

function once(ws, type) {
  return new Promise((resolve, reject) => {
    const t = setTimeout(() => reject(new Error('timeout ' + type)), 4000)
    const onMsg = (raw) => {
      const msg = JSON.parse(String(raw))
      if (msg.type === type) {
        clearTimeout(t)
        ws.off('message', onMsg)
        resolve(msg)
      }
    }
    ws.on('message', onMsg)
  })
}

function send(ws, msg) {
  ws.send(JSON.stringify(msg))
}

const a = new WebSocket('ws://127.0.0.1:8787')
const b = new WebSocket('ws://127.0.0.1:8787')
await Promise.all([
  new Promise((r) => a.once('open', r)),
  new Promise((r) => b.once('open', r)),
])
send(a, { type: 'hello', name: 'Ania' })
send(b, { type: 'hello', name: 'Kuba' })
const ha = await once(a, 'helloOk')
const hb = await once(b, 'helloOk')
send(a, { type: 'addFriend', code: hb.you.friendCode })
await once(a, 'friends')
send(a, {
  type: 'host',
  name: 'Test',
  sizeX: 2,
  sizeY: 2,
  sizeZ: 2,
  blocks: [0, 0, 0, 0, 0, 0, 0, 0],
})
const room = await once(a, 'room')
send(b, { type: 'join', code: room.code })
await once(b, 'room')
send(a, { type: 'chat', text: 'Cześć!' })
const chat = await once(b, 'chat')
if (chat.line.text !== 'Cześć!') throw new Error('chat failed')
send(a, { type: 'place', x: 1, y: 0, z: 0, id: 4 })
const place = await once(b, 'place')
if (place.id !== 4) throw new Error('place failed')
a.close()
b.close()
console.log('e2e ok', room.code, ha.you.friendCode, hb.you.friendCode)
