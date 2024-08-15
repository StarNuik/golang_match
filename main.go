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
	kernel           *matching.Kernel
	matchingTickRate time.Duration
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

	user, err := userQueue.Parse(&req)
	if err != nil {
		errStatus(ctx, http.StatusBadRequest, err)
		return
	}

	err = userQueue.Add(context.TODO(), user)
	if err != nil {
		errStatus(ctx, http.StatusInternalServerError, err)
		return
	}
}

func matchUsers() {
	count, err := userQueue.Count(context.TODO())
	if err != nil {
		log.Println(err)
		return // no db connection anyway
	}

	fmt.Printf("%d users queued\n", count)

	matches, err := kernel.Match(context.TODO(), userQueue)
	if err != nil {
		log.Println(err)
		return
	}
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

func setupEnv() (model.GridConfig, matching.KernelConfig) {
	tickSeconds := atoiEnv("TICK_SECONDS")
	matchingTickRate = time.Duration(tickSeconds) * time.Second

	matchSize := atoiEnv("MATCH_SIZE")
	skillCeil := atoiEnv("TUNING_SKILL_CEIL")
	latencyCeil := atoiEnv("TUNING_LATENCY_CEIL")
	gridSide := atoiEnv("TUNING_GRID_SIDE")
	if gridSide <= 0 {
		log.Panicln("TUNING_GRID_SIDE <= 0")
	}
	return model.GridConfig{
			SkillCeil:   float64(skillCeil),
			LatencyCeil: float64(latencyCeil),
			Side:        gridSide,
		}, matching.KernelConfig{
			MatchSize: matchSize,
			GridSide:  gridSide,
		}
}

func atoiEnv(key string) int {
	out, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		log.Panicln(err)
	}
	return out
}

func main() {
	gridCfg, kernelCfg := setupEnv()

	userQueue = model.NewUserQueueInmemory(gridCfg)
	kernel = matching.NewKernel(kernelCfg)

	r := gin.Default()

	r.POST("/api/users", queueUser)

	go matchUsersLoop()

	r.Run()
}
