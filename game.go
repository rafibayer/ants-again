package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rafibayer/ants-again/spatial"
	"github.com/rafibayer/ants-again/vector"
)

// todo: time based things should be relative to TPS.. I think?
// we want the same stuff to happen, but faster if we increase TPS
const (
	TPS       = 60
	GAME_SIZE = 1000

	ANT_FOOD_RADIUS = GAME_SIZE / 200.0 // radius in which an ant will pick up food
	ANT_HILL_RADIUS = GAME_SIZE / 30.0  // radius in which an ant will return to hill
	ANT_BOUNDARY    = TURN              // if true, ants will wrap around to the other side instead of turning at boundaries

	FOOD_START = 50
)

type Params struct {
	AntSpeed             float64 // 2.5
	AntRotation          float64 // 9.0
	PheromoneSenseRadius float64 // 166.67 = (game_size / 6.0)
	PheromoneDecay       float32 // 0.001111 = (1.0 / TPS) / 15.0
	PheromoneDropProb    float64 // 0.008333 = 1.0 / (2 * TPS)
	PheromoneInfluence   float64 // 2.0
	PheromoneSenseProb   float64 // 1.0 / 5.0
}

var DefaultParams = Params{
	AntSpeed:             2.5,
	AntRotation:          9.0,
	PheromoneSenseRadius: 166.67,   //  = (game_size / 6.0)
	PheromoneDecay:       0.001111, //  = (1.0 / TPS) / 15.0
	PheromoneDropProb:    0.008333, //  = 1.0 / (2 * TPS)
	PheromoneInfluence:   2.0,
	PheromoneSenseProb:   1.0 / 5.0,
}

type Food struct {
	// Position
	*vector.Vector

	amount int
}

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
}

type BoundaryBehavior int

const (
	WRAP BoundaryBehavior = iota
	TURN
)

type Pheromone struct {
	// position
	*vector.Vector

	amount float32
}

type Game struct {
	params *Params

	frameCount, tickCount int

	camX, camY float64
	zoom       float64

	world *ebiten.Image
	px    []byte // pixel buffer: width * height * 4 (R,G,B,A)

	ants []*Ant
	food spatial.Spatial[*Food]

	hills         spatial.Spatial[vector.Vector]
	collectedFood int

	foragingPheromone  spatial.Spatial[*Pheromone]
	returningPheromone spatial.Spatial[*Pheromone]

	// cached spatial data structure sizes for stat reporting
	// todo: spatial.Hash can track size in O(1), we could drop all this
	cachedForagingCount          int
	cachedReturningCount         int
	cachedForagingPheromoneCount int
	cachedReturningPheromone     int
	cachedRemainingFood          int
}

func NewGame(params *Params) *Game {
	if params == nil {
		params = &DefaultParams
	}

	ants := []*Ant{}
	food := spatial.NewHash[*Food](GAME_SIZE / 100.0)
	hills := spatial.NewHash[vector.Vector](GAME_SIZE / 5.0)

	for range 1000 {
		ants = append(ants, &Ant{
			Vector: vector.Vector{X: GAME_SIZE / 2, Y: GAME_SIZE / 2},
			dir:    vector.Vector{X: Rand(-1, 1), Y: Rand(-1, 1)},
			state:  FORAGE,
		})
	}

	for r := range 30 {
		for c := range 10 {
			// top left
			food.Insert(&Food{
				Vector: &vector.Vector{X: GAME_SIZE/5 + float64(r)*1.5, Y: GAME_SIZE/5 + float64(c)*1.5},
				amount: FOOD_START,
			})

			// mid right
			food.Insert(&Food{
				Vector: &vector.Vector{X: GAME_SIZE*(5.0/6.0) + float64(r)*1.5, Y: GAME_SIZE/2 + float64(c)*1.5},
				amount: FOOD_START,
			})

			// far bottom right
			food.Insert(&Food{
				Vector: &vector.Vector{X: GAME_SIZE*(9.0/10.0) + float64(r)*1.5, Y: GAME_SIZE*(9.0/10.0) + float64(c)*1.5},
				amount: FOOD_START,
			})
		}
	}

	hills.Insert(vector.Vector{X: GAME_SIZE / 2, Y: GAME_SIZE / 2})

	return &Game{
		params:     params,
		frameCount: 0,
		tickCount:  0,

		camX: 0,
		camY: 0,
		zoom: 1.0,

		world: ebiten.NewImage(GAME_SIZE, GAME_SIZE),
		px:    make([]byte, GAME_SIZE*GAME_SIZE*4), // pheromone buffer: 4 bytes per pixel (R,G,B,A)

		ants:  ants,
		food:  food,
		hills: hills,

		// spatial hash size his a fairly import perf knob -- if these are mis-sized,
		// the cells get too crowded, or we have to search too many of them
		foragingPheromone:  spatial.NewHash[*Pheromone](GAME_SIZE / 20),
		returningPheromone: spatial.NewHash[*Pheromone](GAME_SIZE / 20),
	}
}

func (g *Game) Update() error {

	g.pollInput()
	g.updateAnts()
	g.updatePheromones()
	g.updateFood()

	g.tickCount++
	return nil
}

func (g *Game) updateAnts() {
	g.cachedForagingCount = 0
	g.cachedReturningCount = 0

	// update each ant, add to next and state
	for _, ant := range g.ants {
		ant.Vector = ant.Add(ant.dir.Normalize().Mul(float64(g.params.AntSpeed)))

		// randomly rotate a few degrees
		ant.dir = ant.dir.Rotate(Rand(-g.params.AntRotation, g.params.AntRotation))

		keepInbounds(ant)

		if Chance(g.params.PheromoneSenseProb) {
			// pheromone field to search based on ant state
			pheromone := g.returningPheromone
			if ant.state == RETURN {
				pheromone = g.foragingPheromone
			}

			// influence direction based on pheromone
			pheromoneDir := vector.ZERO

			nearby := pheromone.RadialSearch(ant.Vector, g.params.PheromoneSenseRadius)

			for _, pher := range nearby {
				// direction to pheromone and signal strength
				dirToSpot := pher.Sub(ant.Vector).Normalize()

				// scale by weight, distance to ant, and angular similarity
				strength := float64(pher.amount)
				strength = strength / max(0.2, ant.Vector.Distance(*pher.Vector)) // prevent overweighting really close smells
				strength *= max(ant.dir.CosineSimilarity(dirToSpot), 0)           // LinearRemap(ant.dir.CosineSimilarity(dirToSpot))      // discount dissimilar angles

				pheromoneDir = pheromoneDir.Add(dirToSpot.Mul(strength))
			}

			ant.dir = ant.dir.Add(pheromoneDir.Mul(g.params.PheromoneInfluence))
			ant.dir = ant.dir.Normalize()
		}

		if ant.state == FORAGE {
			g.cachedForagingCount++

			// check for food nearby, change state and turn around
			nearFood := g.food.RadialSearch(ant.Vector, ANT_FOOD_RADIUS)

			for _, food := range nearFood {
				if food.amount > 0 {
					food.amount--
					ant.state = RETURN
					ant.dir = ant.dir.Mul(-1.0)
					break // only grab 1 food
				}
			}
		}

		if ant.state == RETURN {
			g.cachedReturningCount++

			nearHill := g.hills.RadialSearch(ant.Vector, ANT_HILL_RADIUS)
			// check for hill nearby, change state and turn around
			if len(nearHill) > 0 {
				// turn around and go back to foraging
				ant.state = FORAGE
				g.collectedFood++
				ant.dir = ant.dir.Mul(-1.0)
			}
		}

		if Chance(g.params.PheromoneDropProb) {
			switch ant.state {
			case FORAGE:
				// todo: copy ant vector instead of realloc?
				g.foragingPheromone.Insert(&Pheromone{Vector: &vector.Vector{X: ant.X, Y: ant.Y}, amount: 1.0})
			case RETURN:
				g.returningPheromone.Insert(&Pheromone{Vector: &vector.Vector{X: ant.X, Y: ant.Y}, amount: 1.0})
			}
		}

	}
}

func keepInbounds(ant *Ant) {
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

func (g *Game) updatePheromones() {
	g.cachedForagingPheromoneCount = 0
	toRemove := make([]*Pheromone, 0)
	for pher := range g.foragingPheromone.Chan() {
		g.cachedForagingPheromoneCount++

		pher.amount -= g.params.PheromoneDecay
		if pher.amount <= 0 {
			toRemove = append(toRemove, pher)
		}
	}

	for _, r := range toRemove {
		g.foragingPheromone.Remove(r)
	}

	g.cachedReturningPheromone = 0
	toRemove = make([]*Pheromone, 0)
	for pher := range g.returningPheromone.Chan() {
		g.cachedReturningPheromone++

		pher.amount -= g.params.PheromoneDecay
		if pher.amount <= 0 {
			toRemove = append(toRemove, pher)
		}
	}

	for _, r := range toRemove {
		g.returningPheromone.Remove(r)
	}
}

func (g *Game) updateFood() {
	g.cachedRemainingFood = 0

	for food := range g.food.Chan() {
		g.cachedRemainingFood += food.amount

		if food.amount <= 0 {
			g.food.Remove(food)
		}
	}
}

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
