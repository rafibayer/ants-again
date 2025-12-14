package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rafibayer/ants-again/spatial"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenW = 800
	screenH = 600
)

func (g *Game) Draw(screen *ebiten.Image) {

	drawScreenSpace(g, screen)
	drawWorldSpace(g)

	screen.DrawImage(g.world, g.cameraOpts())

	g.frameCount++
}

func (g *Game) cameraOpts() *ebiten.DrawImageOptions {
	// translate op for drawing world in screen space with pan and zoom
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-g.camX, -g.camY)       // Pan (world space)
	op.GeoM.Translate(-screenW/2, -screenH/2) //  Move pivot to center of screen
	op.GeoM.Scale(g.zoom, g.zoom)             // Zoom
	op.GeoM.Translate(screenW/2, screenH/2)   //  Move pivot back
	return op
}

func drawScreenSpace(g *Game, screen *ebiten.Image) {
	stats := g.Stats()
	ebitenutil.DebugPrint(screen, fmt.Sprintf("%+v", stats))
}

// drawCalls should be ordered from back to front.
func drawWorldSpace(g *Game) {
	// has to be first because of use of "writePixels"
	g.drawPheromones()
	// g.naiveDrawPheromones()

	g.drawAnts()
	g.drawFood()
	g.drawHills()

	// game world bounding box
	vector.StrokeRect(g.world, 0, 0, GAME_SIZE, GAME_SIZE, 5, color.White, false)
}

func (g *Game) drawAnts() {
	for _, ant := range g.ants {
		tail := ant.Add(ant.dir.Normalize().Mul(-5))

		c := GREEN
		if ant.state == RETURN {
			c = LILAC
		}

		vector.StrokeLine(g.world, float32(ant.X), float32(ant.Y), float32(tail.X), float32(tail.Y), 2, c, false)
	}
}

func (g *Game) drawFood() {
	for food := range g.food.PointsIter() {
		c := Fade(BROWN, float32(food.amount/FOOD_START))
		vector.FillRect(g.world, float32(food.X), float32(food.Y), 1.5, 1.5, c, false)
	}
}

func (g *Game) drawHills() {
	for hill := range g.hills.PointsIter() {
		vector.FillCircle(g.world, float32(hill.X), float32(hill.Y), ANT_HILL_RADIUS, WHITE, false)
	}
}

func (g *Game) drawPheromones() {
	// Clear buffer to black (or background color)
	for i := range g.px {
		g.px[i] = 0
	}

	writePheromones := func(ph spatial.Spatial[*Pheromone], color color.RGBA) {
		for pher := range ph.PointsIter() {
			// Fade color by pheromone amount (0..1)
			c := Fade(color, pher.amount)

			x := int(pher.X)
			y := int(pher.Y)
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

func (g *Game) naiveDrawPheromones() {
	for pher := range g.foragingPheromone.PointsIter() {
		c := Fade(DARK_GREEN, pher.amount)
		vector.FillRect(g.world, float32(pher.X), float32(pher.Y), 3.0, 3.0, c, false)
	}

	for pher := range g.returningPheromone.PointsIter() {
		c := Fade(DARK_LILAC, pher.amount)
		vector.FillRect(g.world, float32(pher.X), float32(pher.Y), 3.0, 3.0, c, false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screenW, screenH
}
