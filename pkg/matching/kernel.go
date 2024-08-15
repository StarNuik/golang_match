package matching

import (
	"context"
	"math"
	"time"

	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

func NewKernel(cfg KernelConfig) *Kernel {
	return &Kernel{
		cfg: cfg,
	}
}

type KernelConfig struct {
	MatchSize int
	GridSide  int
	// WaitTime1 time.Duration
	// WaitTime2 time.Duration
}

type Kernel struct {
	cfg KernelConfig
}

func (k *Kernel) Match(ctx context.Context, queue model.UserQueue) ([]schema.MatchResponse, error) {
	matches := []schema.MatchResponse{}
	some, err := k.pass0(ctx, queue)
	if err != nil {
		return nil, err
	}

	matches = append(matches, some...)
	return matches, nil
}

func (k *Kernel) pass0(ctx context.Context, users model.UserQueue) ([]schema.MatchResponse, error) {
	matches := []schema.MatchResponse{}

	limit := model.BinIdx{S: k.cfg.GridSide, L: k.cfg.GridSide}
	for s := 0; s < limit.S; s++ {
		for l := 0; l < limit.L; l++ {
			idx := model.BinIdx{S: s, L: l}

			bin, err := users.GetBin(ctx, idx)
			if err != nil {
				return nil, err
			}

			blockSize := k.cfg.MatchSize
			if len(bin) < blockSize {
				continue
			}

			// todo: sort the bin desc/asc by WaitTime ?
			blocks := len(bin) / blockSize
			for b := range blocks {
				match := make([]*model.QueuedUser, 0, blockSize)

				for uidx := range blockSize {
					idx := b*blockSize + uidx
					match = append(match, bin[idx])
				}

				matches = append(matches, fillResponse(match))
			}
		}
	}

	matchedUsers := make([]string, 0, len(matches)*k.cfg.MatchSize)
	for _, match := range matches {
		matchedUsers = append(matchedUsers, match.Names...)
	}

	err := users.Remove(ctx, matchedUsers)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func fillResponse(match []*model.QueuedUser) schema.MatchResponse {
	resp := schema.MatchResponse{}
	resp.Skill.Min = math.MaxFloat64
	resp.Latency.Min = math.MaxFloat64
	resp.WaitSeconds.Min = math.MaxFloat64

	now := time.Now().UTC()

	for _, user := range match {
		waitSeconds := now.Sub(user.QueuedAt).Seconds()

		resp.Names = append(resp.Names, user.Name)
		subfillCandle(&resp.Skill, user.Skill)
		subfillCandle(&resp.Latency, user.Latency)
		subfillCandle(&resp.WaitSeconds, waitSeconds)
	}

	finalizeCandle(&resp.Skill, len(match))
	finalizeCandle(&resp.Latency, len(match))
	finalizeCandle(&resp.WaitSeconds, len(match))

	return resp
}

func subfillCandle(c *schema.Candle, in float64) {
	c.Min = min(in, c.Min)
	c.Max = max(in, c.Max)
	c.Average += in
	c.Deviation += in * in
}

// https://stackoverflow.com/a/1175084
func finalizeCandle(c *schema.Candle, len int) {
	c.Average = c.Average / float64(len)
	c.Deviation = math.Sqrt(c.Deviation/float64(len) - c.Average*c.Average)
}
