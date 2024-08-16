package matching

import (
	"context"

	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

func NewBasicKernel(cfg KernelConfig) Kernel {
	return &basicKernel{
		cfg,
	}
}

type basicKernel struct {
	KernelConfig
}

func (k *basicKernel) Match(ctx context.Context, users model.UserQueue) ([]schema.MatchResponse, error) {
	matches := []schema.MatchResponse{}

	some, err := k.pass0(ctx, users)
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

func (k *basicKernel) pass0(ctx context.Context, users model.UserQueue) ([]schema.MatchResponse, error) {
	matches := []schema.MatchResponse{}

	for _, idx := range binIndices(k.GridSide) {
		bin, err := users.GetBin(ctx, idx)
		if err != nil {
			return nil, err
		}

		blockSize := k.MatchSize
		if len(bin) < blockSize {
			continue
		}

		// todo: sort the bin desc/asc by WaitTime ? -- no difference found in simulations
		// slices.SortFunc(bin, func(l *model.QueuedUser, r *model.QueuedUser) int {
		// 	return l.QueuedAt.Compare(r.QueuedAt)
		// })

		matches = append(matches, matchBin(bin, blockSize)...)
	}

	return matches, nil
}
