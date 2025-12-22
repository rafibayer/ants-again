package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rafibayer/ants-again/control"
	"github.com/rafibayer/ants-again/spatial"
	"github.com/rafibayer/ants-again/util"
	"github.com/rafibayer/ants-again/vector"
)

const (
	TPS       = 60
	GAME_SIZE = 1000
	ANTS      = 1000

	ANT_FOOD_RADIUS = GAME_SIZE / 200.0 // radius in which an ant will pick up food
	ANT_HILL_RADIUS = GAME_SIZE / 30.0  // radius in which an ant will return to hill
	ANT_BOUNDARY    = TURN              // if true, ants will wrap around to the other side instead of turning at boundaries

	FOOD_START = 50 // starting amount per food
)

// spatial hash densities
// these are fairly import perf knobs, especially for pheromones.
// if these are missized the cells either get too crowded or we have to search too many of them
const (
	PHEROMONE_HASH_CELL_SIZE = GAME_SIZE / 20.0
	FOOD_HASH_CELL_SIZE      = GAME_SIZE / 20.0
	HILL_HASH_CELL_SIZE      = GAME_SIZE / 5.0
)

type Game struct {
	params *Params

	frameCount, tickCount int

	camX, camY float64
	zoom       float64
	controls   control.State

	world *ebiten.Image
	px    []byte // pixel buffer: width * height * 4 (R,G,B,A)

	ants []*Ant
	food spatial.Spatial[*Food]

	hills         spatial.Spatial[vector.Vector]
	collectedFood int

	foragingPheromone  spatial.Spatial[*Pheromone]
	returningPheromone spatial.Spatial[*Pheromone]

	foragingAntCount   int
	returningAntCount  int
	remainingFoodCount int
}

func NewGame(params *Params) *Game {
	if params == nil {
		params = &DefaultParams
	}

	ants := []*Ant{}
	food := spatial.NewHash[*Food](FOOD_HASH_CELL_SIZE)
	hills := spatial.NewHash[vector.Vector](HILL_HASH_CELL_SIZE)

	for range ANTS {
		ants = append(ants, &Ant{
			Vector:          vector.Vector{X: GAME_SIZE / 2, Y: GAME_SIZE / 2},
			dir:             vector.Vector{X: util.Rand(-1, 1), Y: util.Rand(-1, 1)},
			state:           FORAGE,
			pheromoneStored: params.AntPheromoneStart,
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

		camX: 100,
		camY: 150,
		zoom: 0.5,

		world: ebiten.NewImage(GAME_SIZE, GAME_SIZE),
		px:    make([]byte, GAME_SIZE*GAME_SIZE*4), // pheromone buffer: 4 bytes per pixel (R,G,B,A)

		ants:  ants,
		food:  food,
		hills: hills,

		foragingPheromone:  spatial.NewHash[*Pheromone](PHEROMONE_HASH_CELL_SIZE),
		returningPheromone: spatial.NewHash[*Pheromone](PHEROMONE_HASH_CELL_SIZE),
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
