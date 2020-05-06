package predictor

import (
	"github.com/peake100/turnup-go/models"
	"sync"
)

// Holds the sync objects for the goroutines handling the price phase permutations for
// each pattern.
type patternsPredictionSync struct {
	ResultChan chan *models.PotentialPattern
	WaitGroup  *sync.WaitGroup
}

type weekPredictionSync struct {
	ResultChan chan *models.PotentialWeek
	WaitGroup  *sync.WaitGroup
}
