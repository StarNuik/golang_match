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

		priorityBin, err := users.GetRect(ctx, lo, hi, minWait)
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

		bin = combineBins(bin, priorityBin)
		slices.SortFunc(bin, func(l *model.QueuedUser, r *model.QueuedUser) int {
			return l.QueuedAt.Compare(r.QueuedAt)
		})

		some := matchBin(bin, blockSize)
		err = cleanMatches(ctx, users, some, k.MatchSize)
		if err != nil {
			return nil, err
		}

		matches = append(matches, some...)
	}
	return matches, nil
}

func combineBins(left []*model.QueuedUser, right []*model.QueuedUser) []*model.QueuedUser {
	unique := make(map[string]*model.QueuedUser)
	for _, user := range left {
		unique[user.Name] = user
	}
	for _, user := range right {
		unique[user.Name] = user
	}

	slice := make([]*model.QueuedUser, 0, len(unique))
	for _, user := range unique {
		slice = append(slice, user)
	}
	return slice
}

func cleanMatches(ctx context.Context, users model.UserQueue, matches []schema.MatchResponse, matchSize int) error {
	matchedUsers := make([]string, 0, len(matches)*matchSize)
	for _, match := range matches {
		matchedUsers = append(matchedUsers, match.Names...)
	}

	err := users.Remove(ctx, matchedUsers)
	if err != nil {
		return err
	}
	return nil
}
