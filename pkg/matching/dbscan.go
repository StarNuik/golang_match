package matching

import (
	"math"

	"github.com/kelindar/dbscan"
	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

func NewDbscan(cfg KernelConfig) Kernel {
	return &dbscanKernel{
		cfg: cfg,
	}
}

type dbscanKernel struct {
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

func (k *dbscanKernel) Match(queue []*model.QueuedUser) []schema.MatchResponse {
	points := make([]dbscan.Point, len(queue))
	for idx := range queue {
		points[idx] = &userPoint{queue[idx]}
	}

	matches := []schema.MatchResponse{}
	for {
		passPoints, passMatches := k.dbscanPass(points)
		if len(passMatches) == 0 {
			break
		}
		matches = append(matches, passMatches...)
		points = passPoints
	}

	return matches
}

func (k *dbscanKernel) dbscanPass(points []dbscan.Point) ([]dbscan.Point, []schema.MatchResponse) {
	matches := []schema.MatchResponse{}
	toPrune := make(map[string]struct{})

	// todo: better epsilon
	epsilon := 100.0
	clusters := dbscan.Cluster(k.cfg.MatchSize, epsilon, points...)

	if len(clusters) == 0 {
		return nil, nil
	}

	for _, cluster := range clusters {
		match := []*model.QueuedUser{}
		for idx := range k.cfg.MatchSize {
			user := cluster[idx].(*userPoint).user
			match = append(match, user)
			toPrune[user.Name] = struct{}{}
		}

		matches = append(matches, fillResponse(match))
	}

	prunedPoints := []dbscan.Point{}
	for _, p := range points {
		user := p.(*userPoint).user
		if _, ok := toPrune[user.Name]; ok {
			continue
		}
		prunedPoints = append(prunedPoints, p)
	}

	return prunedPoints, matches
}
