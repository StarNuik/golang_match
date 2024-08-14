package model

import (
	"context"
	"time"
)

type QueuedUser struct {
	Name     string
	Skill    float64
	Latency  float64
	QueuedAt time.Time
}

type UserQueue interface {
	Add(context.Context, QueuedUser) error
	GetAll(context.Context) ([]*QueuedUser, error)
	Remove(context.Context, []string) error
}
