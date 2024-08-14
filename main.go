package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/starnuik/golang_match/pkg/model"
	"github.com/starnuik/golang_match/pkg/schema"
)

var (
	userQueue model.UserQueue
)

func queueUser(ctx *gin.Context) {
	var req schema.QueueUserRequest

	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	userQueue.AddUser(context.TODO(), model.QueuedUser{
		Name:     req.Name,
		Skill:    req.Skill,
		Latency:  req.Latency,
		QueuedAt: time.Now().UTC(),
	})
}

func matchUsers() {
	fmt.Printf("%d users queued\n", len(userQueue.GetUsers(context.TODO())))
}

func matchUsersLoop() {
	for {
		time.Sleep(time.Second)
		matchUsers()
	}
}

func main() {
	userQueue = model.NewUserQueue()

	r := gin.Default()

	r.POST("/api/users", queueUser)

	go matchUsersLoop()

	r.Run()
}
