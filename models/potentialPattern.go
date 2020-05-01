package models

type PotentialPattern struct {
	// The pattern
	Pattern Pattern
	// The chance, min price and max price
	analysis *Analysis
	// The potential week's price patterns
	PotentialWeeks []*PotentialWeek
}

func (potential *PotentialPattern) Analysis() *Analysis {
	if potential.analysis == nil {
		potential.analysis = new(Analysis)
	}
	return potential.analysis
}
