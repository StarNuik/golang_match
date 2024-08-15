package schema

type QueueUserRequest struct {
	Name    string
	Skill   float64
	Latency float64
}

type Candle struct {
	Min       float64
	Average   float64
	Max       float64
	Deviation float64
}

type MatchResponse struct {
	Serial      int
	Skill       Candle
	Latency     Candle
	WaitSeconds Candle
	Names       []string
}
