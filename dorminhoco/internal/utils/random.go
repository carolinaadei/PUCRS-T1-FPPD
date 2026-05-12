package utils

import (
	"math/rand"
	"time"
)

func RandomDelay(maxMs int) {
	if maxMs <= 0 {
		return
	}

	time.Sleep(
		time.Duration(rand.Intn(maxMs)) * time.Millisecond,
	)
}
