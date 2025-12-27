package main

import (
	"image"

	"github.com/ebitengine/debugui"
)

func ui(g *Game) func(ctx *debugui.Context) error {
	return func(ctx *debugui.Context) error {
		const x0 = 50
		const y0 = 300
		const width = 250
		const height = 350
		const x1 = x0 + width
		const y1 = y0 + height

		// Window(title, default position/size, contents)
		ctx.Window("settings", image.Rect(x0, y0, x1, y1), func(layout debugui.ContainerLayout) {
			// Slider for ant speed
			ctx.Text("ant speed")
			// SliderF takes a pointer to float64, low, high, step, and number of decimals
			ctx.SliderF(&g.params.AntSpeed, 0.5, 5.0, 0.1, 2)

			ctx.Text("ant rotation")
			ctx.SliderF(&g.params.AntRotation, 0.0, 20, 0.5, 1)

			ctx.Text("pheromone influence")
			ctx.SliderF(&g.params.PheromoneInfluence, 0.0, 5, 0.5, 1)

			ctx.Text("pheromone sense radius")
			ctx.SliderF(&g.params.PheromoneSenseRadius, 50.0, 250, 5, 1)

			ctx.Checkbox(&g.params.DebugDrawSensorRange, "debug sense range")

			ctx.Text("Cursor mode (left: add, right: remove)")
			ctx.Dropdown(&g.cursorModeIndex, cursorOptions)

			ctx.Text("Boundary mode")
			ctx.Dropdown(&g.params.BoundaryModeIndex, boundaryModes)
		})
		return nil
	}
}
