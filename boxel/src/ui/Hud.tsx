import type { PointerEvent } from 'react'
import type { HudState, InputState } from '../engine/types.ts'

interface HudProps {
  hud: HudState
  playing: boolean
  onInput: (partial: Partial<InputState>) => void
  onJump: () => void
  onReplay: () => void
  onStop: () => void
}

export function Hud({ hud, playing, onInput, onJump, onReplay, onStop }: HudProps) {
  const hold = (flag: keyof InputState) => {
    const start = (e: PointerEvent<HTMLButtonElement>) => {
      e.preventDefault()
      e.stopPropagation()
      e.currentTarget.setPointerCapture(e.pointerId)
      onInput({ [flag]: true })
    }
    const end = () => onInput({ [flag]: false })
    return {
      onPointerDown: start,
      onPointerUp: end,
      onPointerCancel: end,
      onLostPointerCapture: end,
    }
  }

  return (
    <>
      <div className="hud-top" aria-live="polite">
        <span data-testid="hud-score">PKT {hud.score}</span>
        <span data-testid="hud-coins">MONETY {hud.coins}</span>
        <span className="hud-lives" data-testid="hud-lives">
          ŻYCIA {hud.lives}
        </span>
      </div>
      {hud.message ? <div className="hud-msg">{hud.message}</div> : null}

      {playing ? (
        <div className="touch-controls">
          <div className="dpad">
            <button type="button" className="pad up" data-testid="pad-up" {...hold('forward')}>
              ▲
            </button>
            <button type="button" className="pad left" data-testid="pad-left" {...hold('left')}>
              ◀
            </button>
            <button type="button" className="pad right" data-testid="pad-right" {...hold('right')}>
              ▶
            </button>
            <button type="button" className="pad down" data-testid="pad-down" {...hold('back')}>
              ▼
            </button>
          </div>
          <button
            type="button"
            className="jump-btn"
            data-testid="pad-jump"
            onPointerDown={(e) => {
              e.preventDefault()
              e.stopPropagation()
              onJump()
              onInput({ jump: true })
            }}
            onPointerUp={() => onInput({ jump: false })}
            onPointerCancel={() => onInput({ jump: false })}
            onClick={() => onJump()}
          >
            SKOK
          </button>
        </div>
      ) : null}

      {hud.phase === 'won' || hud.phase === 'lost' ? (
        <div className="result-card" data-testid="result-card">
          <h2>{hud.phase === 'won' ? 'WYGRANA!' : 'PRZEGRANA'}</h2>
          <p>{hud.phase === 'won' ? 'Super, dobiegłeś do mety!' : 'Spróbuj jeszcze raz albo ułatw parkour.'}</p>
          <p className="result-score">Wynik: {hud.score}</p>
          <div className="result-actions">
            <button type="button" onClick={onReplay} data-testid="replay">
              Jeszcze raz
            </button>
            <button type="button" onClick={onStop}>
              Do studia
            </button>
          </div>
        </div>
      ) : null}
    </>
  )
}
