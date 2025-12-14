package main

import (
	"github.com/rafibayer/ants-again/util"
	"github.com/rafibayer/ants-again/vector"
)

type AntState int

const (
	FORAGE AntState = iota
	RETURN
)

type Ant struct {
	// position
	vector.Vector

	dir   vector.Vector
	state AntState

	pheromoneStored int
}

type BoundaryBehavior int

const (
	WRAP BoundaryBehavior = iota
	TURN
)

func (g *Game) updateAnts() {
	g.foragingAntCount = 0
	g.returningAntCount = 0

	for _, ant := range g.ants {
		ant.Vector = ant.Add(ant.dir.Normalize().Mul(float64(g.params.AntSpeed)))

		keepInbounds(ant)

		if util.Chance(g.params.PheromoneSenseProb) {
			// pheromone field to search based on ant state
			pheromone := g.returningPheromone
			if ant.state == RETURN {
				pheromone = g.foragingPheromone
			}

			// influence direction based on pheromone
			pheromoneDir := vector.ZERO

			nearby := pheromone.RadialSearchIter(ant.Vector, g.params.PheromoneSenseRadius)

			for pher := range nearby {
				// direction to pheromone and signal strength
				dirToSpot := pher.Sub(ant.Vector).Normalize()

				// scale by weight, distance to ant, and angular similarity
				strength := float64(pher.amount)
				strength = strength / max(0.1, ant.Vector.Distance(*pher.Vector)) // prevent overweighting really close smells

				cosineSim := ant.dir.CosineSimilarity(dirToSpot)
				if cosineSim < g.params.PheromoneSenseCosineSimilarity {
					strength *= 0
				}
				strength *= cosineSim

				pheromoneDir = pheromoneDir.Add(dirToSpot.Mul(strength))
			}

			ant.dir = ant.dir.Add(pheromoneDir.Mul(g.params.PheromoneInfluence))
			ant.dir = ant.dir.Normalize()
		}

		if ant.state == FORAGE {
			g.foragingAntCount++
			// check for food nearby, change state and turn around
			nearFood := g.food.RadialSearchIter(ant.Vector, ANT_FOOD_RADIUS)

			for food := range nearFood {
				if food.amount > 0 {
					ant.state = RETURN
					food.amount--
					ant.dir = ant.dir.Mul(-1.0)
					ant.pheromoneStored = g.params.AntPheromoneStart
					break // only grab 1 food
				}
			}
		}

		if ant.state == RETURN {
			g.returningAntCount++
			nearHill := g.hills.RadialSearchIter(ant.Vector, ANT_HILL_RADIUS)

			// check for hill nearby, change state and turn around
			for range nearHill {
				// turn around and go back to foraging
				ant.state = FORAGE
				g.collectedFood++
				ant.dir = ant.dir.Mul(-1.0)
				ant.pheromoneStored = g.params.AntPheromoneStart
				break
			}
		}

		if ant.pheromoneStored > 0 && util.Chance(g.params.PheromoneDropProb) {
			ant.pheromoneStored--
			switch ant.state {
			case FORAGE:
				g.foragingPheromone.Insert(&Pheromone{Vector: &vector.Vector{X: ant.X, Y: ant.Y}, amount: 1.0})
			case RETURN:
				g.returningPheromone.Insert(&Pheromone{Vector: &vector.Vector{X: ant.X, Y: ant.Y}, amount: 1.0})
			}
		}

		// randomly rotate a few degrees
		ant.dir = ant.dir.Rotate(util.Rand(-g.params.AntRotation, g.params.AntRotation))
	}
}

func keepInbounds(ant *Ant) {
	// wrapping behavior: ant teleports to other side when it hits boundary,
	//  retaints direction.
	if ANT_BOUNDARY == WRAP {
		if ant.Y < 0 {
			ant.Y = GAME_SIZE
		}
		if ant.Y >= GAME_SIZE {
			ant.Y = 0
		}
		if ant.X < 0 {
			ant.X = GAME_SIZE
		}
		if ant.X >= GAME_SIZE {
			ant.X = 0
		}
	}

	// turn behavior: ant turns around when it hits boundary,
	// retains position
	if ANT_BOUNDARY == TURN {
		if ant.Y < 0 {
			ant.dir.Y = 1
		}
		if ant.Y >= GAME_SIZE {
			ant.dir.Y = -1
		}
		if ant.X < 0 {
			ant.dir.X = 1
		}
		if ant.X >= GAME_SIZE {
			ant.dir.X = -1
		}
	}
}
