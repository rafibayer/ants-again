package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rafibayer/ants-again/kdtree"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenW = 800
	screenH = 600
)

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Clear()
	g.world.Clear()

	drawScreenSpace(g, screen)
	drawWorldSpace(g)

	screen.DrawImage(g.world, g.getCameraOpts())

	g.frameCount++
}

func (g *Game) getCameraOpts() *ebiten.DrawImageOptions {
	// translate op for drawing world in screen space with pan and zoom
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-g.camX, -g.camY)       // Pan (world space)
	op.GeoM.Translate(-screenW/2, -screenH/2) //  Move pivot to center of screen
	op.GeoM.Scale(g.zoom, g.zoom)             // Zoom
	op.GeoM.Translate(screenW/2, screenH/2)   //  Move pivot back
	return op
}

func drawScreenSpace(_ *Game, screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %.0f\nTPS: %.0f", ebiten.ActualFPS(), ebiten.ActualTPS()))
}

// drawCalls should be ordered from back to front.
func drawWorldSpace(g *Game) {
	// has to be first because of use of "writePixels"
	g.drawPheromones()

	g.drawAnts()
	g.drawFood()
	g.drawHills()

	// game world bounding box
	vector.StrokeRect(g.world, 0, 0, GAME_SIZE, GAME_SIZE, 5, color.White, true)
}

func (g *Game) drawAnts() {
	for a := range g.ants.Chan() {
		ant := a.(Ant)

		// debug circle -- food search area
		// sensor := ant.Vector.Add(ant.dir.Normalize().Mul(ANT_SENSOR_DIST))
		// vector.StrokeCircle(g.world, float32(sensor.x), float32(sensor.y), ANT_SENSOR_RADIUS, 1.0, color.White, true)

		// 1 away from ant facing
		tail := ant.Add(ant.dir.Normalize().Mul(-5))

		c := GREEN
		if ant.state == RETURN {
			c = LILAC
		}

		vector.StrokeLine(g.world, float32(ant.x), float32(ant.y), float32(tail.x), float32(tail.y), 2, c, true)
	}
}

func (g *Game) drawFood() {
	for f := range g.food.Chan() {
		food := f.(*Food)

		c := Fade(BROWN, float32(food.amount/FOOD_START))
		vector.FillCircle(g.world, float32(food.x), float32(food.y), 1.5, c, true)
	}
}

func (g *Game) drawHills() {
	for h := range g.hills.Chan() {
		hill := h.(Vector)

		vector.FillCircle(g.world, float32(hill.x), float32(hill.y), 15.0, WHITE, true)
	}
}

func (g *Game) drawPheromones() {
	// Clear buffer to black (or background color)
	for i := range g.px {
		g.px[i] = 0
	}

	writePheromones := func(ph *kdtree.KDTree, color color.RGBA) {
		for p := range ph.Chan() {
			pher := p.(*Pheromone)

			// Fade color by pheromone amount (0..1)
			c := Fade(color, pher.amount)

			x := int(pher.x)
			y := int(pher.y)
			if x < 0 || x >= GAME_SIZE || y < 0 || y >= GAME_SIZE {
				continue
			}

			idx := 4 * (y*GAME_SIZE + x)
			g.px[idx+0] = c.R
			g.px[idx+1] = c.G
			g.px[idx+2] = c.B
			g.px[idx+3] = 255
		}
	}

	writePheromones(g.foragingPheromone, DARK_GREEN)
	writePheromones(g.returningPheromone, DARK_LILAC)

	// Write the pixel buffer to the ebiten.Image once
	g.world.WritePixels(g.px)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screenW, screenH
}
