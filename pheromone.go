package main

import "github.com/rafibayer/ants-again/vector"

type Pheromone struct {
	// position
	*vector.Vector

	amount float32
}

func (g *Game) updatePheromones() {
	toRemove := make([]*Pheromone, 0)
	for pher := range g.foragingPheromone.PointsIter() {
		pher.amount -= g.params.PheromoneDecay
		if pher.amount <= 0 {
			toRemove = append(toRemove, pher)
		}
	}

	for _, r := range toRemove {
		g.foragingPheromone.Remove(r)
	}

	toRemove = make([]*Pheromone, 0)
	for pher := range g.returningPheromone.PointsIter() {
		pher.amount -= g.params.PheromoneDecay
		if pher.amount <= 0 {
			toRemove = append(toRemove, pher)
		}
	}

	for _, r := range toRemove {
		g.returningPheromone.Remove(r)
	}
}
