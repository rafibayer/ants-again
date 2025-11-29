package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

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
	for _, a := range g.ants.Points() {
		ant := a.(Ant)

		// debug circle -- food search area
		// front := ant.Vector.Add(ant.dir.Normalize().Mul(2.0))
		// vector.StrokeCircle(g.world, float32(front.x), float32(front.y), 1.75, 1.0, color.White, true)

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
	for _, f := range g.food.Points() {
		food := f.(*Food)

		c := Fade(BROWN, float32(food.amount/FOOD_START))
		vector.FillCircle(g.world, float32(food.x), float32(food.y), 1.5, c, true)
	}
}

func (g *Game) drawHills() {
	for _, h := range g.hills.Points() {
		hill := h.(Vector)

		vector.FillCircle(g.world, float32(hill.x), float32(hill.y), 15.0, WHITE, true)
	}
}

func (g *Game) drawPheromones() {
	idx := 0 // byte index into g.px

	for r := range GAME_SIZE {
		for c := range GAME_SIZE {

			foraging := g.foragingPheromone[r][c]
			returning := g.returningPheromone[r][c]

			// Combine the two pheromones into one final color by blending values at position
			fc := Fade(DARK_GREEN, foraging)
			rc := Fade(DARK_LILAC, returning)

			// fc and rc are color.Color → extract RGBA
			fr, fg, fb, _ := fc.RGBA()
			rr, rg, rb, _ := rc.RGBA()

			// 16-bit → 8-bit convert: v>>8
			r8 := byte((fr >> 8) + (rr >> 8))
			g8 := byte((fg >> 8) + (rg >> 8))
			b8 := byte((fb >> 8) + (rb >> 8))

			// write RGBA to pixel buffer
			g.px[idx+0] = r8
			g.px[idx+1] = g8
			g.px[idx+2] = b8
			g.px[idx+3] = 255

			idx += 4
		}
	}

	g.world.WritePixels(g.px)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screenW, screenH
}
