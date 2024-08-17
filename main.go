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

	log.Printf("%d users queued\n", count)

	matches, err := kernel.Match(context.TODO(), userQueue)
	if err != nil {
		log.Println(err)
		return
	}
	if len(matches) == 0 {
		return
	}

	log.Printf("matched %d teams\n", len(matches))
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

	packed, err := json.MarshalIndent(resp, "", " ")
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

func setupEnv() model.GridConfig {
	tickMs := atoiEnv("TICK_MS")
	matchingTickRate = time.Duration(tickMs) * time.Millisecond

	skillCeil := atoiEnv("TUNING_SKILL_CEIL")
	if skillCeil <= 0 {
		log.Panicln("TUNING_SKILL_CEIL must be > 0")
	}
	latencyCeil := atoiEnv("TUNING_LATENCY_CEIL")
	if latencyCeil <= 0 {
		log.Panicln("TUNING_LATENCY_CEIL must be > 0")
	}

	gridSide := atoiEnv("TUNING_GRID_SIDE")
	if gridSide <= 0 {
		log.Panicln("TUNING_GRID_SIDE must be > 0")
	}

	return model.GridConfig{
		SkillCeil:   float64(skillCeil),
		LatencyCeil: float64(latencyCeil),
		Side:        gridSide,
	}
}

func setupMatching(gridSide int) matching.Kernel {
	matchSize := atoiEnv("MATCH_SIZE")
	if matchSize < 2 {
		log.Panicln("MATCH_SIZE must be >= 2")
	}

	priorityRadius := atoiEnv("TUNING_PRIORITY_RADIUS")
	if priorityRadius < 1 {
		log.Panicln("TUNING_PRIORITY_RADIUS must be >= 1")
	}

	waitLimitMs := atoiEnv("TUNING_WAIT_SOFT_LIMIT_MS")
	waitLimit := time.Duration(waitLimitMs) * time.Millisecond

	kernelType := os.Getenv("MATCHING_TYPE")

	cfg := matching.KernelConfig{
		MatchSize:      matchSize,
		GridSide:       gridSide,
		PriorityRadius: priorityRadius,
		WaitSoftLimit:  waitLimit,
	}

	switch kernelType {
	case "basic":
		return matching.NewBasicKernel(cfg)
	case "priority":
		return matching.NewBasicKernel(cfg)
	default:
		log.Panicln("MATCHING_TYPE is invalid")
	}
	panic("unreachable")
}

func atoiEnv(key string) int {
	out, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		log.Panicln(err)
	}
	return out
}

func main() {
	gridCfg := setupEnv()

	userQueue = model.NewUserQueueInmemory(gridCfg)
	kernel = setupMatching(gridCfg.Side)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.POST("/api/users", queueUser)

	go matchUsersLoop()

	r.Run()
}
