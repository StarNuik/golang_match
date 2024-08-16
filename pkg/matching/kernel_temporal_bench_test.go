package matching_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/starnuik/golang_match/pkg/matching"
	"github.com/starnuik/golang_match/pkg/model"
	"github.com/stretchr/testify/require"
)

func TestKernelTemporal(t *testing.T) {
	rangeKernels(func(kernelLabel string, factory func(matching.KernelConfig) matching.Kernel) {
		// overwatch is ~75 ups
		rangeOver([]int{4, 8, 16}, []int{5, 10, 25, 50, 75}, func(usersPerTick, matchSize int) {
			require := require.New(t)

			ctx := context.Background()
			gridSide := 25
			gcfg := model.GridConfig{
				SkillCeil:   5000,
				LatencyCeil: 5000,
				Side:        gridSide,
			}
			kcfg := matching.KernelConfig{
				MatchSize:      matchSize,
				GridSide:       gridSide,
				WaitSoftLimit:  15 * time.Second,
				PriorityRadius: 2,
			}
			ticks := 900 // 15 minutes with 1 tps

			stats := onlineVariance{}
			kernel := factory(kcfg)
			dataset := newDataset(gcfg, 0)

			label := fmt.Sprintf("%s_matchSize(%3d)_usersPerTick(%6d)", kernelLabel, matchSize, usersPerTick)
			t.Run(label, func(t *testing.T) {
				for range ticks {
					dataset.add(usersPerTick)
					dataset.moveQueuedAt(time.Second)

					matches, err := kernel.Match(ctx, dataset.model)
					require.Nil(err)

					stats.Push(dataset.dict, matches, matchSize)
					dataset.remove(matches)
				}
			})
			fmt.Printf("   %s\n", stats.Stat(false, true))
		})
	})
}
