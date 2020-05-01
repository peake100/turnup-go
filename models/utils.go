package models

import "math"

func RoundBells(bells float64) int {
	return int(math.Ceil(bells))
}
