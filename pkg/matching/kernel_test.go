package matching_test

import (
	"fmt"
	"math"

	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

// func TestDbscanVariance(t *testing.T) {
// 	variance("dbscan", func(cfg matching.KernelConfig) matching.Kernel {
// 		return matching.NewDbscan(cfg)
// 	})
// }

// func TestFifoVariance(t *testing.T) {
// 	variance("fifo", func(cfg matching.KernelConfig) matching.Kernel {
// 		return matching.NewFifo(cfg)
// 	})
// }

// func variance(kernelName string, factory func(matching.KernelConfig) matching.Kernel) {
// 	sampleSize := 100

// 	rangeOver(func(userSize, matchSize int) {
// 		kernel := factory(matching.KernelConfig{
// 			MatchSize: matchSize,
// 		})
// 		dataset := dataset(userSize)
// 		stats := onlineVariance{}

// 		for range sampleSize {
// 			matches := kernel.Match(dataset)
// 			stats.Push(dataset, matches, matchSize)
// 		}

// 		fmt.Printf("variance/%s: %s\n", label(kernelName, matchSize, userSize), stats.Stat())
// 	})
// }

type onlineVariance struct {
	sumSkillSd   float64
	sumLatencySd float64
	n            int
}

func (s *onlineVariance) Push(users map[string]*model.QueuedUser, matches []schema.MatchResponse, matchSize int) {
	for _, match := range matches {
		// https://stackoverflow.com/a/1175084
		var skillMean, latencyMean, skillSd, latencySd float64
		for _, username := range match.Names {
			user := users[username]
			skillMean += user.Skill
			skillSd += user.Skill * user.Skill
			latencyMean += user.Latency
			latencySd += user.Latency * user.Latency
		}
		skillMean = skillMean / float64(matchSize)
		skillSd = math.Sqrt(skillSd/float64(matchSize) - skillMean*skillMean)
		latencyMean = latencyMean / float64(matchSize)
		latencySd = math.Sqrt(latencySd/float64(matchSize) - latencyMean*latencyMean)

		// https://www.statology.org/averaging-standard-deviations/
		s.sumSkillSd += skillSd * skillSd
		s.sumLatencySd += latencySd * latencySd
	}
	s.n += len(matches)
}

func (s *onlineVariance) Stat() string {
	skillSd := math.Sqrt(s.sumSkillSd / float64(s.n))
	latencySd := math.Sqrt(s.sumLatencySd / float64(s.n))
	return fmt.Sprintf("AvgSkillSd(%f), AvgLatencySd(%f)", skillSd, latencySd)
}
