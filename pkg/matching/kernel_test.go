package matching_test

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/starnuik/golang_match/pkg/matching"
// 	"github.com/stretchr/testify/require"
// )

// func TestGridBasic(t *testing.T) {
// 	basicTests(t, func(cfg matching.KernelConfig) matching.Kernel {
// 		return matching.NewGrid(cfg)
// 	})
// }

// func TestDbscanBasic(t *testing.T) {
// 	basicTests(t, func(cfg matching.KernelConfig) matching.Kernel {
// 		return matching.NewDbscan(cfg)
// 	})
// }

// func TestFifoBasic(t *testing.T) {
// 	basicTests(t, func(cfg matching.KernelConfig) matching.Kernel {
// 		return matching.NewFifo(cfg)
// 	})
// }

// func basicTests(t *testing.T, factory func(matching.KernelConfig) matching.Kernel) {
// 	require := require.New(t)

// 	matchSize := 8
// 	dataset := newDataset(10_000)
// 	kernel := factory(matching.KernelConfig{
// 		MatchSize:   matchSize,
// 		SkillCeil:   5000,
// 		LatencyCeil: 1000,
// 	})
// 	matches := kernel.Match(dataset.slice)
// 	// beware: this has a non-zero probability of happening
// 	require.True(len(matches) > 0)

// 	names := make(map[string]struct{})
// 	nameOverlaps := 0
// 	usersMatched := 0
// 	for _, match := range matches {
// 		require.Equal(matchSize, len(match.Names))

// 		for _, name := range match.Names {
// 			if _, exists := names[name]; exists {
// 				nameOverlaps += 1
// 			}
// 			usersMatched += 1
// 			names[name] = struct{}{}
// 		}
// 	}
// 	fmt.Printf("usersMatched(%d), overlaps(%d)\n", usersMatched, nameOverlaps)
// 	require.Equal(0, nameOverlaps)

// 	//
// }
