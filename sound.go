package main

import (
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 44100

// sounds holds the effects, generated at start-up so the game needs no
// audio files.
type sounds struct {
	ctx *audio.Context
	on  bool

	click, start, stop, win, bigWin, bye []byte
}

func newSounds() *sounds {
	s := &sounds{ctx: audio.NewContext(sampleRate), on: true}

	s.start = append(tone(330, 40, 0.2, true), tone(440, 60, 0.2, true)...)
	s.stop = tone(110, 70, 0.4, true)
	s.click = tone(660, 30, 0.15, true)

	for _, f := range []float64{523, 659, 784} {
		s.win = append(s.win, tone(f, 110, 0.25, true)...)
	}
	for _, f := range []float64{523, 659, 784, 1047, 784, 1047, 1319, 1047, 1319, 1568} {
		s.bigWin = append(s.bigWin, tone(f, 110, 0.25, true)...)
	}

	// The original said goodbye with a burst of random PC-speaker beeps.
	for range 60 {
		s.bye = append(s.bye, tone(200+rand.Float64()*4000, 20, 0.12, false)...)
	}
	return s
}

func (s *sounds) play(pcm []byte) {
	if s.on {
		s.ctx.NewPlayerFromBytes(pcm).Play()
	}
}

// tone returns a square wave as 16-bit stereo PCM. With fade set the note
// dies away over its length; otherwise it holds its volume.
func tone(freq float64, ms int, volume float64, fade bool) []byte {
	n := sampleRate * ms / 1000
	pcm := make([]byte, 0, n*4)
	for i := range n {
		v := volume
		if math.Sin(2*math.Pi*freq*float64(i)/sampleRate) < 0 {
			v = -volume
		}
		if fade {
			v *= 1 - float64(i)/float64(n)
		}
		sample := int16(v * math.MaxInt16)
		lo, hi := byte(sample), byte(sample>>8)
		pcm = append(pcm, lo, hi, lo, hi)
	}
	return pcm
}
