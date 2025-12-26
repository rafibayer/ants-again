package main

import (
	"image"

	"github.com/ebitengine/debugui"
)

func ui(g *Game) func(ctx *debugui.Context) error {
	return func(ctx *debugui.Context) error {
		// Window(title, default position/size, contents)
		ctx.Window("settings", image.Rect(50, 50, 300, 300), func(layout debugui.ContainerLayout) {
			// Slider for ant speed
			ctx.Text("ant speed")
			// SliderF takes a pointer to float64, low, high, step, and number of decimals
			ctx.SliderF(&g.params.AntSpeed, 0.5, 5.0, 0.1, 2)

			ctx.Text("ant rotation")
			ctx.SliderF(&g.params.AntRotation, 0.0, 20, 0.5, 1)

			ctx.Text("ant sensor radius")
			ctx.SliderF(&g.params.PheromoneSenseRadius, 50.0, 250, 5, 1)

			ctx.Checkbox(&g.params.DebugDrawSensorRange, "debug sensor range")
		})
		return nil
	}
}
