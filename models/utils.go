package models

import (
	"math"
)

func RoundBells(bells float32) int {
	return int(math.Ceil(float64(bells)))
}
