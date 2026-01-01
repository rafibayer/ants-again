package main

import (
	"image"

	"github.com/ebitengine/debugui"
)

func ui(g *Game) func(ctx *debugui.Context) error {
	return func(ctx *debugui.Context) error {
		const x0 = 0
		const y0 = 80
		const width = 160
		const height = 450
		const x1 = x0 + width
		const y1 = y0 + height

		// Window(title, default position/size, contents)
		ctx.Window("", image.Rect(x0, y0, x1, y1), func(layout debugui.ContainerLayout) {
			ctx.Header("settings", true, func() {
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

				ctx.Text("Boundary mode")
				ctx.Dropdown(&g.params.BoundaryModeIndex, boundaryModes)
			})

			ctx.Text("")

			ctx.Header("edit", true, func() {
				ctx.Text("left click - add")
				ctx.Text("right click - remove")
				ctx.Dropdown(&g.cursorModeIndex, cursorOptions)
			})

		})
		return nil
	}
}
