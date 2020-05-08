package models

import (
	"github.com/peake100/turnup-go/errs"
)

type Predictor struct {
	// The price ticker to use for this prediction
	Ticker *PriceTicker

	// The prediction result
	result *Prediction

	// The total probability width
	totalWidth float64
}

func (predictor *Predictor) increaseBinWidth(amount float64) {
	predictor.totalWidth += amount
}

func (predictor *Predictor) Predict() (*Prediction, error) {
	result := &Prediction{
		Spikes:      &SpikeChancesAll{
			small: new(SpikeChance),
			big:   new(SpikeChance),
			any:   new(SpikeChance),
		},
		Patterns:    nil,
	}
	predictor.result = result

	currentWeek := predictor.Ticker

	validPrices := false
	for _, pattern := range PATTERNSGAME {
		patternPredictor := &patternPredictor{
			Ticker:  predictor.Ticker,
			Pattern: pattern,
		}

		potentialPattern, binWidth := patternPredictor.Predict()

		if len(potentialPattern.PotentialWeeks) > 0 {
			validPrices = true
		}

		// Integrate this data with our top-level summary
		result.Patterns = append(result.Patterns, potentialPattern)
		result.updatePriceRangeFromOther(potentialPattern)
		result.Spikes.updateRanges(potentialPattern.Spikes)
		predictor.increaseBinWidth(binWidth)
	}

	// If there are no possible price patterns based on this ticker, return an error
	if !validPrices {
		return nil, errs.ErrImpossibleTickerPrices
	}
	predictor.calculateChances(currentWeek, result)

	return result, nil
}
