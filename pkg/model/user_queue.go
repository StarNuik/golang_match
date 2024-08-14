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
	AddUser(context.Context, QueuedUser)
	GetUsers(context.Context) []QueuedUser
}

func NewUserQueue() UserQueue {
	return &inmemoryUserQueue{}
}

type inmemoryUserQueue struct {
	// todo: better storage primitive?
	users []QueuedUser
}

func (m *inmemoryUserQueue) AddUser(_ context.Context, user QueuedUser) {
	m.users = append(m.users, user)
}

func (m *inmemoryUserQueue) GetUsers(context.Context) []QueuedUser {
	return m.users
}
