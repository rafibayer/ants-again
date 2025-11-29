package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kyroy/kdtree"
)

const (
	GAME_SIZE = 500

	ANT_SPEED = 0.5

	PHEROMONE_DECAY              = (1.0 / 60.0) / 10.0 // denominator is number of seconds until decay
	PHEROMONE_INFLUENCE_DISCOUNT = 0.01                // reduce effect of pheromone on ant direction
	PHEROMONE_PENDING_TICKS      = 45                  // how long before a pheromone becomes active

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
	frameCount, tickCount int

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
	hills := kdtree.New(nil)

	for range 300 {
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

		foragingPheromone:  [GAME_SIZE][GAME_SIZE]float32{},
		returningPheromone: [GAME_SIZE][GAME_SIZE]float32{},
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
	for _, a := range g.ants.Points() {
		ant := a.(Ant)
		ant.Vector = ant.Add(ant.dir.Normalize().Mul(ANT_SPEED))

		// randomly rotate a few degrees
		ant.dir = ant.dir.Rotate(Rand(-2.5, 2.5))

		row, col := ant.ToGrid()
		keepInbounds(&ant, row, col)

		front := ant.Vector.Add(ant.dir.Normalize().Mul(3.0))

		// pheromone field to search based on ant state
		pheromone := &g.returningPheromone
		if ant.state == RETURN {
			pheromone = &g.foragingPheromone
		}

		// influence direction based on pheromone
		pheromoneDir := ZERO
		VisitCircle(pheromone, front.y, front.x, 1.5, func(value float32, weight float64, r, c int) {
			// direction to pheromone and signal strength
			spot := Vector{x: float64(c), y: float64(r)}
			dirToSpot := spot.Sub(ant.Vector).Normalize()

			// discount by weight and distance from ant
			strength := (float64(value) * weight) / (1 + ant.Vector.Distance(spot))
			pheromoneDir = pheromoneDir.Add(dirToSpot.Mul(strength))
		})

		ant.dir = ant.dir.Add(pheromoneDir.Mul(PHEROMONE_INFLUENCE_DISCOUNT)).Normalize()

		if ant.state == FORAGE {
			// check for food in front, change state and turn around
			nearFood := KDSearchRadius(g.food, front, 3)
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
			nearHill := KDSearchRadius(g.hills, ant.Vector, 15.0)
			if len(nearHill) > 0 {
				// turn around and go back to foraging
				ant.state = FORAGE
				ant.dir = ant.dir.Mul(-1.0)
			}
		}

		// check if ant inbounds -- can prob remove if we prevent them from going oob somehow
		if row >= 0 && row < GAME_SIZE && col >= 0 && col < GAME_SIZE {
			// append to queue, becomes active after delay
			switch ant.state {
			case FORAGE:
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
	// decay pheromones
	for r := range GAME_SIZE {
		for c := range GAME_SIZE {
			foraging := g.foragingPheromone[r][c]
			returning := g.returningPheromone[r][c]
			g.foragingPheromone[r][c] = max(foraging-PHEROMONE_DECAY, 0.0)
			g.returningPheromone[r][c] = max(returning-PHEROMONE_DECAY, 0.0)
		}
	}
}

func (g *Game) updateFood() {
	nextFood := kdtree.New(nil)
	for _, f := range g.food.Points() {
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
