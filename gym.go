// the idea of this package is to perform a very basic form of parameter optimization for the game
// we will run the game without rendering or delaying frames, and with random parameters.
// we will record which do well (measured by collected food) after some elapsed (sim) time
package main

import (
	"log"
	"slices"
)

const (
	// run for 1 simulated minutes
	GYM_SIM_TIME = 60 * TPS
	GYM_SAMPLES  = 4
)

func runGym() error {
	var bestCollected int
	var bestParams Params
	var bestStats Stats

	defer func() {
		log.Printf("best: %d food collected\nparams: %+v\nstats: %+v",
			bestCollected, bestParams, bestStats)
	}()

	for iteration := 0; true; iteration++ {

		log.Printf("====== [%d] ======", iteration)

		params := Params{
			AntSpeed:             Rand(0.5, 7.0),
			AntRotation:          Rand(0.0, 20.0),
			PheromoneSenseRadius: Rand(GAME_SIZE/50, GAME_SIZE/4),
			PheromoneDecay:       float32(Rand(1/120.0, 1/1.0)),
			PheromoneDropProb:    Rand(1/120.0, 1/1.0),
			PheromoneInfluence:   Rand(0.1, 10.0),
			PheromoneSenseProb:   Rand(0.05, 1.0),
		}

		scores := make([]int, 0, GYM_SAMPLES)
		statsList := make([]Stats, 0, GYM_SAMPLES)

		for range GYM_SAMPLES {
			game := NewGame(nil)
			for range GYM_SIM_TIME {
				if err := game.Update(); err != nil {
					log.Printf("[%d] error: %v\nparams: %+v", iteration, err, params)
				}
			}

			st := game.Stats()
			scores = append(scores, st.food.collected)
			statsList = append(statsList, *st)
		}

		medianScore, medianStats := medianSample(scores, statsList)

		if medianScore > bestCollected {
			bestCollected = medianScore
			bestParams = params
			bestStats = medianStats

			// compute min and max only for best-case iterations
			minScore := scores[0]
			maxScore := scores[0]
			for _, sc := range scores[1:] {
				if sc < minScore {
					minScore = sc
				}
				if sc > maxScore {
					maxScore = sc
				}
			}

			log.Printf("[%d] New Best: %d", iteration, medianScore)
			log.Printf("[%d] Scores: min=%d median=%d max=%d",
				iteration, minScore, medianScore, maxScore)
			log.Printf("Params: %+v", params)
			log.Printf("Stats (median sample): %+v", medianStats)
		}
	}

	return nil
}

// medianSample returns the median score and the Stats belonging to
// the actual sample that produced that score.
func medianSample(scores []int, stats []Stats) (int, Stats) {
	type pair struct {
		val int
		idx int
	}

	n := len(scores)
	arr := make([]pair, n)
	for i, v := range scores {
		arr[i] = pair{v, i}
	}

	// sort by score only
	slices.SortFunc(arr, func(a, b pair) int {
		return a.val - b.val
	})

	mid := n / 2
	var medianIndex int
	var medianScore int

	if n%2 == 1 {
		medianScore = arr[mid].val
		medianIndex = arr[mid].idx
	} else {
		// average for median, pick upper-mid as representative
		medianScore = (arr[mid-1].val + arr[mid].val) / 2
		medianIndex = arr[mid].idx
	}

	return medianScore, stats[medianIndex]
}
