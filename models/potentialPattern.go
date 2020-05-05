package models

// Describes the potential prices and chance of a given price pattern.
type PotentialPattern struct {
	SpikeRange
	// The pattern
	Pattern PricePattern
	// The chance, min price and max price
	analysis *Analysis
	// The potential week's price patterns
	PotentialWeeks []*PotentialWeek
}

// The chance, min price and max price
func (potential *PotentialPattern) Analysis() *Analysis {
	if potential.analysis == nil {
		potential.analysis = new(Analysis)
	}
	return potential.analysis
}
