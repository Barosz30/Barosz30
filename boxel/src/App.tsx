import { useEffect, useRef, useState } from 'react'
import { BoxelRuntime } from './engine/runtime.ts'
import { downloadGame, loadGameLocal, readGameFile, saveGameLocal } from './engine/storage.ts'
import { loadTemplate } from './engine/templates.ts'
import type { BlockId, HudState, ScriptStack, StudioMode } from './engine/types.ts'
import { deserializeGame, serializeGame } from './engine/world.ts'
import { BlockEditor } from './ui/BlockEditor.tsx'
import { Hud } from './ui/Hud.tsx'
import { Palette } from './ui/Palette.tsx'
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
  const [hud, setHud] = useState<HudState>(EMPTY_HUD)
  const [mode, setMode] = useState<StudioMode>('build')
  const [brush, setBrush] = useState(1)
  const [scripts, setScripts] = useState<ScriptStack[]>([])
  const [name, setName] = useState('Parkour monetkowy')
  const [help, setHelp] = useState(false)
  const [toast, setToast] = useState('')

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

  const flash = (text: string) => {
    setToast(text)
    window.setTimeout(() => setToast(''), 1800)
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

  return (
    <div className={`app mode-${mode}`}>
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
      </div>

      {mode === 'code' ? (
        <BlockEditor scripts={scripts} onChange={setScripts} />
      ) : null}

      {toast ? <div className="toast">{toast}</div> : null}

      {help ? (
        <div className="modal" role="dialog" aria-labelledby="help-title">
          <div className="modal-card">
            <h2 id="help-title">Jak grać w BOXEL</h2>
            <ol>
              <li>
                <b>Buduj</b> — klikaj w świat, żeby stawiać klocki. PPM kasuje.
              </li>
              <li>
                <b>Klocki</b> — układaj skrypty jak w Scratchu: zdarzenie, potem akcje.
              </li>
              <li>
                <b>Graj</b> — WASD albo strzałki, spacja skacze. Mysz obraca kamerę.
              </li>
              <li>Zbierz monety, dobiegnij do złotej mety. Lawa i kolce zabierają życia.</li>
            </ol>
            <p>
              Szablon <b>Parkour monetkowy</b> ma już gotowe skrypty — otwórz zakładkę Klocki i
              zobacz, jak działają.
            </p>
            <button type="button" onClick={() => setHelp(false)}>
              OK, buduję!
            </button>
          </div>
        </div>
      ) : null}
    </div>
  )
}
