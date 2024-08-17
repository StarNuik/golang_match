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

	ctx := context.Background()
	gridSide := 25
	datasetSize := 1_000
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

	for _, mFactory := range overModels(dbUrl) {
		for _, kFactory := range overKernels() {
			label := fmt.Sprintf("%s_%s", mFactory.label, kFactory.label)

			t.Run(label, func(t *testing.T) {
				require := require.New(t)

				dataset := newDataset(mFactory, gcfg, datasetSize)

				kernel := kFactory.build(kcfg)
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
		}
	}

}

func TestKernelSimulated(t *testing.T) {
	gridSide := 25
	gcfg := model.GridConfig{
		SkillCeil:   5000,
		LatencyCeil: 5000,
		Side:        gridSide,
	}
	mFactory := overModels("")[0]

	for _, kFactory := range overKernels() {
		for _, matchSize := range []int{4, 8, 16} {
			for _, usersPerTick := range []int{1, 2, 5, 10, 25, 50} {
				kcfg := matching.KernelConfig{
					MatchSize:      matchSize,
					GridSide:       gridSide,
					WaitSoftLimit:  15 * time.Second,
					PriorityRadius: 2,
				}
				kernel := kFactory.build(kcfg)
				dataset := newDataset(mFactory, gcfg, 0)
				label := fmt.Sprintf("%s_matchSize(%3d)_usersPerTick(%6d)", kFactory.label, matchSize, usersPerTick)

				stats := onlineVariance{}
				t.Run(label, func(t *testing.T) {
					stats = runSimulation(dataset, kernel, matchSize, usersPerTick)
				})
				fmt.Printf("   %s\n", stats.Stat(false, true))
			}
		}
	}
}

func TestKernelTuning(t *testing.T) {
	mFactory := overModels("")[0]
	kFactory := overKernels()[1]
	matchSize := 8
	usersPerTick := 10
	gridSide := 25
	priorityRadius := 1
	waitLimit := 60

	run := func(gridSide int, priorityRadius int, waitLimit int) {
		gcfg := model.GridConfig{
			SkillCeil:   5000,
			LatencyCeil: 5000,
			Side:        gridSide,
		}
		kcfg := matching.KernelConfig{
			MatchSize:      matchSize,
			GridSide:       gridSide,
			WaitSoftLimit:  time.Duration(waitLimit) * time.Second,
			PriorityRadius: priorityRadius,
		}
		kernel := kFactory.build(kcfg)
		dataset := newDataset(mFactory, gcfg, 0)
		label := fmt.Sprintf("gridSide(%d)_priorityRadius(%d)_waitLimitSec(%d)", gridSide, priorityRadius, waitLimit)

		stats := onlineVariance{}
		t.Run(label, func(t *testing.T) {
			stats = runSimulation(dataset, kernel, matchSize, usersPerTick)
		})
		fmt.Printf("   %s\n", stats.Stat(false, true))
	}

	run(25, 1, 15)
	run(25, 2, 15)
	for _, gridSide := range []int{5, 10, 25, 50, 75, 100} {
		run(gridSide, priorityRadius, waitLimit)
	}
	for _, priorityRadius := range []int{1, 2, 4, 5, 8, 10} {
		run(gridSide, priorityRadius, waitLimit)
	}
	for _, waitLimit := range []int{5, 10, 15, 30, 45, 60, 120} {
		run(gridSide, priorityRadius, waitLimit)
	}
}

func runSimulation(dataset *dataset, kernel matching.Kernel, matchSize int, usersPerTick int) onlineVariance {
	ctx := context.Background()
	ticks := 600 // 10 minutes with 1 tps

	stats := onlineVariance{}
	for range ticks {
		dataset.add(usersPerTick)
		dataset.moveQueuedAt(time.Second)

		matches, err := kernel.Match(ctx, dataset.model)
		if err != nil {
			panic(err)
		}

		stats.Push(dataset.dict, matches, matchSize)
		dataset.remove(matches)
	}

	return stats
}
