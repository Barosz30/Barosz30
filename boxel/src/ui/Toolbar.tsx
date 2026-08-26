import { TEMPLATES } from '../engine/templates.ts'
import type { StudioMode } from '../engine/types.ts'

interface ToolbarProps {
  mode: StudioMode
  name: string
  onName: (name: string) => void
  onMode: (mode: StudioMode) => void
  onTemplate: (id: string) => void
  onSave: () => void
  onLoad: () => void
  onDownload: () => void
  onOpenFile: (file: File) => void
  onUndo: () => void
  onHelp: () => void
  onSocial: () => void
  social: boolean
  roomCode: string | null
  fps?: string
}

export function Toolbar({
  mode,
  name,
  onName,
  onMode,
  onTemplate,
  onSave,
  onLoad,
  onDownload,
  onOpenFile,
  onUndo,
  onHelp,
  onSocial,
  social,
  roomCode,
  fps,
}: ToolbarProps) {
  return (
    <header className="toolbar">
      <div className="brand">
        <span className="logo-cube" aria-hidden="true" />
        <div>
          <h1>BOXEL</h1>
          <p>Aplikacja klocków {fps ? `· ${fps}` : ''}</p>
        </div>
      </div>

      <div className="mode-switch" role="tablist" aria-label="Tryb">
        <button
          type="button"
          role="tab"
          aria-selected={mode === 'build'}
          className={mode === 'build' ? 'on' : ''}
          onClick={() => onMode('build')}
          data-testid="mode-build"
        >
          Buduj
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={mode === 'code'}
          className={mode === 'code' ? 'on' : ''}
          onClick={() => onMode('code')}
          data-testid="mode-code"
        >
          Klocki
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={mode === 'play'}
          className={mode === 'play' ? 'on play' : 'play'}
          onClick={() => onMode(mode === 'play' ? 'build' : 'play')}
          data-testid="mode-play"
        >
          {mode === 'play' ? 'Stop' : 'Graj'}
        </button>
      </div>

      <label className="name-field">
        <span>Gra</span>
        <input
          value={name}
          onChange={(e) => onName(e.target.value)}
          maxLength={32}
          data-testid="game-name"
        />
      </label>

      <label className="select-field">
        <span>Szablon</span>
        <select onChange={(e) => onTemplate(e.target.value)} defaultValue="parkour">
          {TEMPLATES.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name}
            </option>
          ))}
        </select>
      </label>

      <div className="toolbar-actions">
        <button type="button" onClick={onUndo}>
          Cofnij
        </button>
        <button type="button" onClick={onSave} data-testid="save-game">
          Zapisz
        </button>
        <button type="button" onClick={onLoad}>
          Wczytaj
        </button>
        <button type="button" onClick={onDownload}>
          Plik
        </button>
        <label className="file-btn">
          Otwórz
          <input
            type="file"
            accept="application/json,.json"
            onChange={(e) => {
              const file = e.target.files?.[0]
              if (file) onOpenFile(file)
              e.target.value = ''
            }}
          />
        </label>
        <button type="button" className={social ? 'on' : ''} onClick={onSocial} data-testid="toggle-social">
          Znajomi
        </button>
        {roomCode ? <span className="room-chip">Pokój {roomCode}</span> : null}
        <button type="button" className="help-btn" onClick={onHelp} data-testid="help">
          ?
        </button>
      </div>
    </header>
  )
}
