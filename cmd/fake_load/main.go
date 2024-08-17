package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	density     = 10
	endpointUrl = "http://service-match:8080/api/users"
)

func postRequest() {
	req := randomUser()
	packed, err := json.Marshal(&req)
	if err != nil {
		log.Println(err)
		return
	}
	body := bytes.NewBuffer(packed)

	resp, err := http.Post(endpointUrl, "application/json", body)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		log.Println(fmt.Errorf("%s", resp.Status))
		return
	}
}

type QueueUserRequest struct {
	Name    string
	Skill   float64
	Latency float64
}

func randomUser() QueueUserRequest {
	nameBytes := make([]byte, 16)
	rand.Read(nameBytes)
	name := hex.EncodeToString(nameBytes)

	out := QueueUserRequest{
		Name:    name,
		Skill:   rand.NormFloat64()*800.0 + 2500.0, // mean(2500), sd(800)
		Latency: rand.NormFloat64()*250.0 + 500.0,  // mean(500), sd(250)
	}
	return out
}

func setup() {
	density = atoiEnv("USERS_PER_SECOND")
	endpointUrl = os.Getenv("ENDPOINT_URL")
}

func atoiEnv(key string) int {
	out, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		log.Panicf("%s: %v\n", key, err)
	}
	return out
}

func main() {
	setup()

	for {
		delay := time.Second / time.Duration(density)

		for range density {
			go postRequest()
			time.Sleep(delay)
		}
	}
}
