package matching_test

import (
	"testing"

	"github.com/starnuik/golang_match/pkg/matching"
)

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
		dataset := dataset(userSize)
		b.StartTimer()

		b.Run(label(kernelName, matchSize, userSize), func(b *testing.B) {
			for range b.N {
				kernel.Match(dataset)
			}
		})
	})
}
