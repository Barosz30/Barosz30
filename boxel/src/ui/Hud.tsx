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
  const hold = (partial: Partial<InputState>) => ({
    onPointerDown: (e: PointerEvent<HTMLButtonElement>) => {
      e.preventDefault()
      ;(e.currentTarget as HTMLButtonElement).setPointerCapture(e.pointerId)
      onInput(partial)
    },
    onPointerUp: () => onInput({ forward: false, back: false, left: false, right: false, jump: false }),
    onPointerCancel: () => onInput({ forward: false, back: false, left: false, right: false, jump: false }),
  })

  return (
    <>
      <div className="hud-top" aria-live="polite">
        <span data-testid="hud-score">PKT {hud.score}</span>
        <span data-testid="hud-coins">● {hud.coins}</span>
        <span data-testid="hud-lives">{'❤'.repeat(Math.max(0, hud.lives)) || '☠'}</span>
      </div>
      {hud.message ? <div className="hud-msg">{hud.message}</div> : null}

      {playing ? (
        <div className="touch-controls">
          <div className="dpad">
            <button type="button" className="pad up" data-testid="pad-up" {...hold({ forward: true })}>
              ▲
            </button>
            <button type="button" className="pad left" data-testid="pad-left" {...hold({ left: true })}>
              ◀
            </button>
            <button type="button" className="pad right" data-testid="pad-right" {...hold({ right: true })}>
              ▶
            </button>
            <button type="button" className="pad down" data-testid="pad-down" {...hold({ back: true })}>
              ▼
            </button>
          </div>
          <button
            type="button"
            className="jump-btn"
            data-testid="pad-jump"
            onPointerDown={(e) => {
              e.preventDefault()
              onJump()
              onInput({ jump: true })
            }}
            onPointerUp={() => onInput({ jump: false })}
          >
            SKOK
          </button>
        </div>
      ) : null}

      {hud.phase === 'won' || hud.phase === 'lost' ? (
        <div className="result-card" data-testid="result-card">
          <h2>{hud.phase === 'won' ? 'WYGRANA!' : 'KONIEC ŻYĆ'}</h2>
          <p>{hud.phase === 'won' ? 'Zdobyłeś wszystkie punkty.' : 'Spróbuj ułożyć łatwiejszy parkour.'}</p>
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
