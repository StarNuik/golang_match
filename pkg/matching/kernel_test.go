package matching_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/starnuik/golang_match/pkg/matching"
	"github.com/starnuik/golang_match/pkg/schema"
)

func TestDbscanVariance(t *testing.T) {
	variance("dbscan", func(cfg matching.KernelConfig) matching.Kernel {
		return matching.NewDbscan(cfg)
	})
}

func TestFifoVariance(t *testing.T) {
	variance("fifo", func(cfg matching.KernelConfig) matching.Kernel {
		return matching.NewFifo(cfg)
	})
}

func variance(kernelName string, factory func(matching.KernelConfig) matching.Kernel) {
	sampleSize := 100

	rangeOver(func(userSize, matchSize int) {
		kernel := factory(matching.KernelConfig{
			MatchSize: matchSize,
		})
		dataset := dataset(userSize)
		stats := onlineVariance{}

		for range sampleSize {
			matches := kernel.Match(dataset)
			stats.Push(matches)
		}

		fmt.Printf("variance/%s: %s\n", label(kernelName, matchSize, userSize), stats.Stat())
	})
}

// https://www.statology.org/averaging-standard-deviations/
type onlineVariance struct {
	sumSkillSd   float64
	sumLatencySd float64
	n            int
}

func (s *onlineVariance) Push(matches []schema.MatchResponse) {
	for _, match := range matches {
		s.sumSkillSd += match.Skill.Deviation * match.Skill.Deviation
		s.sumLatencySd += match.Latency.Deviation * match.Latency.Deviation
	}
	s.n += len(matches)
}

func (s *onlineVariance) Stat() string {
	skillSd := math.Sqrt(s.sumSkillSd / float64(s.n))
	latencySd := math.Sqrt(s.sumLatencySd / float64(s.n))
	return fmt.Sprintf("AvgSkillSd(%f), AvgLatencySd(%f)", skillSd, latencySd)
}
