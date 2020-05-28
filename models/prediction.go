package models

import "github.com/peake100/turnup-go/errs"

// Holds the potential pattern information for a prediction.
type Patterns []*PotentialPattern

// Returns the potential pattern predictions for a given pattern. Returns nil if
// ``pattern`` is not a valid pattern.
func (patterns Patterns) Get(pattern PricePattern) (*PotentialPattern, error) {
	for _, potentialPattern := range patterns {
		if potentialPattern.Pattern == pattern {
			return potentialPattern, nil
		}
	}

	return nil, errs.ErrPatternStringValue
}

type Prediction struct {
	PriceSeries
	Heat     int
	Future   PriceSeries
	Spikes   *SpikeChancesAll
	Patterns Patterns
}
