package main

import (
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rafibayer/ants-again/spatial"
	vec "github.com/rafibayer/ants-again/vector"
)

const (
	GAME_SIZE = 1000

	ANT_SPEED       = 2.5
	ANT_FOOD_RADIUS = 5.0              // radius in which an ant will pick up food
	ANT_HILL_RADIUS = GAME_SIZE / 30.0 // radius in which an ant will return to hill
	ANT_ROTATION    = 11.0             // max rotation degrees in either direction per tick
	ANT_BOUNDARY    = TURN             // if true, ants will wrap around to the other side instead of turning at boundaries

	PHEROMONE_SENSE_RADIUS float64 = GAME_SIZE / 6.0     // radius in which an ant will smell pheromones
	PHEROMONE_DECAY                = (1.0 / 60.0) / 15.0 // denominator is number of seconds until decay
	PHEROMONE_DROP_PROB            = 1.0 / 120.0         // odds of dropping a pheromone per tick

	// we can tune perf with these, less sensing, but with more effect to reduce kdtree searches
	PHEROMONE_INFLUENCE  = 2.00      // increase or decrease effect of pheromone on direction
	PHEROMONE_SENSE_PROB = 1.0 / 5.0 // odds of smelling for pheromones per tick

	FOOD_START = 50
)

type Food struct {
	// Position
	*vec.Vector

	amount int
}

type AntState int

const (
	FORAGE AntState = iota
	RETURN
)

type Ant struct {
	// position
	vec.Vector

	dir   vec.Vector
	state AntState
}

type BoundaryBehavior int

const (
	WRAP BoundaryBehavior = iota
	TURN
)

type Pheromone struct {
	// position
	*vec.Vector

	amount float32
}

type Game struct {
	frameCount, tickCount int

	camX, camY float64
	zoom       float64

	world *ebiten.Image
	px    []byte // pixel buffer: width * height * 4 (R,G,B,A)

	ants  []*Ant
	food  spatial.Spatial[*Food]
	hills spatial.Spatial[vec.Vector]

	foragingPheromone  spatial.Spatial[*Pheromone]
	returningPheromone spatial.Spatial[*Pheromone]

	// cached tree sizes for stat reporting
	cachedForagingCount          int
	cachedReturningCount         int
	cachedForagingPheromoneCount int
	cachedReturningPheromone     int
	cachedFood                   int
}

func NewGame() *Game {
	ants := []*Ant{}
	food := spatial.NewHash[*Food](GAME_SIZE / 100.0)
	hills := spatial.NewHash[vec.Vector](GAME_SIZE / 5.0)

	for range 1000 {
		ants = append(ants, &Ant{
			Vector: vec.Vector{X: GAME_SIZE / 2, Y: GAME_SIZE / 2},
			dir:    vec.Vector{X: Rand(-1, 1), Y: Rand(-1, 1)},
			state:  FORAGE,
		})
	}

	for r := range 30 {
		for c := range 10 {
			food.Insert(&Food{
				Vector: &vec.Vector{X: GAME_SIZE/5 + float64(r)*1.5, Y: GAME_SIZE/5 + float64(c)*1.5},
				amount: FOOD_START,
			})

			food.Insert(&Food{
				Vector: &vec.Vector{X: GAME_SIZE*(5.0/6.0) + float64(r)*1.5, Y: GAME_SIZE/2 + float64(c)*1.5},
				amount: FOOD_START,
			})
		}
	}

	hills.Insert(vec.Vector{X: GAME_SIZE / 2, Y: GAME_SIZE / 2})

	return &Game{
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
		ant.Vector = ant.Add(ant.dir.Normalize().Mul(ANT_SPEED))

		// randomly rotate a few degrees
		ant.dir = ant.dir.Rotate(Rand(-ANT_ROTATION, ANT_ROTATION))

		row, col := ant.ToGrid()
		keepInbounds(ant, row, col)

		if rand.Float32() < PHEROMONE_SENSE_PROB {
			// pheromone field to search based on ant state
			pheromone := g.returningPheromone
			if ant.state == RETURN {
				pheromone = g.foragingPheromone
			}

			// influence direction based on pheromone
			pheromoneDir := vec.ZERO

			nearby := pheromone.RadialSearch(&Pheromone{Vector: &ant.Vector}, PHEROMONE_SENSE_RADIUS, func(a, b *Pheromone) float64 {
				return a.Distance(*b.Vector)
			})

			for _, pher := range nearby {

				// direction to pheromone and signal strength
				dirToSpot := pher.Sub(ant.Vector).Normalize()

				// scale by weight, distance to ant, and angular similarity
				strength := float64(pher.amount)
				strength = strength / max(0.1, ant.Vector.Distance(*pher.Vector)) // prevent overweighting really close smells
				strength *= LinearRemap(ant.dir.CosineSimilarity(dirToSpot))      // discount dissimilar angles

				pheromoneDir = pheromoneDir.Add(dirToSpot.Mul(strength))
			}

			ant.dir = ant.dir.Add(pheromoneDir.Mul(PHEROMONE_INFLUENCE))
			ant.dir = ant.dir.Normalize()
		}

		if ant.state == FORAGE {
			g.cachedForagingCount++

			// check for food nearby, change state and turn around
			nearFood := g.food.RadialSearch(&Food{Vector: &ant.Vector}, ANT_FOOD_RADIUS, func(a, b *Food) float64 {
				return a.Distance(*b.Vector)
			})

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

			nearHill := g.hills.RadialSearch(ant.Vector, ANT_HILL_RADIUS, vec.Vector.Distance)
			// check for hill nearby, change state and turn around
			if len(nearHill) > 0 {
				// turn around and go back to foraging
				ant.state = FORAGE
				ant.dir = ant.dir.Mul(-1.0)
			}
		}

		dropPheromone := rand.Float32() < PHEROMONE_DROP_PROB
		// only drop pheromone if inbounds
		if dropPheromone && (row >= 0 && row < GAME_SIZE && col >= 0 && col < GAME_SIZE) {
			switch ant.state {
			case FORAGE:
				g.foragingPheromone.Insert(&Pheromone{Vector: &vec.Vector{X: ant.X, Y: ant.Y}, amount: 1.0})
			case RETURN:
				g.returningPheromone.Insert(&Pheromone{Vector: &vec.Vector{X: ant.X, Y: ant.Y}, amount: 1.0})
			}
		}

	}
}

func keepInbounds(ant *Ant, row int, col int) {
	if ANT_BOUNDARY == WRAP {
		if row < 0 {
			ant.Y = GAME_SIZE
		}
		if row >= GAME_SIZE {
			ant.Y = 0
		}
		if col < 0 {
			ant.X = GAME_SIZE
		}
		if col >= GAME_SIZE {
			ant.X = 0
		}
	}

	if ANT_BOUNDARY == TURN {
		if row < 0 {
			ant.dir.Y = 1
		}
		if row >= GAME_SIZE {
			ant.dir.Y = -1
		}
		if col < 0 {
			ant.dir.X = 1
		}
		if col >= GAME_SIZE {
			ant.dir.X = -1
		}
	}
}

func (g *Game) updatePheromones() {
	g.cachedForagingPheromoneCount = 0
	toRemove := make([]*Pheromone, 0)
	for pher := range g.foragingPheromone.Chan() {
		g.cachedForagingPheromoneCount++

		pher.amount -= PHEROMONE_DECAY
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

		pher.amount -= PHEROMONE_DECAY
		if pher.amount <= 0 {
			toRemove = append(toRemove, pher)
		}
	}

	for _, r := range toRemove {
		g.returningPheromone.Remove(r)
	}
}

func (g *Game) updateFood() {
	g.cachedFood = 0

	for food := range g.food.Chan() {
		g.cachedFood += food.amount

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
