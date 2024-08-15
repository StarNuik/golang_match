package matching

import (
	"math"

	"github.com/kelindar/dbscan"
	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

func NewDbscan(cfg KernelConfig) Kernel {
	return &DbscanKernel{
		cfg: cfg,
	}
}

type DbscanKernel struct {
	cfg KernelConfig
}

type userPoint struct {
	user *model.QueuedUser
}

func (p *userPoint) DistanceTo(otherDbscan dbscan.Point) float64 {
	other := otherDbscan.(*userPoint)
	skillDelta := p.user.Skill - other.user.Skill
	latencyDelta := p.user.Latency - other.user.Latency
	return math.Sqrt(skillDelta*skillDelta + latencyDelta*latencyDelta)
}

func (p *userPoint) Name() string {
	return p.user.Name
}

func (k *DbscanKernel) Match(queue []*model.QueuedUser) []schema.MatchResponse {
	points := make(map[string]dbscan.Point, len(queue))
	for _, user := range queue {
		point := userPoint{user: user}
		points[user.Name] = &point
	}

	matches := []schema.MatchResponse{}
	backing := make([]dbscan.Point, len(queue))
	for {
		passMatches := k.dbscanPass(points, backing)
		if len(passMatches) == 0 {
			break
		}
		matches = append(matches, passMatches...)
	}

	return matches
}

func (k *DbscanKernel) dbscanPass(points map[string]dbscan.Point, backing []dbscan.Point) []schema.MatchResponse {
	matches := []schema.MatchResponse{}

	idx := 0
	for _, p := range points {
		backing[idx] = p
		idx++
	}
	ptSlice := backing[:len(points)]

	// todo: better epsilon
	epsilon := 100.0
	clusters := dbscan.Cluster(k.cfg.MatchSize, epsilon, ptSlice...)

	if len(clusters) == 0 {
		return nil
	}

	for _, cluster := range clusters {
		matches = append(matches, clusterToRespones(points, cluster, k.cfg.MatchSize)...)
	}

	return matches
}

// this is a de-duping function, because dbscan.Cluster returns arrays that MAY repeat any point
func clusterToRespones(points map[string]dbscan.Point, cluster []dbscan.Point, blockSize int) []schema.MatchResponse {
	out := []schema.MatchResponse{}

	match := make([]*model.QueuedUser, 0, 8)
	for _, p := range cluster {
		point := p.(*userPoint)
		if _, ok := points[point.user.Name]; ok {
			match = append(match, point.user)
			delete(points, point.user.Name)
		}
		if len(match) == blockSize {
			out = append(out, fillResponse(match))
			match = match[:0]
		}
	}
	return out
}
