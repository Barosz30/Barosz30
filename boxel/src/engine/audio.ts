import type { SoundName } from './types.ts'

let ctx: AudioContext | null = null

function audio(): AudioContext | null {
  if (typeof window === 'undefined') return null
  const AC = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext
  if (!AC) return null
  ctx ??= new AC()
  if (ctx.state === 'suspended') void ctx.resume()
  return ctx
}

function tone(freq: number, duration: number, type: OscillatorType, gain = 0.07): void {
  const ac = audio()
  if (!ac) return
  const osc = ac.createOscillator()
  const g = ac.createGain()
  osc.type = type
  osc.frequency.value = freq
  g.gain.value = gain
  osc.connect(g)
  g.connect(ac.destination)
  const now = ac.currentTime
  g.gain.setValueAtTime(gain, now)
  g.gain.exponentialRampToValueAtTime(0.001, now + duration)
  osc.start(now)
  osc.stop(now + duration)
}

export function playSound(name: SoundName): void {
  switch (name) {
    case 'coin':
      tone(880, 0.12, 'square', 0.05)
      tone(1320, 0.16, 'square', 0.04)
      break
    case 'jump':
      tone(240, 0.1, 'square', 0.04)
      break
    case 'win':
      tone(523, 0.18, 'square')
      setTimeout(() => tone(659, 0.18, 'square'), 120)
      setTimeout(() => tone(784, 0.28, 'square'), 240)
      break
    case 'lose':
      tone(196, 0.35, 'sawtooth', 0.05)
      break
    case 'bump':
      tone(90, 0.18, 'square', 0.06)
      break
    case 'pop':
      tone(420, 0.08, 'square', 0.04)
      break
    default:
      break
  }
}

export function unlockAudio(): void {
  audio()
}
