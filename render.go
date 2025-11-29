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

	op := &ebiten.DrawImageOptions{}

	// --- IMPORTANT ORDER ---
	op.GeoM.Translate(-g.camX, -g.camY)       // Pan (world space)
	op.GeoM.Translate(-screenW/2, -screenH/2) //  Move pivot to center of screen
	op.GeoM.Scale(g.zoom, g.zoom)             // Zoom
	op.GeoM.Translate(screenW/2, screenH/2)   //  Move pivot back

	///// draw in world space
	drawWorldSpace(g, g.world)

	screen.DrawImage(g.world, op)

	g.frameCount++
}

func drawScreenSpace(_ *Game, screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, fmt.Sprintf("%.0f", ebiten.ActualFPS()))
}

func drawWorldSpace(g *Game, world *ebiten.Image) {
	ebitenutil.DebugPrint(world, "Worldspace")

	// game world bounding box
	vector.StrokeRect(g.world, 0, 0, 1000, 1000, 5, color.White, true)

	g.drawAnts()
	g.drawPheromones()
}

func (g *Game) drawAnts() {
	for _, a := range g.ants.Points() {
		ant := a.(Ant)

		// 1 away from ant facing
		tail := ant.Add(ant.dir.Normalize().Mul(-5))

		c := GREEN
		if ant.state == RETURN {
			c = LILAC
		}

		vector.StrokeLine(g.world, float32(ant.x), float32(ant.y), float32(tail.x), float32(tail.y), 2, c, true)
	}
}

func (g *Game) drawPheromones() {
	for _, f := range g.foragingPheromone.Points() {
		foraging := f.(Pheromone)

		c := Fade(DARK_GREEN, foraging.amount)
		vector.StrokeCircle(g.world, float32(foraging.x), float32(foraging.y), 2, 2, c, true)
	}

	for _, r := range g.returningPheromone.Points() {
		returning := r.(Pheromone)

		c := Fade(DARK_LILAC, returning.amount)
		vector.StrokeCircle(g.world, float32(returning.x), float32(returning.y), 2, 2, c, true)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return screenW, screenH
}
