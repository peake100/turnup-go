package models

// Describes the potential prices and chance of a given price pattern.
type PotentialPattern struct {
	// The chance, min price and max price
	*Analysis
	Spikes *SpikeRangeAll
	// The pattern
	Pattern PricePattern
	// The potential week's price patterns
	PotentialWeeks []*PotentialWeek
}
