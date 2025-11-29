package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kyroy/kdtree"
)

const (
	ANT_SPEED = 3.0

	PHEROMONE_DECAY = (1.0 / 60.0) / 10.0 // 60 FPS -> pheromone decays after 10 seconds
)

type Food struct {
	// Position
	Vector

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
	// Position
	Vector

	amount float32
}

type Game struct {
	frameCount, updateCount uint64

	camX, camY float64
	zoom       float64

	world *ebiten.Image

	food               *kdtree.KDTree
	ants               *kdtree.KDTree
	hills              *kdtree.KDTree
	foragingPheromone  *kdtree.KDTree
	returningPheromone *kdtree.KDTree
}

func NewGame() *Game {
	ants := kdtree.New(nil)

	for range 100 {
		ants.Insert(Ant{
			Vector: Vector{10, 10},
			dir:    Vector{1, 1},
			state:  FORAGE,
		})
	}

	return &Game{
		frameCount:  0,
		updateCount: 0,

		camX:  0,
		camY:  0,
		zoom:  1.0,
		world: ebiten.NewImage(1000, 1000),

		ants:               ants,
		foragingPheromone:  kdtree.New(nil),
		returningPheromone: kdtree.New(nil),
	}
}

func (g *Game) Update() error {

	g.pollInput()
	g.updateAnts()
	g.updatePheromones()

	g.updateCount++
	return nil
}

func (g *Game) updateAnts() {
	nextAnts := kdtree.New(nil)

	// update each ant, add to next and state
	for _, a := range g.ants.Points() {
		ant := a.(Ant)
		ant.Vector = ant.Add(ant.dir.Normalize().Mul(ANT_SPEED))

		// randomly rotate by up to 5 degrees
		ant.dir = ant.dir.Rotate(Rand(-5, 5))

		if g.updateCount%10 == 0 {
			switch ant.state {
			case FORAGE:
				// do we need to do this in "nextForagingPheromone" or something?
				// currently, ants would smell this pheromone the same frame that it's placed, fine?
				g.foragingPheromone.Insert(Pheromone{Vector: ant.Vector, amount: 1.0})
			case RETURN:
				g.returningPheromone.Insert(Pheromone{Vector: ant.Vector, amount: 1.0})
			}
		}

		nextAnts.Insert(ant)
	}

	nextAnts.Balance()
	g.ants = nextAnts
}

func (g *Game) updatePheromones() {
	nextForagingPheromone := kdtree.New(nil)
	nextReturningPheromone := kdtree.New(nil)

	for _, f := range g.foragingPheromone.Points() {
		foraging := f.(Pheromone)

		nextAmount := foraging.amount - PHEROMONE_DECAY
		if nextAmount > 0 {
			nextForagingPheromone.Insert(Pheromone{Vector: foraging.Vector, amount: nextAmount})
		}
	}

	for _, r := range g.returningPheromone.Points() {
		returning := r.(Pheromone)

		nextAmount := returning.amount - PHEROMONE_DECAY
		if nextAmount > 0 {
			nextReturningPheromone.Insert(Pheromone{Vector: returning.Vector, amount: nextAmount})
		}
	}

	nextForagingPheromone.Balance()
	nextReturningPheromone.Balance()
	g.foragingPheromone = nextForagingPheromone
	g.returningPheromone = nextReturningPheromone
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
