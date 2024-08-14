package model_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/starnuik/golang_match/pkg/model"
	"github.com/stretchr/testify/require"
)

func TestUserQueueInmem(t *testing.T) {
	queue := model.NewUserQueueInmemory()
	testUserQueue(t, queue)
}

func testUserQueue(t *testing.T, queue model.UserQueue) {
	require := require.New(t)

	ctx := context.Background()
	now := time.Now().UTC()
	u1 := model.QueuedUser{
		Name:     "user 1",
		Skill:    123.4,
		Latency:  567.8,
		QueuedAt: now.Add(-100 * time.Second),
	}
	u2 := model.QueuedUser{
		Name:     "user 2",
		Skill:    9876.5,
		Latency:  4321.0,
		QueuedAt: now.Add(-200 * time.Hour),
	}

	// .Add
	err := queue.Add(ctx, u1)
	require.Nil(err)

	err = queue.Add(ctx, u2)
	require.Nil(err)

	err = queue.Add(ctx, u1)
	require.NotNil(err)

	err = queue.Add(ctx, u2)
	require.NotNil(err)

	// .GetAll
	containsFunc := func(user model.QueuedUser) func(e *model.QueuedUser) bool {
		return func(e *model.QueuedUser) bool {
			return *e == user
		}
	}
	haveAll, err := queue.GetAll(ctx)
	require.Nil(err)
	require.Equal(2, len(haveAll))
	require.True(slices.ContainsFunc(haveAll, containsFunc(u1)))
	require.True(slices.ContainsFunc(haveAll, containsFunc(u2)))

	// .Remove (and re-.Add)
	err = queue.Remove(ctx, []string{u2.Name})
	require.Nil(err)

	haveAll, err = queue.GetAll(ctx)
	require.Nil(err)
	require.Equal(1, len(haveAll))
	require.True(*haveAll[0] == u1)

	err = queue.Add(ctx, u2)
	require.Nil(err)

	err = queue.Remove(ctx, []string{u1.Name, u2.Name})
	require.Nil(err)

	haveAll, err = queue.GetAll(ctx)
	require.Nil(err)
	require.Zero(len(haveAll))
}
