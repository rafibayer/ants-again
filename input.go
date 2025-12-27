package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rafibayer/ants-again/vector"
)

type CursorMode int

const (
	CursorModeNone CursorMode = iota
	CursorModeFood
	CursorModeObstacle
)

var cursorOptions = []string{"None", "Food", "Obstacle"}

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

	// ignore game mouse inputs if captured by UI
	if g.uiCapture {
		return
	}

	xs, ys := ebiten.CursorPosition()
	xw, yw := g.screenToWorldSpace(float64(xs), float64(ys))
	v := vector.Vector{X: xw, Y: yw}

	cursorMode := CursorMode(g.cursorModeIndex)

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		switch cursorMode {
		case CursorModeFood:
			g.food.Insert(&Food{amount: FOOD_START, Vector: &v})
		case CursorModeObstacle:
			g.obstacles.Insert(&Obstacle{Vector: v})
		default:
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		switch cursorMode {
		case CursorModeFood:
			toRemove := g.food.RadialSearch(v, ANT_FOOD_RADIUS)
			for _, r := range toRemove {
				g.food.Remove(r)
			}
		case CursorModeObstacle:
			toRemove := g.obstacles.RadialSearch(v, OBSTACLE_HASH_CELL_SIZE)
			for _, r := range toRemove {
				g.obstacles.Remove(r)
			}
		default:
		}
	}
}
