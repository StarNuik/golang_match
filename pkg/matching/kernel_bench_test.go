package matching_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/starnuik/golang_match/pkg/matching"
	"github.com/starnuik/golang_match/pkg/model"
)

func BenchmarkKernel(b *testing.B) {
	// requires "compose.test.yaml" deployed
	var dbUrl = "postgres://pg:insecure@localhost:5432/test"

	b.StopTimer()
	for _, mFactory := range overModels(dbUrl) {
		for _, kFactory := range overKernels() {
			ctx := context.Background()
			gridSide := 25
			matchSize := 8
			population := 1000

			gcfg := model.GridConfig{
				SkillCeil:   5000,
				LatencyCeil: 5000,
				Side:        gridSide,
			}
			kcfg := matching.KernelConfig{
				MatchSize:      25,
				GridSide:       gridSide,
				WaitSoftLimit:  15 * time.Second,
				PriorityRadius: 2,
			}

			stats := onlineVariance{}
			label := fmt.Sprintf("%s_%s_matchSize(%3d)_population(%6d)", mFactory.label, kFactory.label, matchSize, population)

			b.Run(label, func(b *testing.B) {
				for range b.N {
					dataset := newDataset(mFactory, gcfg, population)
					kernel := kFactory.build(kcfg)

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
		}

	}
}
