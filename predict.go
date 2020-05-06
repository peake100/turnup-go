package turnup

import (
	"github.com/peake100/turnup-go/models"
	"github.com/peake100/turnup-go/predictor"
)

// Make an alias to the ticker model here. The high level API is just the ticker and
// Predict function
var NewPriceTicker = models.NewTicker

type Prediction = models.Prediction

// Predict the possible price patterns given the current week's turnip prices on an
// island.
func Predict(currentWeek *models.PriceTicker) (*Prediction, error) {
	thisPredictor := &predictor.Predictor{
		Ticker: currentWeek,
	}
	return thisPredictor.Predict()
}
