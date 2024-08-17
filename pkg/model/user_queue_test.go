package model_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
	"github.com/stretchr/testify/require"
)

var (
	// requires "compose.test.yaml" deployed
	dbUrl = "postgres://pg:insecure@localhost:5432/test"
	ctx   = context.Background()
	cfg   = model.GridConfig{
		SkillCeil:   10,
		LatencyCeil: 10,
		Side:        2,
	}
	wantUsers = []*model.QueuedUser{
		{Skill: 2.5, Latency: 2.5, Name: "user0-bin00", QueuedAt: now().Add(-100 * time.Second)},
		{Skill: 1.25, Latency: 2.5, Name: "user1-bin00", QueuedAt: now().Add(-200 * time.Second)},
		{Skill: 2.5, Latency: 1.25, Name: "user2-bin00", QueuedAt: now().Add(-300 * time.Second)},
		{Skill: 1.25, Latency: 1.25, Name: "user3-bin00", QueuedAt: now().Add(-400 * time.Second)},
		{Skill: 7.5, Latency: 2.5, Name: "user4-bin10", QueuedAt: now().Add(-500 * time.Second)},
		{Skill: 7.5, Latency: 2.5, Name: "user5-bin10", QueuedAt: now().Add(-600 * time.Second)},
		{Skill: 7.5, Latency: 2.5, Name: "user6-bin10", QueuedAt: now().Add(-700 * time.Second)},
		{Skill: 2.5, Latency: 7.5, Name: "user7-bin01", QueuedAt: now().Add(-800 * time.Second)},
		{Skill: 2.5, Latency: 7.5, Name: "user8-bin01", QueuedAt: now().Add(-900 * time.Second)},
		{Skill: 7.5, Latency: 7.5, Name: "user9-bin11", QueuedAt: now().Add(-1000 * time.Second)},
	}
)

// todo: should i look into subtests?
func TestUserQueueAdd(t *testing.T) {
	rangeUserQueue(t, func(t *testing.T, factory factoryUserQueue) {
		require := require.New(t)
		users := factory(cfg)

		err := users.Add(ctx, wantUsers[0])
		require.Nil(err)

		err = users.Add(ctx, wantUsers[0])
		require.Error(err)

		err = users.Add(ctx, wantUsers[1])
		require.Nil(err)
	})
}

func TestUserQueueGet(t *testing.T) {
	rangeUserQueue(t, func(t *testing.T, factory factoryUserQueue) {
		require := require.New(t)
		users := factory(cfg)

		bin, err := users.GetBin(ctx, model.BinIdx{0, 0})
		require.Nil(err)
		require.Len(bin, 0)

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
	})
}
func TestInmemUserQueueCount(t *testing.T) {
	rangeUserQueue(t, func(t *testing.T, factory factoryUserQueue) {
		require := require.New(t)
		users := factory(cfg)

		for idx, user := range wantUsers {
			err := users.Add(ctx, user)
			require.Nil(err)

			count, err := users.Count(ctx)
			require.Nil(err)

			require.Equal(idx+1, count)
		}
	})
}

func TestInmemUserQueueParse(t *testing.T) {
	rangeUserQueue(t, func(t *testing.T, factory factoryUserQueue) {
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
	})
}

func TestInmemUserQueueRemove(t *testing.T) {
	rangeUserQueue(t, func(t *testing.T, factory factoryUserQueue) {
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
		require.Len(bin, 3)
		require.False(binContains(bin, wantUsers[3]))

		bin, err = users.GetBin(ctx, model.BinIdx{1, 0})
		require.Nil(err)
		require.Len(bin, 2)
		require.False(binContains(bin, wantUsers[6]))

		bin, err = users.GetBin(ctx, model.BinIdx{0, 1})
		require.Nil(err)
		require.Len(bin, 1)
		require.False(binContains(bin, wantUsers[8]))

		bin, err = users.GetBin(ctx, model.BinIdx{1, 1})
		require.Nil(err)
		require.Len(bin, 0)
	})
}

func TestInmemUserQueueGetBins(t *testing.T) {
	rangeUserQueue(t, func(t *testing.T, factory factoryUserQueue) {
		require := require.New(t)
		users := factory(cfg)

		for _, user := range wantUsers {
			err := users.Add(ctx, user)
			require.Nil(err)
		}

		bin, err := users.GetBins(ctx, model.BinIdx{0, 0}, model.BinIdx{1, 0}, 350*time.Second)
		require.Nil(err)
		require.Len(bin, 4)
		require.True(binContains(bin, wantUsers[3]))
		require.True(binContains(bin, wantUsers[4]))
		require.True(binContains(bin, wantUsers[5]))
		require.True(binContains(bin, wantUsers[6]))

		bin, err = users.GetBins(ctx, model.BinIdx{0, 0}, model.BinIdx{1, 1}, 350*time.Second)
		require.Nil(err)
		require.Len(bin, 7)
		require.True(binContains(bin, wantUsers[3]))
		require.True(binContains(bin, wantUsers[4]))
		require.True(binContains(bin, wantUsers[5]))
		require.True(binContains(bin, wantUsers[6]))
		require.True(binContains(bin, wantUsers[7]))
		require.True(binContains(bin, wantUsers[8]))
		require.True(binContains(bin, wantUsers[9]))
	})
}

type factoryUserQueue func(cfg model.GridConfig) model.UserQueue

func rangeUserQueue(t *testing.T, run func(*testing.T, factoryUserQueue)) {
	table := []struct {
		label   string
		factory factoryUserQueue
	}{
		{
			"inmem", func(cfg model.GridConfig) model.UserQueue {
				return model.NewUserQueueInmemory(cfg)
			},
		},
		{
			"postgres", func(cfg model.GridConfig) model.UserQueue {
				db, _ := pgxpool.New(context.Background(), dbUrl)
				db.Exec(context.Background(), `delete from UserQueue`)
				// can't `defer db.Close()`
				return model.NewUserQueuePostgres(cfg, db)
			},
		},
	}
	for _, row := range table {
		t.Run(row.label, func(t *testing.T) {
			run(t, row.factory)
		})
	}
}

func binContains(bin []*model.QueuedUser, user *model.QueuedUser) bool {
	return slices.ContainsFunc(bin, func(other *model.QueuedUser) bool {
		return user.Name == other.Name &&
			user.Skill == other.Skill &&
			user.Latency == other.Latency &&
			user.QueuedAt.Sub(other.QueuedAt).Abs() <= time.Second
	})
}

func now() time.Time {
	return time.Now().UTC()
}
