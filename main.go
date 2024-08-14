package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/starnuik/golang_match/pkg/matching"
	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

var (
	userQueue        model.UserQueue
	kernel           matching.Kernel
	matchingTickRate time.Duration
	matchSize        int
)

func errStatus(ctx *gin.Context, status int, err error) {
	log.Println(err)
	ctx.Status(status)
}

func queueUser(ctx *gin.Context) {
	var req schema.QueueUserRequest

	err := ctx.BindJSON(&req)
	if err != nil {
		errStatus(ctx, http.StatusBadRequest, err)
		return
	}

	err = userQueue.Add(context.TODO(), model.QueuedUser{
		Name:     req.Name,
		Skill:    req.Skill,
		Latency:  req.Latency,
		QueuedAt: time.Now().UTC(),
	})
	if err != nil {
		errStatus(ctx, http.StatusInternalServerError, err)
		return
	}
}

func matchUsers() {
	queue, err := userQueue.GetAll(context.TODO())
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%d users queued\n", len(queue))
	matches := kernel.Match(queue)
	if len(matches) == 0 {
		return
	}

	fmt.Printf("matched %d teams\n", len(matches))
	for _, resp := range matches {
		finalizeTeam(resp)
	}
}

func finalizeTeam(resp schema.MatchResponse) {
	err := userQueue.Remove(context.TODO(), resp.Names)
	if err != nil {
		log.Println(err)
		return
	}

	packed, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(string(packed))
}

func matchUsersLoop() {
	for {
		time.Sleep(matchingTickRate)
		matchUsers()
	}
}

func setupEnv() {
	tickSeconds, err := strconv.Atoi(os.Getenv("TICK_SECONDS"))
	if err != nil {
		log.Panicln(err)
	}

	matchingTickRate = time.Duration(tickSeconds) * time.Second

	matchSize, err = strconv.Atoi(os.Getenv("GROUP_SIZE"))
	if err != nil {
		log.Panicln(err)
	}
}

func main() {
	setupEnv()

	userQueue = model.NewUserQueue()
	kernel = matching.NewFifo(matching.KernelConfig{
		MatchSize: matchSize,
	})

	r := gin.Default()

	r.POST("/api/users", queueUser)

	go matchUsersLoop()

	r.Run()
}
