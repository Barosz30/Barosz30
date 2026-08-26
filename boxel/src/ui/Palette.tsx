import type { CSSProperties } from 'react'
import { HOTBAR } from '../engine/blocks.ts'
import type { BlockId } from '../engine/types.ts'

interface PaletteProps {
  brush: number
  onBrush: (id: BlockId) => void
}

export function Palette({ brush, onBrush }: PaletteProps) {
  return (
    <aside className="palette" aria-label="Paleta klocków">
      <h2>Klocki</h2>
      <div className="palette-grid">
        {HOTBAR.map((b) => (
          <button
            key={b.id}
            type="button"
            className={brush === b.id ? 'swatch selected' : 'swatch'}
            style={{ '--sw': b.color, '--sd': b.dark } as CSSProperties}
            onClick={() => onBrush(b.id)}
            title={`${b.name} — ${b.hint}`}
            data-testid={`brush-${b.id}`}
          >
            <span className="swatch-cube" />
            <span className="swatch-name">{b.name}</span>
          </button>
        ))}
      </div>
      <p className="hint">
        LPM stawia, PPM kasuje. 1–9 wybiera klocek, 0 gumka. Cofnij: Ctrl+Z.
      </p>
    </aside>
  )
}
