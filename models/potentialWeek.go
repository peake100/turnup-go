package models

type PotentialWeek struct {
	analysis     *Analysis
	PricePeriods []*PotentialPricePeriod
}

func (potential *PotentialWeek) Analysis() *Analysis {
	if potential.analysis == nil {
		potential.analysis = new(Analysis)
	}
	return potential.analysis
}
