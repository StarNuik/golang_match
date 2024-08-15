package model

import (
	"context"
	"fmt"
	"math"

	"github.com/starnuik/golang_match/pkg/schema"
)

func NewUserQueueInmemory(cfg GridConfig) UserQueue {
	return &inmemoryUserQueue{
		cfg:  cfg,
		bins: make(map[BinIdx]map[string]*QueuedUser),
	}
}

type inmemoryUserQueue struct {
	cfg  GridConfig
	bins map[BinIdx]map[string]*QueuedUser
}

func (m *inmemoryUserQueue) Parse(req *schema.QueueUserRequest) (*QueuedUser, error) {
	return parse(req)
}

func (m *inmemoryUserQueue) Add(_ context.Context, user *QueuedUser) error {
	idx := m.toIndex(user)

	if _, exists := m.bins[idx]; !exists {
		m.bins[idx] = make(map[string]*QueuedUser)
	}

	bin := m.bins[idx]

	if _, exists := bin[user.Name]; exists {
		return fmt.Errorf("user already exists")
	}

	bin[user.Name] = user
	return nil
}

func (m *inmemoryUserQueue) GetBin(_ context.Context, idx BinIdx) ([]*QueuedUser, error) {
	if _, exists := m.bins[idx]; !exists {
		return nil, nil
	}

	bin := m.bins[idx]

	if len(bin) == 0 {
		return nil, nil
	}

	slice := make([]*QueuedUser, 0, len(bin))
	for _, user := range bin {
		slice = append(slice, user)
	}

	return slice, nil
}

func (m *inmemoryUserQueue) Remove(_ context.Context, keys []string) error {
	// ouch
	for _, bin := range m.bins {
		for _, key := range keys {
			delete(bin, key)
		}
	}
	return nil
}

func (m *inmemoryUserQueue) Count(context.Context) (int, error) {
	var count int
	for _, bin := range m.bins {
		count += len(bin)
	}
	return count, nil
}

func (m *inmemoryUserQueue) toIndex(req *QueuedUser) BinIdx {
	return BinIdx{
		S: remap(req.Skill, m.cfg.SkillCeil, m.cfg.Side),
		L: remap(req.Latency, m.cfg.LatencyCeil, m.cfg.Side),
	}
}

func remap(value float64, ceil float64, sides int) int {
	// inverse lerp
	value = value / ceil
	// lerp
	value = value * float64(sides)
	// floor
	value = math.Floor(value)
	// clamp
	value = max(0, value)
	return min(sides-1, int(value))
}
