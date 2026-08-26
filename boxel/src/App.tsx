import { useEffect, useRef, useState } from 'react'
import { BoxelRuntime } from './engine/runtime.ts'
import { downloadGame, loadGameLocal, readGameFile, saveGameLocal } from './engine/storage.ts'
import { loadTemplate } from './engine/templates.ts'
import type { BlockId, HudState, ScriptStack, StudioMode } from './engine/types.ts'
import { deserializeGame, serializeGame } from './engine/world.ts'
import { BoxelNet, loadProfile, saveProfile } from './net/client.ts'
import { sanitizeName } from './net/chatFilter.ts'
import type { ChatLine, FriendInfo, Identity, RoomInfo } from './net/protocol.ts'
import { BlockEditor } from './ui/BlockEditor.tsx'
import { Hud } from './ui/Hud.tsx'
import { Palette } from './ui/Palette.tsx'
import { ChatDock, SocialPanel } from './ui/Social.tsx'
import { Toolbar } from './ui/Toolbar.tsx'

const EMPTY_HUD: HudState = {
  mode: 'build',
  score: 0,
  coins: 0,
  lives: 3,
  message: '',
  phase: 'build',
}

export default function App() {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const runtimeRef = useRef<BoxelRuntime | null>(null)
  const scriptsRef = useRef<ScriptStack[]>([])
  const netRef = useRef<BoxelNet | null>(null)
  const [hud, setHud] = useState<HudState>(EMPTY_HUD)
  const [mode, setMode] = useState<StudioMode>('build')
  const [brush, setBrush] = useState(1)
  const [scripts, setScripts] = useState<ScriptStack[]>([])
  const [name, setName] = useState('Parkour monetkowy')
  const [help, setHelp] = useState(false)
  const [toast, setToast] = useState('')
  const [social, setSocial] = useState(true)
  const [profile, setProfile] = useState<Identity>(() => loadProfile())
  const [connected, setConnected] = useState(false)
  const [friends, setFriends] = useState<FriendInfo[]>([])
  const [rooms, setRooms] = useState<RoomInfo[]>([])
  const [roomCode, setRoomCode] = useState<string | null>(null)
  const [chat, setChat] = useState<ChatLine[]>([])
  const [whisperTo, setWhisperTo] = useState<string | null>(null)

  const flash = (text: string) => {
    setToast(text)
    window.setTimeout(() => setToast(''), 2600)
  }

  useEffect(() => {
    scriptsRef.current = scripts
    runtimeRef.current?.setScripts(scripts)
  }, [scripts])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const packed = loadTemplate('parkour')
    const runtime = new BoxelRuntime(canvas, packed.world, {
      onHud: setHud,
      onBrush: (id) => setBrush(id),
      onEdit: (x, y, z, id) => netRef.current?.send({ type: 'place', x, y, z, id }),
      onPos: (x, y, z, yaw) => netRef.current?.send({ type: 'pos', x, y, z, yaw }),
    })
    runtime.setScripts(packed.scripts)
    runtimeRef.current = runtime
    setScripts(packed.scripts)
    setName(packed.name)
    return () => {
      runtime.dispose()
      runtimeRef.current = null
    }
  }, [])

  useEffect(() => {
    const net = new BoxelNet(loadProfile())
    netRef.current = net
    const off = net.on((msg) => {
      switch (msg.type) {
        case 'helloOk':
          setProfile(msg.you)
          setConnected(true)
          break
        case 'friends':
          setFriends(msg.friends)
          break
        case 'rooms':
          setRooms(msg.rooms)
          break
        case 'room':
          setRoomCode(msg.code)
          runtimeRef.current?.applyWorldData(msg.sizeX, msg.sizeY, msg.sizeZ, msg.blocks)
          flash(`Pokój ${msg.code}`)
          break
        case 'players':
          runtimeRef.current?.setRemotes(msg.players)
          break
        case 'place':
          runtimeRef.current?.applyNetworkEdit(msg.x, msg.y, msg.z, msg.id)
          break
        case 'chat':
          setChat((prev) => [...prev.slice(-40), msg.line])
          break
        case 'invite':
          flash(`${msg.fromName} zaprasza do pokoju ${msg.roomCode}`)
          break
        case 'leftRoom':
          setRoomCode(null)
          runtimeRef.current?.setRemotes([])
          break
        case 'error':
          flash(msg.text)
          break
        default:
          break
      }
    })
    net.connect()
    return () => {
      off()
      net.disconnect()
      netRef.current = null
    }
  }, [])

  const changeMode = (next: StudioMode) => {
    const rt = runtimeRef.current
    if (!rt) return
    rt.setScripts(scriptsRef.current)
    rt.setMode(next)
    setMode(next)
  }

  const changeBrush = (id: BlockId) => {
    setBrush(id)
    runtimeRef.current?.setBrush(id)
  }

  const applyTemplate = (id: string) => {
    const packed = loadTemplate(id)
    runtimeRef.current?.stop('build')
    runtimeRef.current?.loadWorld(packed.world)
    runtimeRef.current?.setScripts(packed.scripts)
    setScripts(packed.scripts)
    setName(packed.name)
    setMode('build')
    flash('Wczytano szablon')
  }

  const currentSave = () => {
    const rt = runtimeRef.current
    if (!rt) throw new Error('brak runtime')
    return serializeGame(name, rt.world, scriptsRef.current)
  }

  const loadSaved = (data: ReturnType<typeof currentSave>) => {
    const packed = deserializeGame(data)
    runtimeRef.current?.stop('build')
    runtimeRef.current?.loadWorld(packed.world)
    runtimeRef.current?.setScripts(packed.scripts)
    setScripts(packed.scripts)
    setName(packed.name)
    setMode('build')
  }

  const hostRoom = () => {
    const rt = runtimeRef.current
    if (!rt) return
    netRef.current?.send({
      type: 'host',
      name,
      sizeX: rt.world.sizeX,
      sizeY: rt.world.sizeY,
      sizeZ: rt.world.sizeZ,
      blocks: Array.from(rt.world.blocks),
    })
  }

  const whisperName = friends.find((f) => f.id === whisperTo)?.name ?? null

  return (
    <div className={`app mode-${mode} ${social ? 'with-social' : ''}`}>
      <Toolbar
        mode={mode}
        name={name}
        onName={setName}
        onMode={changeMode}
        onTemplate={applyTemplate}
        onSave={() => {
          saveGameLocal(currentSave())
          flash('Zapisano na tym komputerze')
        }}
        onLoad={() => {
          const data = loadGameLocal()
          if (!data) {
            flash('Brak zapisu')
            return
          }
          loadSaved(data)
          flash('Wczytano zapis')
        }}
        onDownload={() => downloadGame(currentSave())}
        onOpenFile={(file) => {
          void readGameFile(file)
            .then((data) => {
              loadSaved(data)
              flash('Otwarto plik')
            })
            .catch(() => flash('Zły plik'))
        }}
        onUndo={() => runtimeRef.current?.undo()}
        onHelp={() => setHelp(true)}
        onSocial={() => setSocial((v) => !v)}
        social={social}
        roomCode={roomCode}
        fps={hud.debug}
      />

      <div className="stage">
        {mode === 'build' ? <Palette brush={brush} onBrush={changeBrush} /> : null}
        <div className="viewport">
          <canvas ref={canvasRef} tabIndex={0} data-testid="viewport" />
          <Hud
            hud={hud}
            playing={mode === 'play'}
            onInput={(partial) => runtimeRef.current?.setVirtualInput(partial)}
            onJump={() => runtimeRef.current?.queueJump()}
            onReplay={() => changeMode('play')}
            onStop={() => changeMode('build')}
          />
        </div>
        {social ? (
          <SocialPanel
            profile={profile}
            connected={connected}
            friends={friends}
            rooms={rooms}
            roomCode={roomCode}
            onRename={(next) => {
              const name = sanitizeName(next)
              const updated = { ...profile, name }
              setProfile(updated)
              saveProfile(updated)
              netRef.current?.send({ type: 'rename', name })
            }}
            onAddFriend={(code) => netRef.current?.send({ type: 'addFriend', code })}
            onHost={hostRoom}
            onJoin={(code) => netRef.current?.send({ type: 'join', code })}
            onLeave={() => netRef.current?.send({ type: 'leaveRoom' })}
            onInvite={(friendId) => netRef.current?.send({ type: 'invite', friendId })}
            onWhisper={setWhisperTo}
          />
        ) : null}
      </div>

      {mode === 'code' ? (
        <BlockEditor scripts={scripts} onChange={setScripts} />
      ) : (
        <ChatDock
          lines={chat}
          whisperTo={whisperTo}
          whisperName={whisperName}
          onSend={(text) =>
            netRef.current?.send({
              type: 'chat',
              text,
              toId: whisperTo ?? undefined,
            })
          }
          onClearWhisper={() => setWhisperTo(null)}
        />
      )}

      {toast ? <div className="toast">{toast}</div> : null}

      {help ? (
        <div className="modal" role="dialog" aria-labelledby="help-title">
          <div className="modal-card">
            <h2 id="help-title">Jak grać w BOXEL</h2>
            <ol>
              <li>
                <b>Buduj</b> w oknie aplikacji — klocki stawiasz myszą.
              </li>
              <li>
                <b>Znajomi</b> — daj kolegom swój kod BX-…. Razem załóżcie pokój.
              </li>
              <li>
                <b>Czat</b> działa w pokoju. Szept tylko do znajomych. Linki są blokowane.
              </li>
              <li>
                <b>Graj</b> — WASD, spacja. Widzisz klocki przyjaciół w tym samym świecie.
              </li>
            </ol>
            <button type="button" onClick={() => setHelp(false)}>
              OK, buduję!
            </button>
          </div>
        </div>
      ) : null}
    </div>
  )
}
