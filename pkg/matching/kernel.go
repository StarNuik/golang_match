package matching

import (
	"context"
	"math"
	"time"

	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

type KernelConfig struct {
	MatchSize      int
	GridSide       int
	PriorityRadius int
	WaitSoftLimit  time.Duration
}

type Kernel interface {
	Match(ctx context.Context, users model.UserQueue) ([]schema.MatchResponse, error)
}

func binIndices(side int) []model.BinIdx {
	slice := make([]model.BinIdx, 0, side*side)
	limit := model.BinIdx{S: side, L: side}
	for s := 0; s < limit.S; s++ {
		for l := 0; l < limit.L; l++ {
			slice = append(slice, model.BinIdx{S: s, L: l})
		}
	}
	return slice
}

func matchBin(bin []*model.QueuedUser, blockSize int) []schema.MatchResponse {
	matches := []schema.MatchResponse{}

	blocks := len(bin) / blockSize

	for b := range blocks {
		match := make([]*model.QueuedUser, 0, blockSize)

		for uidx := range blockSize {
			idx := b*blockSize + uidx
			match = append(match, bin[idx])
		}

		matches = append(matches, fillResponse(match))
	}
	return matches
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
