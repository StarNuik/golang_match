package matching

import (
	"math"
	"time"

	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

type Kernel interface {
	Match([]*model.QueuedUser) []schema.MatchResponse
}

type KernelConfig struct {
	MatchSize int
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

	resp.Skill.Average /= float64(len(match))
	resp.Latency.Average /= float64(len(match))
	resp.WaitSeconds.Average /= float64(len(match))

	return resp
}

func subfillCandle(c *schema.Candle, in float64) {
	c.Min = min(in, c.Min)
	c.Max = max(in, c.Max)
	c.Average += in
}
