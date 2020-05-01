package models

import (
	"golang.org/x/xerrors"
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
func (prediction *Prediction) Pattern(pattern Pattern) (*PotentialPattern, error) {
	for _, potentialPattern := range prediction.Patterns {
		if potentialPattern.Pattern == pattern {
			return potentialPattern, nil
		}
	}

	return nil, xerrors.Errorf("'%v' is not a valid pattern", pattern)
}
