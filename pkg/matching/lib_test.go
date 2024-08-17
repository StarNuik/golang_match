package matching_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/starnuik/golang_match/pkg/matching"
	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

func rangeOver(matchSizes []int, userSizes []int, call func(datasetSize int, matchSize int)) {
	for _, uSize := range userSizes {
		for _, mSize := range matchSizes {
			call(uSize, mSize)
		}
	}
}

type factoryModel func(cfg model.GridConfig) model.UserQueue
type factoryKernel func(matching.KernelConfig) matching.Kernel

func rangeKernels(dbUrl string, run func(string, factoryModel, factoryKernel)) {
	kt := []struct {
		factoryKernel factoryKernel
		label         string
	}{
		{
			factoryKernel: func(cfg matching.KernelConfig) matching.Kernel {
				return matching.NewBasicKernel(cfg)
			},
			label: "basic",
		},
		{
			factoryKernel: func(cfg matching.KernelConfig) matching.Kernel {
				return matching.NewPriorityKernel(cfg)
			},
			label: "priority",
		},
	}
	mt := []struct {
		factoryUsers factoryModel
		label        string
	}{
		{
			factoryUsers: func(cfg model.GridConfig) model.UserQueue {
				return model.NewUserQueueInmemory(cfg)
			},
			label: "inmem",
		},
		{
			factoryUsers: func(cfg model.GridConfig) model.UserQueue {
				db, _ := pgxpool.New(context.Background(), dbUrl)
				db.Exec(context.Background(), `delete from UserQueue`)
				// can't `defer db.Close()`
				return model.NewUserQueuePostgres(cfg, db)
			},
			label: "postgres",
		},
	}
	for _, k := range kt {
		for _, m := range mt {
			run(fmt.Sprintf("%s_%s", k.label, m.label), m.factoryUsers, k.factoryKernel)
		}
	}
}

// randomized users for testing (using a normal distribution)
type dataset struct {
	model model.UserQueue
	dict  map[string]*model.QueuedUser
}

func newDataset(factory factoryModel, cfg model.GridConfig, size int) *dataset {
	out := &dataset{
		model: factory(cfg),
		dict:  make(map[string]*model.QueuedUser),
	}

	out.add(size)

	return out
}

func (d *dataset) add(size int) {
	for range size {
		user := randomUser()
		if _, exists := d.dict[user.Name]; exists {
			continue
		}
		d.dict[user.Name] = user
		d.model.Add(context.Background(), user)
	}
}

func (d *dataset) moveQueuedAt(by time.Duration) {
	for _, user := range d.dict {
		user.QueuedAt = user.QueuedAt.Add(-by)
	}
}

func (d *dataset) remove(matches []schema.MatchResponse) {
	names := []string{}
	for _, match := range matches {
		for _, name := range match.Names {
			delete(d.dict, name)
			names = append(names, name)
		}
	}
	// just in case, but the kernel should remove them anyway
	d.model.Remove(context.Background(), names)
}

func randomUser() *model.QueuedUser {
	nameBytes := make([]byte, 16)
	rand.Read(nameBytes)
	name := hex.EncodeToString(nameBytes)

	out := model.QueuedUser{
		Name:     name,
		Skill:    rand.NormFloat64()*800.0 + 2500.0, // mean(2500), sd(800)
		Latency:  rand.NormFloat64()*250.0 + 500.0,  // mean(500), sd(250)
		QueuedAt: time.Now().UTC(),
	}
	return &out
}

// statistical info for a kernel (hacky & crude)
type onlineVariance struct {
	sumSkillSd   float64
	sumLatencySd float64
	sumWaitMean  float64
	sumWaitSd    float64
	sumMatched   float64
	n            int
	nn           int
}

func (s *onlineVariance) Push(users map[string]*model.QueuedUser, matches []schema.MatchResponse, matchSize int) {
	now := time.Now().UTC()
	for _, match := range matches {
		// https://stackoverflow.com/a/1175084
		var skillMean, latencyMean, waitMean, skillSd, latencySd, waitSd float64
		for _, username := range match.Names {
			user := users[username]
			skillMean += user.Skill
			skillSd += user.Skill * user.Skill
			latencyMean += user.Latency
			latencySd += user.Latency * user.Latency
			wait := float64(now.Sub(user.QueuedAt).Milliseconds())
			waitMean += wait
			waitSd += wait * wait
		}
		mSize := float64(matchSize)
		skillMean = skillMean / mSize
		skillSd = math.Sqrt(skillSd/mSize - skillMean*skillMean)
		latencyMean = latencyMean / mSize
		latencySd = math.Sqrt(latencySd/mSize - latencyMean*latencyMean)
		waitMean = waitMean / mSize
		waitSd = math.Sqrt(waitSd/mSize - waitMean*waitMean)

		// https://www.statology.org/averaging-standard-deviations/
		s.sumSkillSd += skillSd * skillSd
		s.sumLatencySd += latencySd * latencySd
		s.sumWaitMean += waitMean
		s.sumWaitSd += waitSd * waitSd
	}
	s.n += len(matches)
	s.sumMatched += float64(len(matches)*matchSize) / float64(len(users))
	s.nn++
}

func (s *onlineVariance) Stat(printMatchrate bool, printWaittime bool) string {
	skillSd := math.Sqrt(s.sumSkillSd / float64(s.n))
	latencySd := math.Sqrt(s.sumLatencySd / float64(s.n))
	waitMean := s.sumWaitMean / float64(s.n)
	waitSd := math.Sqrt(s.sumWaitSd / float64(s.n))
	matchRate := s.sumMatched / float64(s.nn)

	str := fmt.Sprintf("SkillSd(%5.1f), LatencySd(%5.1f)", skillSd, latencySd)
	if printMatchrate {
		mr := fmt.Sprintf("MatchRate(%5.2f)", matchRate)
		str = strings.Join([]string{str, mr}, ", ")
	}
	if printWaittime {
		wt := fmt.Sprintf("WaitTime(%5.1f), WaitTimeSd(%5.1f)", waitMean/1000, waitSd/1000)
		str = strings.Join([]string{str, wt}, ", ")
	}
	return str
}
