package model_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
	"github.com/stretchr/testify/require"
)

var (
	cfg = model.GridConfig{
		SkillCeil:   10,
		LatencyCeil: 10,
		Side:        2,
	}
	wantUsers = []*model.QueuedUser{
		{Skill: 2.5, Latency: 2.5, Name: "user0", QueuedAt: now().Add(-100 * time.Second)},
		{Skill: 1.25, Latency: 2.5, Name: "user1", QueuedAt: now().Add(-200 * time.Second)},
		{Skill: 2.5, Latency: 1.25, Name: "user2", QueuedAt: now().Add(-300 * time.Second)},
		{Skill: 1.25, Latency: 1.25, Name: "user3", QueuedAt: now().Add(-400 * time.Second)},
		{Skill: 7.5, Latency: 2.5, Name: "user4", QueuedAt: now().Add(-500 * time.Second)},
		{Skill: 7.5, Latency: 2.5, Name: "user5", QueuedAt: now().Add(-600 * time.Second)},
		{Skill: 7.5, Latency: 2.5, Name: "user6", QueuedAt: now().Add(-700 * time.Second)},
		{Skill: 2.5, Latency: 7.5, Name: "user7", QueuedAt: now().Add(-800 * time.Second)},
		{Skill: 2.5, Latency: 7.5, Name: "user8", QueuedAt: now().Add(-900 * time.Second)},
		{Skill: 7.5, Latency: 7.5, Name: "user9", QueuedAt: now().Add(-1000 * time.Second)},
	}
	ctx          = context.Background()
	inmemFactory = func(cfg model.GridConfig) model.UserQueue {
		return model.NewUserQueueInmemory(cfg)
	}
)

// todo: should i look into subtests?
func TestInmemUserQueueAdd(t *testing.T) {
	testUserQueueAdd(t, inmemFactory)
}

func TestInmemUserQueueGet(t *testing.T) {
	testUserQueueGet(t, inmemFactory)
}

func TestInmemUserQueueParse(t *testing.T) {
	testUserQueueParse(t, inmemFactory)
}

func TestInmemUserQueueCount(t *testing.T) {
	testUserQueueCount(t, inmemFactory)
}

func TestInmemUserQueueRemove(t *testing.T) {
	testUserQueueRemove(t, inmemFactory)
}

func testUserQueueAdd(t *testing.T, factory func(model.GridConfig) model.UserQueue) {
	require := require.New(t)
	users := factory(cfg)

	err := users.Add(ctx, wantUsers[0])
	require.Nil(err)

	err = users.Add(ctx, wantUsers[0])
	require.Error(err)

	err = users.Add(ctx, wantUsers[1])
	require.Nil(err)
}

func testUserQueueGet(t *testing.T, factory func(model.GridConfig) model.UserQueue) {
	require := require.New(t)
	users := factory(cfg)

	bin, err := users.GetBin(ctx, model.BinIdx{0, 0})
	require.Nil(err)
	require.Nil(bin)

	for _, user := range wantUsers {
		err := users.Add(ctx, user)
		require.Nil(err)
	}

	bin, err = users.GetBin(ctx, model.BinIdx{0, 0})
	require.Nil(err)
	require.Equal(4, len(bin))
	require.True(binContains(bin, wantUsers[0]))
	require.True(binContains(bin, wantUsers[1]))
	require.True(binContains(bin, wantUsers[2]))
	require.True(binContains(bin, wantUsers[3]))

	bin, err = users.GetBin(ctx, model.BinIdx{1, 0})
	require.Nil(err)
	require.Equal(3, len(bin))
	require.True(binContains(bin, wantUsers[4]))
	require.True(binContains(bin, wantUsers[5]))
	require.True(binContains(bin, wantUsers[6]))

	bin, err = users.GetBin(ctx, model.BinIdx{0, 1})
	require.Nil(err)
	require.Equal(2, len(bin))
	require.True(binContains(bin, wantUsers[7]))
	require.True(binContains(bin, wantUsers[8]))

	bin, err = users.GetBin(ctx, model.BinIdx{1, 1})
	require.Nil(err)
	require.Equal(1, len(bin))
	require.True(binContains(bin, wantUsers[9]))
}

func testUserQueueCount(t *testing.T, factory func(model.GridConfig) model.UserQueue) {
	require := require.New(t)
	users := factory(cfg)

	for idx, user := range wantUsers {
		err := users.Add(ctx, user)
		require.Nil(err)

		count, err := users.Count(ctx)
		require.Nil(err)

		require.Equal(idx+1, count)
	}
}

func testUserQueueRemove(t *testing.T, factory func(model.GridConfig) model.UserQueue) {
	require := require.New(t)
	users := factory(cfg)

	for _, user := range wantUsers {
		err := users.Add(ctx, user)
		require.Nil(err)
	}

	names := []string{wantUsers[9].Name, wantUsers[8].Name, wantUsers[6].Name, wantUsers[3].Name}
	err := users.Remove(ctx, names)
	require.Nil(err)

	count, err := users.Count(ctx)
	require.Nil(err)
	require.Equal(6, count)

	bin, err := users.GetBin(ctx, model.BinIdx{0, 0})
	require.Nil(err)
	require.Equal(len(bin), 3)
	require.False(binContains(bin, wantUsers[3]))

	bin, err = users.GetBin(ctx, model.BinIdx{1, 0})
	require.Nil(err)
	require.Equal(len(bin), 2)
	require.False(binContains(bin, wantUsers[6]))

	bin, err = users.GetBin(ctx, model.BinIdx{0, 1})
	require.Nil(err)
	require.Equal(len(bin), 1)
	require.False(binContains(bin, wantUsers[8]))

	bin, err = users.GetBin(ctx, model.BinIdx{1, 1})
	require.Nil(err)
	require.Nil(bin)
}

// func testUserQueue(t *testing.T, factory func(model.GridConfig) model.UserQueue) {
// 	require := require.New(t)

// 	cfg := model.GridConfig{
// 		SkillCeil:   10,
// 		LatencyCeil: 10,
// 		Side:        2,
// 	}
// 	users := factory(cfg)

// 	ctx := context.Background()
// 	now := time.Now().UTC()
// 	u1 := model.QueuedUser{
// 		Name:     "user 1",
// 		Skill:    123.4,
// 		Latency:  567.8,
// 		QueuedAt: now.Add(-100 * time.Second),
// 	}
// 	u2 := model.QueuedUser{
// 		Name:     "user 2",
// 		Skill:    9876.5,
// 		Latency:  4321.0,
// 		QueuedAt: now.Add(-200 * time.Hour),
// 	}

// 	// .Add
// 	err := queue.Add(ctx, u1)
// 	require.Nil(err)

// 	err = queue.Add(ctx, u2)
// 	require.Nil(err)

// 	err = queue.Add(ctx, u1)
// 	require.NotNil(err)

// 	err = queue.Add(ctx, u2)
// 	require.NotNil(err)

// 	// .GetAll
// 	containsFunc := func(user model.QueuedUser) func(e *model.QueuedUser) bool {
// 		return func(e *model.QueuedUser) bool {
// 			return *e == user
// 		}
// 	}
// 	haveAll, err := queue.GetAll(ctx)
// 	require.Nil(err)
// 	require.Equal(2, len(haveAll))
// 	require.True(slices.ContainsFunc(haveAll, containsFunc(u1)))
// 	require.True(slices.ContainsFunc(haveAll, containsFunc(u2)))

// 	// .Remove (and re-.Add)
// 	err = queue.Remove(ctx, []string{u2.Name})
// 	require.Nil(err)

// 	haveAll, err = queue.GetAll(ctx)
// 	require.Nil(err)
// 	require.Equal(1, len(haveAll))
// 	require.True(*haveAll[0] == u1)

// 	err = queue.Add(ctx, u2)
// 	require.Nil(err)

// 	err = queue.Remove(ctx, []string{u1.Name, u2.Name})
// 	require.Nil(err)

// 	haveAll, err = queue.GetAll(ctx)
// 	require.Nil(err)
// 	require.Zero(len(haveAll))
// }

func testUserQueueParse(t *testing.T, factory func(model.GridConfig) model.UserQueue) {
	require := require.New(t)
	users := factory(model.GridConfig{})

	want := schema.QueueUserRequest{
		Name:    "bob",
		Skill:   13,
		Latency: 37,
	}
	have, err := users.Parse(&want)
	require.Nil(err)
	require.Equal(want.Name, have.Name)
	require.Equal(want.Skill, have.Skill)
	require.Equal(want.Latency, have.Latency)

	invalid := want
	invalid.Skill = -13
	have, err = users.Parse(&invalid)
	require.Nil(have)
	require.Error(err)

	invalid = want
	invalid.Latency = -37
	have, err = users.Parse(&invalid)
	require.Nil(have)
	require.Error(err)

	invalid = want
	invalid.Name = ""
	have, err = users.Parse(&invalid)
	require.Nil(have)
	require.Error(err)
}

func binContains(bin []*model.QueuedUser, user *model.QueuedUser) bool {
	return slices.ContainsFunc(bin, func(other *model.QueuedUser) bool {
		return *other == *user
	})
}

func now() time.Time {
	return time.Now().UTC()
}
