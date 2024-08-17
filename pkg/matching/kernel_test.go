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

func TestKernelBasic(t *testing.T) {
	// requires "compose.test.yaml" deployed
	var dbUrl = "postgres://pg:insecure@localhost:5432/test"

	rangeKernels(dbUrl, func(label string, mFactory factoryModel, kFactory factoryKernel) {
		t.Run(label, func(t *testing.T) {
			require := require.New(t)

			ctx := context.Background()
			gridSide := 25
			datasetSize := 10_000
			matchSize := 8
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

			dataset := newDataset(mFactory, gcfg, datasetSize)

			kernel := kFactory(kcfg)
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
		})
	})
}

// func TestKernelTemporal(t *testing.T) {
// 	// requires "compose.test.yaml" deployed
// 	var dbUrl = "postgres://pg:insecure@localhost:5432/test"

// 	rangeKernels(dbUrl, func(kernelLabel string, mFactory factoryModel, kFactory factoryKernel) {
// 		// overwatch is ~75 ups
// 		rangeOver([]int{4, 8}, []int{5, 10, 25 /* , 50, 75 */}, func(usersPerTick, matchSize int) {
// 			require := require.New(t)

// 			ctx := context.Background()
// 			gridSide := 25
// 			gcfg := model.GridConfig{
// 				SkillCeil:   5000,
// 				LatencyCeil: 5000,
// 				Side:        gridSide,
// 			}
// 			kcfg := matching.KernelConfig{
// 				MatchSize:      matchSize,
// 				GridSide:       gridSide,
// 				WaitSoftLimit:  15 * time.Second,
// 				PriorityRadius: 2,
// 			}
// 			ticks := 900 // 15 minutes with 1 tps

// 			stats := onlineVariance{}
// 			kernel := kFactory(kcfg)
// 			dataset := newDataset(mFactory, gcfg, 0)

// 			label := fmt.Sprintf("%s_matchSize(%3d)_usersPerTick(%6d)", kernelLabel, matchSize, usersPerTick)
// 			t.Run(label, func(t *testing.T) {
// 				for range ticks {
// 					dataset.add(usersPerTick)
// 					dataset.moveQueuedAt(time.Second)

// 					matches, err := kernel.Match(ctx, dataset.model)
// 					require.Nil(err)

// 					stats.Push(dataset.dict, matches, matchSize)
// 					dataset.remove(matches)
// 				}
// 			})
// 			fmt.Printf("   %s\n", stats.Stat(false, true))
// 		})
// 	})
// }
