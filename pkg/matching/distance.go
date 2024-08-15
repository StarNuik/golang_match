package matching

import (
	"math"

	"github.com/starnuik/golang_match/pkg/model"
)

//	type float2 struct {
//		x float64
//		y float64
//	}
type data2i struct {
	s int
	l int
}

type data2 struct {
	s float64
	l float64
}
type range2 struct {
	low  float64
	high float64
}

type distance struct {
	skillRange   range2
	latencyRange range2
	sidePixels   int
}

func (d *distance) Pos(user *model.QueuedUser) data2 {
	pos := data2{}
	pos.s = invLerp(user.Skill, d.skillRange)
	pos.l = invLerp(user.Latency, d.latencyRange)
	return pos
}

func (d *distance) Rasterize(vec data2) data2i {
	vec.s *= float64(d.sidePixels)
	vec.l *= float64(d.sidePixels)

	pix := data2i{}
	// -0.001 is there to avoid having 1f-s
	// that would be bad as the idea is to have [0; sidePixels) bins
	pix.s = int(math.Floor(vec.s - 0.001))
	pix.l = int(math.Floor(vec.l - 0.001))
	return pix
}

func (d *distance) Distance(left *model.QueuedUser, right *model.QueuedUser) float64 {
	lhs := d.Pos(left)
	rhs := d.Pos(right)
	delta := data2{
		s: lhs.s - rhs.s,
		l: lhs.l - rhs.l,
	}
	return math.Sqrt(delta.l*delta.l + delta.s*delta.s)
}

func invLerp(value float64, rng range2) float64 {
	weight := (value - rng.low) / (rng.high - rng.low)
	return clamp(weight, rng)
}

func clamp(value float64, rng range2) float64 {
	return min(rng.high, max(rng.low, value))
}
