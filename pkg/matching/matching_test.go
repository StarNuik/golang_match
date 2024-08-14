package matching_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/starnuik/golang_match/pkg/model"
)

func rangeOver(call func(userSize int, matchSize int)) {
	matchSizes := []int{4, 8, 16}
	userSizes := []int{100, 1_000 /* 10_000, 100_000 */}

	for _, uSize := range userSizes {
		for _, mSize := range matchSizes {
			call(uSize, mSize)
		}
	}
}

func dataset(size int) []*model.QueuedUser {
	slice := make([]*model.QueuedUser, size)
	for idx := range slice {
		slice[idx] = randomUser()
	}
	return slice
}

func randomUser() *model.QueuedUser {
	nameBytes := make([]byte, 16)
	rand.Read(nameBytes)
	name := hex.EncodeToString(nameBytes)

	out := model.QueuedUser{
		Name:     name,
		Skill:    rand.NormFloat64()*500.0 + 2500.0, // mean(2500), sd(500)
		Latency:  rand.NormFloat64()*250.0 + 500.0,  // mean(500), sd(250)
		QueuedAt: time.Now().UTC(),
	}
	return &out
}

func label(kernelName string, mSize int, uSize int) string {
	return fmt.Sprintf("%s_matchSize(%3d)_userbase(%6d)", kernelName, mSize, uSize)
}
