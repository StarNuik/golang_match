package matching_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/starnuik/golang_match/pkg/matching"
	"github.com/starnuik/golang_match/pkg/model"
)

func BenchmarkKernel(b *testing.B) {
	b.StopTimer()

	ctx := context.Background()
	gridSide := 25
	gcfg := model.GridConfig{
		SkillCeil:   5000,
		LatencyCeil: 5000,
		Side:        gridSide,
	}

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
		fmt.Printf("   %s\n", stats.Stat(true, false))
	})
}
