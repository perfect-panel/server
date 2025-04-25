package random

import (
	"math/rand"
	"time"
)

func RandomInRange(min, max int) int {
	if min > max {
		min, max = max, min
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min+1) + min
}
