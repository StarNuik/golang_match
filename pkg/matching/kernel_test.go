package matching_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/starnuik/golang_match/pkg/matching"
	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
	"github.com/stretchr/testify/require"
)

var (
	ctx      = context.Background()
	gridSide = 25
	gcfg     = model.GridConfig{
		SkillCeil:   5000,
		LatencyCeil: 5000,
		Side:        gridSide,
	}
)

func TestKernelBasic(t *testing.T) {
	require := require.New(t)

	datasetSize := 10_000
	matchSize := 8
	kcfg := matching.KernelConfig{
		MatchSize: matchSize,
		GridSide:  gridSide,
	}
	dataset := newDataset(gcfg, datasetSize)

	kernel := matching.NewKernel(kcfg)
	require.NotNil(kernel)

	matches, err := kernel.Match(ctx, dataset.model)
	require.Nil(err)
	require.NotNil(matches)
	// a joke: this condition has a non-zero probability of returning a false-negative
	require.True(len(matches) > 0)

	names := make(map[string]struct{})
	userOverlaps := 0
	for _, match := range matches {
		require.Equal(matchSize, len(match.Names))

		for _, name := range match.Names {
			if _, exists := names[name]; exists {
				userOverlaps += 1
			}
			names[name] = struct{}{}
		}
	}

	matchedCount := len(matches) * matchSize
	fmt.Printf("from(%d), matched(%d)\n", datasetSize, matchedCount)

	require.Zero(userOverlaps, fmt.Sprintf("overlaps(%d)", userOverlaps))
}

// func TestKernelTemporal(t *testing.T) {
// 	rangeOver([]int{4, 8, 16}, []int{10, 50, 100, 250, 500, 1000}, func(datasetSize, matchSize int) {
// 		//
// 	})
// }

func BenchmarkKernel(b *testing.B) {
	b.StopTimer()
	rangeOver([]int{4, 8, 16}, []int{100, 1_000, 10_000}, func(datasetSize, matchSize int) {
		stats := onlineVariance{}
		kcfg := matching.KernelConfig{
			MatchSize: matchSize,
			GridSide:  gridSide,
		}
		label := fmt.Sprintf("_matchSize(%3d)_userbase(%6d)", matchSize, datasetSize)

		b.Run(label, func(b *testing.B) {
			for range b.N {
				dataset := newDataset(gcfg, datasetSize)
				kernel := matching.NewKernel(kcfg)

				b.StartTimer()
				matches, err := kernel.Match(ctx, dataset.model)
				b.StopTimer()

				if err != nil {
					log.Panicln(err)
				}
				stats.Push(dataset.dict, matches, matchSize)
			}
		})
		fmt.Printf("   %s\n", stats.Stat())
	})
}

type dataset struct {
	model model.UserQueue
	dict  map[string]*model.QueuedUser
}

func newDataset(cfg model.GridConfig, size int) *dataset {
	dict := make(map[string]*model.QueuedUser)
	model := model.NewUserQueueInmemory(gcfg)

	for range size {
		user := randomUser()
		if _, exists := dict[user.Name]; exists {
			continue
		}
		dict[user.Name] = user
		model.Add(context.Background(), user)
	}

	return &dataset{
		model: model,
		dict:  dict,
	}
}

func randomUser() *model.QueuedUser {
	nameBytes := make([]byte, 16)
	rand.Read(nameBytes)
	name := hex.EncodeToString(nameBytes)

	out := model.QueuedUser{
		Name:     name,
		Skill:    rand.NormFloat64()*800.0 + 2500.0, // mean(2500), sd(800)
		Latency:  rand.NormFloat64()*250.0 + 500.0,  // mean(500), sd(250)
		QueuedAt: time.Now().UTC(),
	}
	return &out
}

func rangeOver(matchSizes []int, userSizes []int, call func(datasetSize int, matchSize int)) {
	for _, uSize := range userSizes {
		for _, mSize := range matchSizes {
			call(uSize, mSize)
		}
	}
}

// statistical info for a kernel (hacky & crude)
type onlineVariance struct {
	sumSkillSd   float64
	sumLatencySd float64
	sumMatched   float64
	n            int
	nn           int
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
	s.sumMatched += float64(len(matches)*matchSize) / float64(len(users))
	s.nn++
}

func (s *onlineVariance) Stat() string {
	skillSd := math.Sqrt(s.sumSkillSd / float64(s.n))
	latencySd := math.Sqrt(s.sumLatencySd / float64(s.n))
	matchRate := s.sumMatched / float64(s.nn)
	return fmt.Sprintf("AvgSkillSd(%5.1f), AvgLatencySd(%5.1f), AvgMatchRate(%5.2f)", skillSd, latencySd, matchRate)
}
