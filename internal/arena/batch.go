package arena

import (
	"sort"
	"sync"
	"sync/atomic"
)

func RunMatches(options BatchOptions, runMatch func(simulationID int, seed int64) MatchResult) []MatchResult {
	workers := options.Parallel
	if workers > options.Simulations {
		workers = options.Simulations
	}
	if workers < 1 {
		workers = 1
	}

	var counter uint64
	results := make(chan MatchResult, options.Simulations)

	var wg sync.WaitGroup
	var workerPanic interface{}
	var panicMu sync.Mutex

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if recovered := recover(); recovered != nil {
					panicMu.Lock()
					workerPanic = recovered
					panicMu.Unlock()
				}
			}()
			for {
				id := int(atomic.AddUint64(&counter, 1) - 1)
				if id >= options.Simulations {
					return
				}
				seed := deriveSeed(options.Seed, uint64(id), options.SeedIncrement)
				results <- runMatch(id, seed)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	all := make([]MatchResult, 0, options.Simulations)
	for result := range results {
		all = append(all, result)
	}

	if workerPanic != nil {
		panic(workerPanic)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].SimulationID() < all[j].SimulationID()
	})
	return all
}

func deriveSeed(base int64, offset uint64, seedIncrement *int64) int64 {
	if seedIncrement != nil {
		return int64(uint64(base) + offset*uint64(*seedIncrement))
	}
	return mixSeed(base, offset)
}

func mixSeed(base int64, offset uint64) int64 {
	z := uint64(base)
	if offset == 0 {
		return base
	}
	z += offset + 0x9E3779B97F4A7C15
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return int64(z ^ (z >> 31))
}
