package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type Stats struct {
	ticks int
	fps   string
	tps   string

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
		ticks: g.tickCount,
		fps:   fmt.Sprintf("%.0f", ebiten.ActualFPS()),
		tps:   fmt.Sprintf("%.0f", ebiten.ActualTPS()),
		ants: struct {
			foraging  int
			returning int
		}{
			foraging:  g.foragingAntCount,
			returning: g.returningAntCount,
		},
		food: struct {
			left      int
			collected int
		}{
			left:      g.remainingFoodCount,
			collected: g.collectedFood,
		},
		pheromone: struct {
			forage   int
			returing int
		}{
			forage:   g.foragingPheromone.Len(),
			returing: g.returningPheromone.Len(),
		},
	}
}
