package matching_test

import (
	"fmt"
	"testing"

	"github.com/starnuik/golang_match/pkg/matching"
)

func BenchmarkDbscan(b *testing.B) {
	benchmark(b, "dbscan", func(cfg matching.KernelConfig) matching.Kernel {
		return matching.NewDbscan(cfg)
	})
}

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

		stats := onlineVariance{}
		b.Run(label(kernelName, matchSize, userSize), func(b *testing.B) {
			for range b.N {
				dataset := newDataset(userSize)
				b.StartTimer()
				matches := kernel.Match(dataset.slice)
				b.StopTimer()
				stats.Push(dataset.dict, matches, matchSize)
			}
		})
		fmt.Println("   ", stats.Stat())
	})
}
