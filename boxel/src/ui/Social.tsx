import { PRESETS } from '../net/chatFilter.ts'
import type { ChatLine, FriendInfo, Identity, RoomInfo } from '../net/protocol.ts'

interface SocialPanelProps {
  profile: Identity
  connected: boolean
  friends: FriendInfo[]
  rooms: RoomInfo[]
  roomCode: string | null
  onRename: (name: string) => void
  onAddFriend: (code: string) => void
  onHost: () => void
  onJoin: (code: string) => void
  onLeave: () => void
  onInvite: (friendId: string) => void
  onWhisper: (friendId: string) => void
}

export function SocialPanel({
  profile,
  connected,
  friends,
  rooms,
  roomCode,
  onRename,
  onAddFriend,
  onHost,
  onJoin,
  onLeave,
  onInvite,
  onWhisper,
}: SocialPanelProps) {
  return (
    <aside className="social" aria-label="Znajomi">
      <h2>Znajomi</h2>
      <p className={`net-dot ${connected ? 'on' : 'off'}`}>
        {connected ? 'W aplikacji / online' : 'Łączenie…'}
      </p>
      <label className="social-field">
        <span>Imię</span>
        <input
          defaultValue={profile.name}
          maxLength={16}
          onBlur={(e) => onRename(e.target.value)}
          data-testid="profile-name"
        />
      </label>
      <div className="friend-code">
        <span>Twój kod</span>
        <b data-testid="friend-code">{profile.friendCode}</b>
      </div>
      <form
        className="add-friend"
        onSubmit={(e) => {
          e.preventDefault()
          const input = e.currentTarget.elements.namedItem('code') as HTMLInputElement
          onAddFriend(input.value)
          input.value = ''
        }}
      >
        <input name="code" placeholder="BX-XXXXXX" data-testid="add-friend-code" />
        <button type="submit" data-testid="add-friend">
          Dodaj
        </button>
      </form>
      <ul className="friend-list" data-testid="friend-list">
        {friends.length === 0 ? <li className="muted">Dodaj znajomego kodem.</li> : null}
        {friends.map((f) => (
          <li key={f.id}>
            <div>
              <b>{f.name}</b>
              <small>
                {f.online ? 'online' : 'offline'}
                {f.roomCode ? ` · pokój ${f.roomCode}` : ''}
              </small>
            </div>
            <div className="friend-actions">
              <button type="button" onClick={() => onWhisper(f.id)} disabled={!f.online}>
                Czat
              </button>
              <button type="button" onClick={() => onInvite(f.id)} disabled={!f.online || !roomCode}>
                Zaproś
              </button>
            </div>
          </li>
        ))}
      </ul>

      <h2>Wspólny świat</h2>
      {roomCode ? (
        <p className="room-code">
          Pokój <b data-testid="room-code">{roomCode}</b>
          <button type="button" onClick={onLeave}>
            Wyjdź
          </button>
        </p>
      ) : (
        <div className="room-actions">
          <button type="button" onClick={onHost} data-testid="host-room">
            Załóż pokój
          </button>
          <form
            onSubmit={(e) => {
              e.preventDefault()
              const input = e.currentTarget.elements.namedItem('room') as HTMLInputElement
              onJoin(input.value)
            }}
          >
            <input name="room" placeholder="kod 1234" data-testid="join-code" />
            <button type="submit" data-testid="join-room">
              Dołącz
            </button>
          </form>
        </div>
      )}
      <ul className="room-list">
        {rooms.map((r) => (
          <li key={r.code}>
            {r.name} · {r.code} · {r.players} graczy
          </li>
        ))}
      </ul>
    </aside>
  )
}

interface ChatDockProps {
  lines: ChatLine[]
  whisperTo: string | null
  whisperName: string | null
  onSend: (text: string) => void
  onClearWhisper: () => void
}

export function ChatDock({ lines, whisperTo, whisperName, onSend, onClearWhisper }: ChatDockProps) {
  return (
    <section className="chat-dock" aria-label="Czat">
      <div className="chat-log" data-testid="chat-log">
        {lines.slice(-30).map((line) => (
          <p key={line.id} className={line.whisper ? 'whisper' : ''}>
            <b>{line.fromName}</b>
            {line.whisper ? ' (szept)' : ''}: {line.text}
          </p>
        ))}
      </div>
      <div className="chat-presets">
        {PRESETS.map((p) => (
          <button key={p} type="button" onClick={() => onSend(p)}>
            {p}
          </button>
        ))}
      </div>
      <form
        className="chat-form"
        onSubmit={(e) => {
          e.preventDefault()
          const input = e.currentTarget.elements.namedItem('msg') as HTMLInputElement
          onSend(input.value)
          input.value = ''
        }}
      >
        <input
          name="msg"
          maxLength={60}
          placeholder={whisperTo ? `Do ${whisperName}…` : 'Napisz na czacie pokoju…'}
          data-testid="chat-input"
        />
        <button type="submit" data-testid="chat-send">
          Wyślij
        </button>
        {whisperTo ? (
          <button type="button" onClick={onClearWhisper}>
            Anuluj szept
          </button>
        ) : null}
      </form>
    </section>
  )
}
