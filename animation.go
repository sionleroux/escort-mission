// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

// Animate determines the next animation frame for a sprite
func Animate(frame, tick int, ft FrameTags) int {
	from, to := ft.From, ft.To

	// Instantly start animation if state changed
	if frame < from || frame >= to {
		return from
	}

	// Update only in every 5th cycle
	if tick%5 != 0 {
		return frame
	}

	// Continuously increase the Frame counter between from and to
	return frame + 1
}
