package models

import (
	"math"
)

func roundBells(bells float32) int {
	return int(math.Ceil(float64(bells)))
}
