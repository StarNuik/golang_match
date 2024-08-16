package model

import (
	"context"
	"fmt"
	"time"

	"github.com/starnuik/golang_match/pkg/schema"
)

type QueuedUser struct {
	Name     string
	Skill    float64
	Latency  float64
	QueuedAt time.Time
}

type BinIdx struct {
	S int // Skill
	L int // Latency
}

type GridConfig struct {
	SkillCeil   float64
	LatencyCeil float64
	Side        int
}

type UserQueue interface {
	Parse(*schema.QueueUserRequest) (*QueuedUser, error)
	Add(context.Context, *QueuedUser) error
	GetBin(context.Context, BinIdx) ([]*QueuedUser, error)
	GetBins(ctx context.Context, lo BinIdx, hi BinIdx, minWait time.Duration) ([]*QueuedUser, error)
	Remove(context.Context, []string) error
	Count(context.Context) (int, error)
}

func parse(req *schema.QueueUserRequest) (*QueuedUser, error) {
	if req.Skill <= 0 {
		return nil, fmt.Errorf("skill <= 0")
	}
	if req.Latency <= 0 {
		return nil, fmt.Errorf("latency <= 0")
	}
	if len(req.Name) == 0 {
		return nil, fmt.Errorf("empty name")
	}
	return &QueuedUser{
		Name:     req.Name,
		Skill:    req.Skill,
		Latency:  req.Latency,
		QueuedAt: time.Now().UTC(),
	}, nil
}
