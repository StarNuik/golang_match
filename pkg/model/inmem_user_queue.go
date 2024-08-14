package model

import (
	"context"
	"fmt"
)

func NewUserQueueInmemory() UserQueue {
	return &inmemoryUserQueue{
		users: make(map[string]*QueuedUser),
	}
}

type inmemoryUserQueue struct {
	users map[string]*QueuedUser
}

func (m *inmemoryUserQueue) Add(_ context.Context, user QueuedUser) error {
	if _, ok := m.users[user.Name]; ok {
		return fmt.Errorf("user_queue: user already exists")
	}
	m.users[user.Name] = &user
	return nil
}

func (m *inmemoryUserQueue) GetAll(context.Context) ([]*QueuedUser, error) {
	slice := make([]*QueuedUser, len(m.users))
	i := 0
	for _, user := range m.users {
		slice[i] = user
		i++
	}
	return slice, nil
}

func (m *inmemoryUserQueue) Remove(_ context.Context, keys []string) error {
	for _, key := range keys {
		delete(m.users, key)
	}
	return nil
}
