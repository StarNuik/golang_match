package matching

import (
	"context"
	"slices"
	"time"

	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

func NewPriorityKernel(cfg KernelConfig) Kernel {
	return &priorityKernel{
		cfg,
	}
}

type priorityKernel struct {
	KernelConfig
}

func (k *priorityKernel) Match(ctx context.Context, users model.UserQueue) ([]schema.MatchResponse, error) {
	matches := []schema.MatchResponse{}

	some, err := k.passX(ctx, users, k.PriorityRadius, k.WaitSoftLimit)
	if err != nil {
		return nil, err
	}
	err = cleanMatches(ctx, users, some, k.MatchSize)
	if err != nil {
		return nil, err
	}
	matches = append(matches, some...)

	return matches, nil
}

func (k *priorityKernel) passX(ctx context.Context, users model.UserQueue, kernelSize int, minWait time.Duration) ([]schema.MatchResponse, error) {
	matches := []schema.MatchResponse{}
	for _, idx := range binIndices(k.GridSide) {
		lo := model.BinIdx{
			S: idx.S - kernelSize,
			L: idx.L - kernelSize,
		}
		hi := model.BinIdx{
			S: idx.S + kernelSize,
			L: idx.L + kernelSize,
		}

		priorityBin, err := users.GetBins(ctx, lo, hi, minWait)
		if err != nil {
			return nil, err
		}

		bin, err := users.GetBin(ctx, idx)
		if err != nil {
			return nil, err
		}

		blockSize := k.MatchSize
		total := len(priorityBin) + len(bin)
		if total < blockSize {
			continue
		}

		bin = append(bin, priorityBin...)
		slices.SortFunc(bin, func(l *model.QueuedUser, r *model.QueuedUser) int {
			return l.QueuedAt.Compare(r.QueuedAt)
		})
		matches = append(matches, matchBin(bin, blockSize)...)
	}
	return matches, nil
}
