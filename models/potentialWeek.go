package models

type PotentialWeek struct {
	*Analysis
	Spikes       *SpikeRange
	PricePeriods []*PotentialPricePeriod
}
