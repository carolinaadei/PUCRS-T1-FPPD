package utils

import (
	"math/rand"
	"time"
)

func RandomDelay(maxMs int) {
	time.Sleep(
		time.Duration(rand.Intn(maxMs)) * time.Millisecond,
	)
}