package app

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"

	"boxel/internal/engine"
)

var audioCtx *audio.Context

func playSound(name engine.SoundName) {
	if audioCtx == nil {
		audioCtx = audio.NewContext(44100)
	}
	freq, dur, gain := 420.0, 0.08, 0.08
	switch name {
	case engine.SndCoin:
		beep(880, 0.12, 0.05)
		time.AfterFunc(40*time.Millisecond, func() { beep(1320, 0.16, 0.04) })
		return
	case engine.SndJump:
		freq, dur, gain = 240, 0.1, 0.04
	case engine.SndWin:
		beep(523, 0.18, 0.07)
		time.AfterFunc(120*time.Millisecond, func() { beep(659, 0.18, 0.07) })
		time.AfterFunc(240*time.Millisecond, func() { beep(784, 0.28, 0.07) })
		return
	case engine.SndLose:
		freq, dur, gain = 196, 0.35, 0.05
	case engine.SndBump:
		freq, dur, gain = 90, 0.18, 0.06
	case engine.SndPop:
		freq, dur, gain = 420, 0.08, 0.04
	}
	beep(freq, dur, gain)
}

func beep(freq, seconds, gain float64) {
	if audioCtx == nil {
		return
	}
	rate := 44100
	n := int(seconds * float64(rate))
	buf := make([]byte, n*4)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(rate)
		amp := math.Sin(2*math.Pi*freq*t) * gain * (1 - t/seconds)
		v := int16(amp * 30000)
		buf[i*4] = byte(v)
		buf[i*4+1] = byte(v >> 8)
		buf[i*4+2] = byte(v)
		buf[i*4+3] = byte(v >> 8)
	}
	p := audioCtx.NewPlayerFromBytes(buf)
	_ = p.Rewind()
	p.Play()
}
