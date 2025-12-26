package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (g *Game) pollInput() {
	// Camera movement
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		g.camY -= 4
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.camY += 4
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.camX -= 4
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.camX += 4
	}

	// Zoom
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.zoom *= 1.02
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.zoom *= 0.98
	}
}
