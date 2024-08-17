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
	"github.com/jackc/pgx/v5/pgxpool"
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
	finalizeTeams(matches)
}

func finalizeTeams(matches []schema.MatchResponse) {
	matchedUsers := []string{}
	for _, match := range matches {
		matchedUsers = append(matchedUsers, match.Names...)
	}

	err := userQueue.Remove(context.TODO(), matchedUsers)
	if err != nil {
		log.Println(err)
		return
	}

	for _, resp := range matches {
		packed, err := json.MarshalIndent(resp, "", " ")
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println(string(packed))
	}
}

func matchUsersLoop() {
	for {
		time.Sleep(matchingTickRate)
		matchUsers()
	}
}

func setupUserQueue(gridSide int) (model.UserQueue, func()) {
	skillCeil := atoiEnv("TUNING_SKILL_CEIL")
	if skillCeil <= 0 {
		log.Panicln("TUNING_SKILL_CEIL must be > 0")
	}
	latencyCeil := atoiEnv("TUNING_LATENCY_CEIL")
	if latencyCeil <= 0 {
		log.Panicln("TUNING_LATENCY_CEIL must be > 0")
	}

	cfg := model.GridConfig{
		SkillCeil:   float64(skillCeil),
		LatencyCeil: float64(latencyCeil),
		Side:        gridSide,
	}

	storageType := os.Getenv("STORAGE_TYPE")
	switch storageType {
	case "inmem":
		return model.NewUserQueueInmemory(cfg), func() {}
	case "postgres":
		dbUrl := os.Getenv("DB_URL")

		db, err := pgxpool.New(context.Background(), dbUrl)
		if err != nil {
			log.Panicln(err)
		}

		err = db.Ping(context.Background())
		if err != nil {
			log.Panicln(err)
		}

		return model.NewUserQueuePostgres(cfg, db), db.Close
	default:
		log.Panicln("STORAGE_TYPE is invalid")
	}
	panic("unreachable")
}

func setupMatching(gridSide int) matching.Kernel {
	matchSize := atoiEnv("MATCH_SIZE")
	if matchSize < 2 {
		log.Panicln("MATCH_SIZE must be >= 2")
	}

	kernelType := os.Getenv("MATCHING_TYPE")

	cfg := matching.KernelConfig{
		MatchSize: matchSize,
		GridSide:  gridSide,
	}

	switch kernelType {
	case "basic":
		return matching.NewBasicKernel(cfg)
	case "priority":
		priorityRadius := atoiEnv("TUNING_PRIORITY_RADIUS")
		if priorityRadius < 1 {
			log.Panicln("TUNING_PRIORITY_RADIUS must be >= 1")
		}

		waitLimitMs := atoiEnv("TUNING_WAIT_SOFT_LIMIT_MS")
		waitLimit := time.Duration(waitLimitMs) * time.Millisecond

		cfg.PriorityRadius = priorityRadius
		cfg.WaitSoftLimit = waitLimit

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
	tickMs := atoiEnv("TICK_MS")
	matchingTickRate = time.Duration(tickMs) * time.Millisecond
	gridSide := atoiEnv("TUNING_GRID_SIDE")
	if gridSide <= 0 {
		log.Panicln("TUNING_GRID_SIDE must be > 0")
	}

	var closeDb func()
	userQueue, closeDb = setupUserQueue(gridSide)
	defer closeDb()
	kernel = setupMatching(gridSide)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.POST("/api/users", queueUser)

	go matchUsersLoop()

	r.Run()
}
