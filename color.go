package main

import "image/color"

var (
	WHITE = color.RGBA{R: 255, G: 255, B: 255, A: 255}

	GREEN      = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	DARK_GREEN = color.RGBA{R: 0, G: 191, B: 0, A: 255}

	LILAC      = color.RGBA{R: 161, G: 131, B: 192, A: 255}
	DARK_LILAC = color.RGBA{R: 121, G: 98, B: 143, A: 255}

	BROWN = color.RGBA{R: 150, G: 75, B: 0, A: 255}
)

func Fade(c color.RGBA, factor float32) color.RGBA {
	return color.RGBA{
		R: uint8(float32(c.R) * factor),
		G: uint8(float32(c.G) * factor),
		B: uint8(float32(c.B) * factor),
		A: c.A,
	}
}
