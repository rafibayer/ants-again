package main

import (
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rafibayer/ants-again/kdtree"
)

const (
	GAME_SIZE = 500

	ANT_SPEED       = 2.5
	ANT_FOOD_RADIUS = 5.0  // radius in which an ant will pick up food
	ANT_HILL_RADIUS = 15.0 // raidus in which an ant will return to hill

	PHEROMONE_SENSE_RADIUS = 100.0               // radius in which an ant will smell pheromones
	PHEROMONE_DECAY        = (1.0 / 60.0) / 15.0 // denominator is number of seconds until decay
	PHEROMONE_DROP_PROB    = 1.0 / 120.0         // odds of dropping a pheromone per tick

	// we can tune perf with these, less sensing, but with more effect to reduce kdtree searches
	PHEROMONE_INFLUENCE  = 2.00      // increase or decrease effect of pheromone on direction
	PHEROMONE_SENSE_PROB = 1.0 / 5.0 // odds of smelling for pheromones per tick

	FOOD_START = 50
)

type Food struct {
	// Position
	*Vector

	amount int
}

type AntState int

const (
	FORAGE AntState = iota
	RETURN
)

type Ant struct {
	// position
	Vector

	dir   Vector
	state AntState
}

type Pheromone struct {
	// position
	*Vector

	amount float32
}

type Game struct {
	frameCount, tickCount int

	camX, camY float64
	zoom       float64

	world *ebiten.Image
	px    []byte // RGBA buffer: width * height * 4

	food  *kdtree.KDTree
	ants  *kdtree.KDTree
	hills *kdtree.KDTree

	foragingPheromone  *kdtree.KDTree
	returningPheromone *kdtree.KDTree

	// cached tree sizes for stat reporting
	cachedAntsCount              int
	cachedForagingPheromoneCount int
	cachedReturningPheromone     int
	cachedFood                   int
}

func NewGame() *Game {
	ants := kdtree.New(nil)
	food := kdtree.New(nil)
	hills := kdtree.New(nil)

	for range 500 {
		ants.Insert(&Ant{
			Vector: Vector{GAME_SIZE / 2, GAME_SIZE / 2},
			dir:    Vector{Rand(-1, 1), Rand(-1, 1)},
			state:  FORAGE,
		})
	}

	for r := range 30 {
		for c := range 10 {
			food.Insert(&Food{
				Vector: &Vector{x: GAME_SIZE/5 + float64(r)*1.5, y: GAME_SIZE/5 + float64(c)*1.5},
				amount: FOOD_START,
			})

			food.Insert(&Food{
				Vector: &Vector{x: GAME_SIZE*(5.0/6.0) + float64(r)*1.5, y: GAME_SIZE/2 + float64(c)*1.5},
				amount: FOOD_START,
			})
		}
	}

	hills.Insert(Vector{GAME_SIZE / 2, GAME_SIZE / 2})

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

		foragingPheromone:  kdtree.New(nil),
		returningPheromone: kdtree.New(nil),
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
	g.cachedAntsCount = 0

	// update each ant, add to next and state
	for a := range g.ants.Chan() {
		g.cachedAntsCount++
		ant := a.(*Ant)
		ant.Vector = ant.Add(ant.dir.Normalize().Mul(ANT_SPEED))

		// randomly rotate a few degrees
		ant.dir = ant.dir.Rotate(Rand(-2.5, 2.5))

		row, col := ant.ToGrid()
		keepInbounds(ant, row, col)

		if rand.Float32() < PHEROMONE_SENSE_PROB {
			// pheromone field to search based on ant state
			pheromone := g.returningPheromone
			if ant.state == RETURN {
				pheromone = g.foragingPheromone
			}

			// influence direction based on pheromone
			pheromoneDir := ZERO
			nearby := KDSearchRadius(pheromone, ant.Vector, PHEROMONE_SENSE_RADIUS)

			for _, p := range nearby {
				pher := p.(*Pheromone)

				// direction to pheromone and signal strength
				dirToSpot := pher.Sub(ant.Vector).Normalize()

				// scale by weight, distance to ant, and angular similarity
				strength := float64(pher.amount)
				strength = strength / max(0.1, ant.Vector.Distance(*pher.Vector)) // prevent overweighting really close smells
				strength *= LinearRemap(ant.dir.CosineSimilarity(dirToSpot))      // discount dissimilar angles
				pheromoneDir = pheromoneDir.Add(dirToSpot.Mul(strength))
			}

			ant.dir = ant.dir.Add(pheromoneDir.Mul(PHEROMONE_INFLUENCE)).Normalize()
		}

		if ant.state == FORAGE {
			// check for food nearby, change state and turn around
			nearFood := KDSearchRadius(g.food, ant.Vector, ANT_FOOD_RADIUS)
			for _, f := range nearFood {
				food := f.(*Food)
				if food.amount > 0 {
					food.amount--
					ant.state = RETURN
					ant.dir = ant.dir.Mul(-1.0)
					break // only grab 1 food
				}
			}
		}

		if ant.state == RETURN {
			// check for hill nearby, change state and turn around
			nearHill := KDSearchRadius(g.hills, ant.Vector, ANT_HILL_RADIUS)
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
				g.foragingPheromone.Insert(&Pheromone{Vector: &Vector{x: ant.x, y: ant.y}, amount: 1.0})
			case RETURN:
				g.returningPheromone.Insert(&Pheromone{Vector: &Vector{x: ant.x, y: ant.y}, amount: 1.0})
			}
		}

	}
}

func keepInbounds(ant *Ant, row int, col int) {
	if row < 0 {
		ant.dir.y = 1
	}
	if row >= GAME_SIZE {
		ant.dir.y = -1
	}
	if col < 0 {
		ant.dir.x = 1
	}
	if col >= GAME_SIZE {
		ant.dir.x = -1
	}
}

func (g *Game) updatePheromones() {
	g.cachedForagingPheromoneCount = 0
	for p := range g.foragingPheromone.Chan() {
		g.cachedForagingPheromoneCount++
		pher := p.(*Pheromone)

		pher.amount -= PHEROMONE_DECAY
		if pher.amount <= 0 {
			g.foragingPheromone.Remove(p)
		}
	}

	g.cachedReturningPheromone = 0
	for p := range g.returningPheromone.Chan() {
		g.cachedReturningPheromone++
		pher := p.(*Pheromone)

		pher.amount -= PHEROMONE_DECAY
		if pher.amount <= 0 {
			g.returningPheromone.Remove(p)
		}
	}

	if g.tickCount%30 == 0 {
		g.foragingPheromone.Balance()
		g.returningPheromone.Balance()
	}
}

func (g *Game) updateFood() {
	g.cachedFood = 0

	for f := range g.food.Chan() {
		food := f.(*Food)
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
