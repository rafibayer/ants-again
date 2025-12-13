package main

type Params struct {
	AntSpeed          float64 // ant movement per tick (suggested: 2.0)
	AntRotation       float64 // random ant rotation in either direction per tick (suggested: 9.0)
	AntPheromoneStart int     // ant pheromone "inventory" (suggested: 30)

	PheromoneSenseRadius           float64 // ant pheromone sense radius. expensive. (suggested: (game_size / 8.0))
	PheromoneSenseCosineSimilarity float64 // if cost(theta) of angle between ant direction and pheromone is < this threshold, the pheromone is ignored. (suggested: 0.33)
	PheromoneDecay                 float32 // pheromone decay per tick. pheromones start with value of 1.0 (suggested: 1.0 / (10 * TPS))
	PheromoneDropProb              float64 // probability for an ant to drop pheromone per tick. (suggested: 0.008333 = 1.0 / (2 * TPS))
	PheromoneInfluence             float64 // pheromone influence multiplier (suggested: 2.0)
	PheromoneSenseProb             float64 // probability of an ant sensing pheromones per tick. expensive. (suggested: 1.0 / 4.0)
}

// Default parameters if nil is passed to NewGame.
var DefaultParams = Params{
	AntSpeed:                       1.8,
	AntRotation:                    9.0,
	AntPheromoneStart:              20,
	PheromoneSenseRadius:           GAME_SIZE / 10.0,
	PheromoneSenseCosineSimilarity: 0.33,
	PheromoneDecay:                 1.0 / (10 * TPS),
	PheromoneDropProb:              1.0 / (TPS),
	PheromoneInfluence:             3.0,
	PheromoneSenseProb:             1.0 / 4,
}
