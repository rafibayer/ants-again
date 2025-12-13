package main

import "github.com/rafibayer/ants-again/vector"

type Food struct {
	// Position
	*vector.Vector

	amount int
}

func (g *Game) updateFood() {
	g.remainingFoodCount = 0

	toRemove := make([]*Food, 0)
	for food := range g.food.PointsIter() {
		g.remainingFoodCount += food.amount
		if food.amount <= 0 {
			toRemove = append(toRemove, food)
		}
	}

	for _, r := range toRemove {
		g.food.Remove(r)
	}
}
