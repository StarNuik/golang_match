package matching

import (
	"math"

	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

func NewGrid(cfg KernelConfig) Kernel {
	return &gridKernel{
		cfg: cfg,
	}
}

type gridKernel struct {
	cfg KernelConfig
}

type binIndex struct {
	s int
	l int
}

func (k *gridKernel) Match(queue []*model.QueuedUser) []schema.MatchResponse {
	binsPerSide := 25

	bins := make(map[binIndex]map[string]*model.QueuedUser)
	for _, user := range queue {
		idx := k.toIndex(user, binsPerSide)
		if bins[idx] == nil {
			bins[idx] = make(map[string]*model.QueuedUser)
		}
		bins[idx][user.Name] = user
	}

	matches := []schema.MatchResponse{}
	limit := binIndex{binsPerSide, binsPerSide}
	for s := 0; s < limit.s; s++ {
		for l := 0; l < limit.l; l++ {
			idx := binIndex{s, l}

			for len(bins[idx]) >= k.cfg.MatchSize {
				match := []*model.QueuedUser{}
				i := 0
				for _, user := range bins[idx] {
					match = append(match, user)
					i++
					if i >= k.cfg.MatchSize {
						break
					}
				}

				resp := fillResponse(match)
				matches = append(matches, resp)

				for _, name := range resp.Names {
					delete(bins[idx], name)
				}
			}
		}
	}

	return matches
}

func (k *gridKernel) toIndex(user *model.QueuedUser, sides int) binIndex {
	return binIndex{
		s: remap(user.Skill, k.cfg.SkillCeil, sides),
		l: remap(user.Latency, k.cfg.LatencyCeil, sides),
	}
}

func remap(value float64, ceil float64, sides int) int {
	value = value / ceil
	value = value * float64(sides)
	value = math.Floor(value)
	return clamp0n(int(value), sides-1)
}

func clamp0n(value int, n int) int {
	return min(n, max(0, value))
}
