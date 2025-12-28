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

	g.handleEdit()
}

func (g *Game) handleEdit() {
	xs, ys := ebiten.CursorPosition()
	xw, yw := g.screenToWorldSpace(float64(xs), float64(ys))
	v := vector.Vector{X: xw, Y: yw}

	cursorMode := CursorMode(g.cursorModeIndex)

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		switch cursorMode {
		case CursorModeFood:
			near := g.food.RadialSearchIter(v, ANT_FOOD_RADIUS/2.0)
			for range near {
				// don't place if there is food nearby already
				return
			}
			g.food.Insert(&Food{amount: FOOD_START, Vector: &v})
		case CursorModeObstacle:
			near := g.obstacles.RadialSearchIter(v, OBSTACLE_HASH_CELL_SIZE/2.0)
			for range near {
				// don't place if there is obstacle nearby already
				return
			}
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
