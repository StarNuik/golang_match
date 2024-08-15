package matching_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/starnuik/golang_match/pkg/matching"
	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

func BenchmarkDbscan(b *testing.B) {
	benchmark(b, "dbscan", func(cfg matching.KernelConfig) matching.Kernel {
		return matching.NewDbscan(cfg)
	})
}

func BenchmarkFifo(b *testing.B) {
	benchmark(b, "fifo", func(cfg matching.KernelConfig) matching.Kernel {
		return matching.NewFifo(cfg)
	})
}

func benchmark(b *testing.B, kernelName string, factory func(matching.KernelConfig) matching.Kernel) {
	rangeOver(func(userSize, matchSize int) {
		b.StopTimer()
		kernel := factory(matching.KernelConfig{
			MatchSize: matchSize,
		})

		stats := onlineVariance{}
		// var avgClusterSize float64
		// var clusterN float64
		b.Run(label(kernelName, matchSize, userSize), func(b *testing.B) {
			for range b.N {
				dataset := newDataset(userSize)
				b.StartTimer()
				matches := kernel.Match(dataset.slice)
				b.StopTimer()
				stats.Push(dataset.dict, matches, matchSize)
				// avgClusterSize += kernel.(*matching.DbscanKernel).DEBUG_avgClusterSize
				// clusterN += 1
			}
		})
		// avgClusterSize /= clusterN
		fmt.Printf("   %s\n", stats.Stat())
	})
}

func rangeOver(call func(userSize int, matchSize int)) {
	matchSizes := []int{4, 8, 16}
	userSizes := []int{100, 1_000, 10_000 /* 100_000, */}

	for _, uSize := range userSizes {
		for _, mSize := range matchSizes {
			call(uSize, mSize)
		}
	}
}

func label(kernelName string, mSize int, uSize int) string {
	return fmt.Sprintf("%s_matchSize(%3d)_userbase(%6d)", kernelName, mSize, uSize)
}

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
