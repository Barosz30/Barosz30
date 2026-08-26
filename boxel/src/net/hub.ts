import { colorForId, makeFriendCode, makeId, makeRoomCode, sanitizeChat, sanitizeName } from './chatFilter.ts'
import { MAX_ROOM_PLAYERS } from './protocol.ts'
import type {
  ChatLine,
  ClientMsg,
  FriendInfo,
  Identity,
  RemotePlayer,
  RoomInfo,
  ServerMsg,
} from './protocol.ts'

export interface NetSink {
  send: (msg: ServerMsg) => void
}

interface Player {
  id: string
  name: string
  friendCode: string
  color: string
  x: number
  y: number
  z: number
  yaw: number
  roomId: string | null
  lastChat: number
  lastPos: number
  sink: NetSink
}

interface Room {
  id: string
  code: string
  name: string
  hostId: string
  sizeX: number
  sizeY: number
  sizeZ: number
  blocks: number[]
}

export class GameHub {
  private players = new Map<string, Player>()
  private rooms = new Map<string, Room>()
  private byCode = new Map<string, string>()
  private friends = new Map<string, Set<string>>()
  private pending = new Map<string, Set<string>>()

  connect(sink: NetSink, msg: Extract<ClientMsg, { type: 'hello' }>): Player {
    const id = msg.id && this.looksLikeId(msg.id) ? msg.id : makeId()
    let player = this.players.get(id)
    const name = sanitizeName(msg.name)
    const friendCode =
      msg.friendCode && /^BX-[A-Z0-9]{6}$/.test(msg.friendCode)
        ? msg.friendCode
        : makeFriendCode()
    if (!player) {
      player = {
        id,
        name,
        friendCode,
        color: colorForId(id),
        x: 12.5,
        y: 4,
        z: 12.5,
        yaw: 0,
        roomId: null,
        lastChat: 0,
        lastPos: 0,
        sink,
      }
      this.players.set(id, player)
      this.byCode.set(friendCode, id)
      this.friends.set(id, this.friends.get(id) ?? new Set())
    } else {
      player.sink = sink
      player.name = name
    }
    this.flushPending(player)
    player.sink.send({
      type: 'helloOk',
      you: { id: player.id, name: player.name, friendCode: player.friendCode },
      color: player.color,
    })
    this.pushFriends(player)
    this.pushRooms(player)
    return player
  }

  disconnect(player: Player): void {
    if (player.roomId) this.leaveRoom(player)
    this.players.delete(player.id)
    this.broadcastFriendsOf(player.id)
  }

  handle(player: Player, msg: ClientMsg): void {
    switch (msg.type) {
      case 'hello':
        break
      case 'rename':
        player.name = sanitizeName(msg.name)
        this.pushFriends(player)
        this.broadcastFriendsOf(player.id)
        if (player.roomId) this.broadcastPlayers(player.roomId)
        break
      case 'addFriend':
        this.addFriend(player, msg.code)
        break
      case 'host':
        this.host(player, msg)
        break
      case 'join':
        this.join(player, msg.code)
        break
      case 'leaveRoom':
        this.leaveRoom(player)
        break
      case 'pos':
        this.pos(player, msg)
        break
      case 'place':
        this.place(player, msg)
        break
      case 'chat':
        this.chat(player, msg.text, msg.toId)
        break
      case 'invite':
        this.invite(player, msg.friendId)
        break
      default:
        break
    }
  }

  identityOf(player: Player): Identity {
    return { id: player.id, name: player.name, friendCode: player.friendCode }
  }

  private addFriend(player: Player, raw: string): void {
    const code = raw.trim().toUpperCase()
    if (code === player.friendCode) {
      player.sink.send({ type: 'error', text: 'To Twój własny kod.' })
      return
    }
    const otherId = this.byCode.get(code)
    const other = otherId ? this.players.get(otherId) : undefined
    if (other) {
      this.linkFriends(player.id, other.id)
      this.pushFriends(player)
      this.pushFriends(other)
      other.sink.send({ type: 'error', text: `${player.name} dodał Cię do znajomych!` })
      return
    }
    const wait = this.pending.get(code) ?? new Set()
    wait.add(player.id)
    this.pending.set(code, wait)
    player.sink.send({ type: 'error', text: 'Kod zapisany. Znajomy pojawi się, gdy wejdzie do BOXEL.' })
  }

  private host(
    player: Player,
    msg: Extract<ClientMsg, { type: 'host' }>,
  ): void {
    if (player.roomId) this.leaveRoom(player)
    const room: Room = {
      id: makeId(),
      code: makeRoomCode(),
      name: sanitizeName(msg.name || 'Pokój'),
      hostId: player.id,
      sizeX: msg.sizeX,
      sizeY: msg.sizeY,
      sizeZ: msg.sizeZ,
      blocks: msg.blocks.slice(),
    }
    this.rooms.set(room.id, room)
    player.roomId = room.id
    player.sink.send({
      type: 'room',
      code: room.code,
      name: room.name,
      host: true,
      sizeX: room.sizeX,
      sizeY: room.sizeY,
      sizeZ: room.sizeZ,
      blocks: room.blocks,
    })
    this.broadcastRooms()
    this.broadcastPlayers(room.id)
    this.broadcastFriendsOf(player.id)
  }

  private join(player: Player, code: string): void {
    const room = [...this.rooms.values()].find((r) => r.code === code.trim())
    if (!room) {
      player.sink.send({ type: 'error', text: 'Nie ma takiego pokoju.' })
      return
    }
    const count = this.roomPlayers(room.id).length
    if (count >= MAX_ROOM_PLAYERS) {
      player.sink.send({ type: 'error', text: 'Pokój jest pełny.' })
      return
    }
    if (player.roomId) this.leaveRoom(player)
    player.roomId = room.id
    player.sink.send({
      type: 'room',
      code: room.code,
      name: room.name,
      host: room.hostId === player.id,
      sizeX: room.sizeX,
      sizeY: room.sizeY,
      sizeZ: room.sizeZ,
      blocks: room.blocks,
    })
    this.broadcastPlayers(room.id)
    this.broadcastFriendsOf(player.id)
  }

  private leaveRoom(player: Player): void {
    const roomId = player.roomId
    if (!roomId) return
    player.roomId = null
    player.sink.send({ type: 'leftRoom' })
    const room = this.rooms.get(roomId)
    const left = this.roomPlayers(roomId)
    if (room && left.length === 0) this.rooms.delete(roomId)
    else if (room && room.hostId === player.id && left[0]) {
      room.hostId = left[0].id
    }
    if (this.rooms.has(roomId)) this.broadcastPlayers(roomId)
    this.broadcastRooms()
    this.broadcastFriendsOf(player.id)
  }

  private pos(player: Player, msg: { x: number; y: number; z: number; yaw: number }): void {
    const now = Date.now()
    if (now - player.lastPos < 40) return
    player.lastPos = now
    player.x = msg.x
    player.y = msg.y
    player.z = msg.z
    player.yaw = msg.yaw
    if (player.roomId) this.broadcastPlayers(player.roomId)
  }

  private place(player: Player, msg: { x: number; y: number; z: number; id: number }): void {
    if (!player.roomId) return
    const room = this.rooms.get(player.roomId)
    if (!room) return
    if (msg.x < 0 || msg.y < 0 || msg.z < 0) return
    if (msg.x >= room.sizeX || msg.y >= room.sizeY || msg.z >= room.sizeZ) return
    const idx = msg.x + room.sizeX * (msg.z + room.sizeZ * msg.y)
    if (idx < 0 || idx >= room.blocks.length) return
    const block = Math.max(0, Math.min(15, Math.floor(msg.id)))
    room.blocks[idx] = block
    for (const other of this.roomPlayers(room.id)) {
      if (other.id === player.id) continue
      other.sink.send({ type: 'place', x: msg.x, y: msg.y, z: msg.z, id: block, fromId: player.id })
    }
  }

  private chat(player: Player, raw: string, toId?: string): void {
    const now = Date.now()
    if (now - player.lastChat < 700) return
    const text = sanitizeChat(raw)
    if (!text) {
      player.sink.send({ type: 'error', text: 'Tej wiadomości nie można wysłać.' })
      return
    }
    player.lastChat = now
    const line: ChatLine = {
      id: makeId(),
      fromId: player.id,
      fromName: player.name,
      text,
      whisper: Boolean(toId),
      at: now,
    }
    if (toId) {
      if (!this.areFriends(player.id, toId)) {
        player.sink.send({ type: 'error', text: 'Szept tylko do znajomych.' })
        return
      }
      const other = this.players.get(toId)
      player.sink.send({ type: 'chat', line })
      other?.sink.send({ type: 'chat', line })
      return
    }
    if (!player.roomId) {
      player.sink.send({ type: 'error', text: 'Wejdź do pokoju, żeby pisać na czacie.' })
      return
    }
    for (const other of this.roomPlayers(player.roomId)) {
      other.sink.send({ type: 'chat', line })
    }
  }

  private invite(player: Player, friendId: string): void {
    if (!player.roomId) {
      player.sink.send({ type: 'error', text: 'Najpierw załóż pokój.' })
      return
    }
    if (!this.areFriends(player.id, friendId)) {
      player.sink.send({ type: 'error', text: 'Możesz zapraszać tylko znajomych.' })
      return
    }
    const room = this.rooms.get(player.roomId)
    const other = this.players.get(friendId)
    if (!room || !other) {
      player.sink.send({ type: 'error', text: 'Znajomy jest offline.' })
      return
    }
    other.sink.send({ type: 'invite', fromName: player.name, roomCode: room.code })
  }

  private roomPlayers(roomId: string): Player[] {
    return [...this.players.values()].filter((p) => p.roomId === roomId)
  }

  private broadcastPlayers(roomId: string): void {
    const list: RemotePlayer[] = this.roomPlayers(roomId).map((p) => ({
      id: p.id,
      name: p.name,
      color: p.color,
      x: p.x,
      y: p.y,
      z: p.z,
      yaw: p.yaw,
    }))
    for (const p of this.roomPlayers(roomId)) {
      p.sink.send({
        type: 'players',
        players: list.filter((r) => r.id !== p.id),
      })
    }
  }

  private friendList(player: Player): FriendInfo[] {
    const ids = this.friends.get(player.id) ?? new Set()
    const out: FriendInfo[] = []
    for (const id of ids) {
      const other = this.players.get(id)
      out.push({
        id,
        name: other?.name ?? 'Znajomy',
        friendCode: other?.friendCode ?? 'BX-??????',
        online: Boolean(other),
        roomCode: other?.roomId ? this.rooms.get(other.roomId)?.code ?? null : null,
      })
    }
    return out
  }

  private pushFriends(player: Player): void {
    player.sink.send({ type: 'friends', friends: this.friendList(player) })
  }

  private roomList(): RoomInfo[] {
    return [...this.rooms.values()].map((r) => ({
      code: r.code,
      name: r.name,
      hostName: this.players.get(r.hostId)?.name ?? '?',
      players: this.roomPlayers(r.id).length,
    }))
  }

  private pushRooms(player: Player): void {
    player.sink.send({ type: 'rooms', rooms: this.roomList() })
  }

  private broadcastRooms(): void {
    const rooms = this.roomList()
    for (const p of this.players.values()) p.sink.send({ type: 'rooms', rooms })
  }

  private broadcastFriendsOf(id: string): void {
    const me = this.players.get(id)
    if (me) this.pushFriends(me)
    for (const fid of this.friends.get(id) ?? []) {
      const f = this.players.get(fid)
      if (f) this.pushFriends(f)
    }
  }

  private linkFriends(a: string, b: string): void {
    const aa = this.friends.get(a) ?? new Set()
    const bb = this.friends.get(b) ?? new Set()
    aa.add(b)
    bb.add(a)
    this.friends.set(a, aa)
    this.friends.set(b, bb)
  }

  private areFriends(a: string, b: string): boolean {
    return this.friends.get(a)?.has(b) === true
  }

  private flushPending(player: Player): void {
    const wait = this.pending.get(player.friendCode)
    if (!wait) return
    for (const id of wait) {
      const other = this.players.get(id)
      if (!other) continue
      this.linkFriends(player.id, other.id)
      this.pushFriends(other)
    }
    this.pending.delete(player.friendCode)
    if (wait.size) this.pushFriends(player)
  }

  private looksLikeId(id: string): boolean {
    return /^[0-9a-f-]{8,}$/i.test(id)
  }
}
