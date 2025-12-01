package main

import (
	"log"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rafibayer/ants-again/kdtree"
)

const (
	GAME_SIZE = 1000

	ANT_SPEED         = 2.5
	ANT_SENSOR_RADIUS = 30.0

	PHEROMONE_DECAY              = (1.0 / 60.0) / 30.0 // denominator is number of seconds until decay
	PHEROMONE_INFLUENCE_DISCOUNT = 1.00                // reduce effect of pheromone on ant direction
	PHEROMONE_DROP_PROB          = 1.0 / 60.0          // odds of dropping a pheromone per tick

	FOOD_START = 10
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
}

func NewGame() *Game {
	ants := kdtree.New(nil)
	food := kdtree.New(nil)
	hills := kdtree.New(nil)

	for range 1000 {
		ants.Insert(Ant{
			Vector: Vector{GAME_SIZE / 2, GAME_SIZE / 2},
			dir:    Vector{Rand(-1, 1), Rand(-1, 1)},
			state:  FORAGE,
		})
	}

	for r := range 20 {
		for c := range 20 {
			food.Insert(&Food{
				Vector: &Vector{x: GAME_SIZE/3 + float64(r), y: GAME_SIZE/3 + float64(c)},
				amount: FOOD_START,
			})

			food.Insert(&Food{
				Vector: &Vector{x: GAME_SIZE*(2.0/3.0) + float64(r), y: GAME_SIZE/2 + float64(c)},
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
	nextAnts := kdtree.New(nil)

	// update each ant, add to next and state
	for a := range g.ants.Chan() {
		ant := a.(Ant)
		ant.Vector = ant.Add(ant.dir.Normalize().Mul(ANT_SPEED))

		// randomly rotate a few degrees
		ant.dir = ant.dir.Rotate(Rand(-2.5, 2.5))

		row, col := ant.ToGrid()
		keepInbounds(&ant, row, col)

		// pheromone field to search based on ant state
		pheromone := g.returningPheromone
		if ant.state == RETURN {
			pheromone = g.foragingPheromone
		}

		// influence direction based on pheromone
		pheromoneDir := ZERO
		nearby := KDSearchRadius(pheromone, ant.Vector, ANT_SENSOR_RADIUS)

		for _, p := range nearby {
			pher := p.(*Pheromone)
			r, c := pher.ToGrid()

			// direction to pheromone and signal strength
			spot := Vector{x: float64(c), y: float64(r)}
			dirToSpot := spot.Sub(ant.Vector).Normalize()

			// scale by weight, distance to ant, and angular similarity
			strength := float64(pher.amount)
			strength = strength / max(0.1, ant.Vector.Distance(spot)) // prevent overweighting really close smells
			strength *= ant.dir.CosineSimilarity(dirToSpot)
			pheromoneDir = pheromoneDir.Add(dirToSpot.Mul(strength))
		}

		ant.dir = ant.dir.Add(pheromoneDir.Mul(PHEROMONE_INFLUENCE_DISCOUNT)).Normalize()

		if ant.state == FORAGE {
			// check for food nearby, change state and turn around
			nearFood := KDSearchRadius(g.food, ant.Vector, 3)
			for _, f := range nearFood {
				food := f.(*Food)
				if food.amount > 0 {
					log.Print("found food")
					food.amount--
					ant.state = RETURN
					ant.dir = ant.dir.Mul(-1.0)
					break // only grab 1 food
				}
			}
		}

		if ant.state == RETURN {
			// check for hill nearby, change state and turn around
			nearHill := KDSearchRadius(g.hills, ant.Vector, 15.0)
			if len(nearHill) > 0 {
				log.Print("found hill")
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

		nextAnts.Insert(ant)
	}

	nextAnts.Balance()
	g.ants = nextAnts
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
	for p := range g.foragingPheromone.Chan() {
		pher := p.(*Pheromone)

		pher.amount -= PHEROMONE_DECAY
		if pher.amount <= 0 {
			g.foragingPheromone.Remove(p)
		}
	}

	for p := range g.returningPheromone.Chan() {
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
	nextFood := kdtree.New(nil)
	for f := range g.food.Chan() {
		food := f.(*Food)

		if food.amount > 0 {
			nextFood.Insert(food)
		}
	}

	nextFood.Balance()
	g.food = nextFood
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
