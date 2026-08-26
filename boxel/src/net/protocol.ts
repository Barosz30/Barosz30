export const MAX_CHAT = 60
export const MAX_ROOM_PLAYERS = 8
export const POS_HZ = 15

export interface Identity {
  id: string
  name: string
  friendCode: string
}

export interface FriendInfo {
  id: string
  name: string
  friendCode: string
  online: boolean
  roomCode: string | null
}

export interface RemotePlayer {
  id: string
  name: string
  color: string
  x: number
  y: number
  z: number
  yaw: number
}

export interface ChatLine {
  id: string
  fromId: string
  fromName: string
  text: string
  whisper: boolean
  at: number
}

export interface RoomInfo {
  code: string
  name: string
  hostName: string
  players: number
}

export type ClientMsg =
  | { type: 'hello'; id?: string; name: string; friendCode?: string }
  | { type: 'rename'; name: string }
  | { type: 'addFriend'; code: string }
  | { type: 'host'; name: string; sizeX: number; sizeY: number; sizeZ: number; blocks: number[] }
  | { type: 'join'; code: string }
  | { type: 'leaveRoom' }
  | { type: 'pos'; x: number; y: number; z: number; yaw: number }
  | { type: 'place'; x: number; y: number; z: number; id: number }
  | { type: 'chat'; text: string; toId?: string }
  | { type: 'invite'; friendId: string }

export type ServerMsg =
  | { type: 'helloOk'; you: Identity; color: string }
  | { type: 'error'; text: string }
  | { type: 'friends'; friends: FriendInfo[] }
  | { type: 'rooms'; rooms: RoomInfo[] }
  | { type: 'room'; code: string; name: string; host: boolean; sizeX: number; sizeY: number; sizeZ: number; blocks: number[] }
  | { type: 'players'; players: RemotePlayer[] }
  | { type: 'place'; x: number; y: number; z: number; id: number; fromId: string }
  | { type: 'chat'; line: ChatLine }
  | { type: 'invite'; fromName: string; roomCode: string }
  | { type: 'leftRoom' }
