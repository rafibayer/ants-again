package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kyroy/kdtree"
)

const (
	GAME_SIZE = 1000

	ANT_SPEED = 3.0

	PHEROMONE_DECAY = (1.0 / 60.0) / 10.0 // denominator is number of seconds until decay

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

type Game struct {
	frameCount, updateCount uint64

	camX, camY float64
	zoom       float64

	world *ebiten.Image
	px    []byte // RGBA buffer: width * height * 4

	food  *kdtree.KDTree
	ants  *kdtree.KDTree
	hills *kdtree.KDTree

	foragingPheromone  [GAME_SIZE][GAME_SIZE]float32
	returningPheromone [GAME_SIZE][GAME_SIZE]float32
}

func NewGame() *Game {
	ants := kdtree.New(nil)
	food := kdtree.New(nil)

	for range 1000 {
		ants.Insert(Ant{
			Vector: Vector{GAME_SIZE / 2, GAME_SIZE / 2},
			dir:    Vector{Rand(-1, 1), Rand(-1, 1)},
			state:  FORAGE,
		})
	}

	for r := range 10 {
		for c := range 10 {
			food.Insert(&Food{
				Vector: &Vector{x: 300.0 + float64(r), y: 300.0 + float64(c)},
				amount: FOOD_START,
			})
		}
	}

	return &Game{
		frameCount:  0,
		updateCount: 0,

		camX: 0,
		camY: 0,
		zoom: 1.0,

		world: ebiten.NewImage(GAME_SIZE, GAME_SIZE),
		px:    make([]byte, GAME_SIZE*GAME_SIZE*4), // pheromone buffer: 4 bytes per pixel (R,G,B,A)

		ants: ants,
		food: food,

		foragingPheromone:  [GAME_SIZE][GAME_SIZE]float32{},
		returningPheromone: [GAME_SIZE][GAME_SIZE]float32{},
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

		if ant.state == FORAGE {
			// check for food nearby
			nearFood := FindWithin(g.food, ant.Vector, 3)
			for _, f := range nearFood {
				food := f.(*Food)
				if food.amount > 0 {
					food.amount--
				}

				ant.state = RETURN
			}
		}

		row := int(math.Round(ant.y))
		col := int(math.Round(ant.x))

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

		// check if ant inbounds -- can prob remove if we prevent them from going oob somehow
		if row >= 0 && row < GAME_SIZE && col >= 0 && col < GAME_SIZE {
			switch ant.state {
			case FORAGE:
				// do we need to do this in "nextForagingPheromone" or something?
				// currently, ants would smell this pheromone the same frame that it's placed, fine?
				g.foragingPheromone[row][col] = 1.0
			case RETURN:
				g.returningPheromone[row][col] = 1.0
			}
		}

		nextAnts.Insert(ant)
	}

	nextAnts.Balance()
	g.ants = nextAnts
}

func (g *Game) updatePheromones() {
	for r := range GAME_SIZE {
		for c := range GAME_SIZE {
			foraging := g.foragingPheromone[r][c]
			returning := g.returningPheromone[r][c]
			g.foragingPheromone[r][c] = max(foraging-PHEROMONE_DECAY, 0.0)
			g.returningPheromone[r][c] = max(returning-PHEROMONE_DECAY, 0.0)
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
