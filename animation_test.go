package main

import "testing"

func TestAnimate(t *testing.T) {
	for _, data := range []struct {
		Tick       int
		StartFrame int
		Want       int
		FrameTags  FrameTags
		Reason     string
	}{
		{0, 1, 2, FrameTags{From: 0, To: 7}, "increments frame (mid-loop)"},
		{0, 7, 0, FrameTags{From: 0, To: 7}, "wraps around (loop back to first frame)"},
		{0, 9, 0, FrameTags{From: 0, To: 7}, "interrupts other animation (out of upper range)"},
		{0, 9, 10, FrameTags{From: 9, To: 10}, "increments frame (toggling 2 frames)"},
		{0, 10, 9, FrameTags{From: 9, To: 10}, "wraps around (toggling 2 frames)"},
		{0, 1, 9, FrameTags{From: 9, To: 10}, "interrupts other animation (out of lower range)"},
		{0, 8, 8, FrameTags{From: 8, To: 8}, "stays in place with single frame"},
		{1, 1, 1, FrameTags{From: 0, To: 7}, "stays in place when tick is not modulo 5"},
		{10, 1, 2, FrameTags{From: 0, To: 7}, "increments frame when tick is modulo 5"},
		{1, 1, 9, FrameTags{From: 9, To: 10}, "interrupts immediately regardless of modulo"},
		{1, 7, 8, FrameTags{From: 8, To: 8}, "immediate interrupt with single frame"},
	} {
		if got := Animate(data.StartFrame, data.Tick, data.FrameTags); got != data.Want {
			t.Errorf(
				"Next animation frame from %d for [%d→%d] at tick %d was %d, want %d, because: %s",
				data.StartFrame, data.FrameTags.From, data.FrameTags.To, data.Tick, got, data.Want, data.Reason,
			)
		}
	}
}
