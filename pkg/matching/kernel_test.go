package matching_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/starnuik/golang_match/pkg/matching"
	"github.com/starnuik/golang_match/pkg/model"
	"github.com/stretchr/testify/require"
)

func TestKernelBasic(t *testing.T) {
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
