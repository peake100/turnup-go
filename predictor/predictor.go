package predictor

import (
	"github.com/peake100/turnup-go/errs"
	"github.com/peake100/turnup-go/models"
	"github.com/peake100/turnup-go/models/patterns"
)

type Predictor struct {
	// The price ticker to use for this prediction
	Ticker *models.PriceTicker

	// The prediction result
	result *models.Prediction

	// The total probability width
	totalWidth float64
}

func (predictor *Predictor) increaseBinWidth(amount float64) {
	predictor.totalWidth += amount
}

func (predictor *Predictor) Predict() (*models.Prediction, error) {
	result := new(models.Prediction)
	predictor.result = result

	currentWeek := predictor.Ticker

	validPrices := false
	for _, pattern := range patterns.PATTERNSGAME {
		patternPredictor := &patternPredictor{
			Ticker:  predictor.Ticker,
			Pattern: pattern,
		}

		potentialPattern, binWidth := patternPredictor.Predict()

		if len(potentialPattern.PotentialWeeks) > 0 {
			validPrices = true
		}

		result.Patterns = append(result.Patterns, potentialPattern)
		result.Analysis().Update(potentialPattern.Analysis(), false)
		result.UpdateSpikeFromRange(potentialPattern)
		predictor.increaseBinWidth(binWidth)
	}

	// If there are no possible price patterns based on this ticker, return an error
	if !validPrices {
		return nil, errs.ErrImpossibleTickerPrices
	}
	predictor.calculateChances(currentWeek, result)

	return result, nil
}
