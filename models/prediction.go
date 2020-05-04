package models

import (
	"github.com/peake100/turnup-go/errs"
)

type Prediction struct {
	analysis *Analysis
	Patterns []*PotentialPattern
}

func (prediction *Prediction) Analysis() *Analysis {
	if prediction.analysis == nil {
		prediction.analysis = new(Analysis)
	}
	return prediction.analysis
}

// Returns the potential pattern predictions for a given pattern. Returns nil if
// ``pattern`` is not a valid pattern.
func (prediction *Prediction) Pattern(pattern PricePattern) (*PotentialPattern, error) {
	for _, potentialPattern := range prediction.Patterns {
		if potentialPattern.Pattern == pattern {
			return potentialPattern, nil
		}
	}

	return nil, errs.ErrPatternStringValue
}
