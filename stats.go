package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type Stats struct {
	fps string
	tps string

	ants struct {
		foraging  int
		returning int
	}
	food struct {
		left      int
		collected int
	}
	pheromone struct {
		forage   int
		returing int
	}
}

func (g *Game) Stats() *Stats {
	return &Stats{
		fps: fmt.Sprintf("%.0f", ebiten.ActualFPS()),
		tps: fmt.Sprintf("%.0f", ebiten.ActualTPS()),
		ants: struct {
			foraging  int
			returning int
		}{
			foraging:  g.cachedForagingCount,
			returning: g.cachedReturningCount,
		},
		food: struct {
			left      int
			collected int
		}{
			left:      g.cachedRemainingFood,
			collected: g.collectedFood,
		},
		pheromone: struct {
			forage   int
			returing int
		}{
			forage:   g.cachedForagingPheromoneCount,
			returing: g.cachedReturningPheromone,
		},
	}
}
