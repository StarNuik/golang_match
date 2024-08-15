package matching_test

import (
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/starnuik/golang_match/pkg/model"
)

type dataset struct {
	slice []*model.QueuedUser
	dict  map[string]*model.QueuedUser
}

func newDataset(size int) dataset {
	out := dataset{
		dict:  make(map[string]*model.QueuedUser, size),
		slice: make([]*model.QueuedUser, 0, size),
	}

	for range size {
		newUser := randomUser()
		out.dict[newUser.Name] = newUser
		out.slice = append(out.slice, newUser)
	}

	return out
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
