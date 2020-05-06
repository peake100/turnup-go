package models

type hasPriceRange interface {
	hasPrices
	MinPeriod() PricePeriod
	MaxPeriod() PricePeriod
}

type hasProbability interface {
	Chance() float64
	setChance(value float64)
}

type hasFullAnalysis interface {
	hasPriceRange
	hasProbability
}

type PriceRange struct {
	prices

	minPeriod PricePeriod
	maxPeriod PricePeriod
}

// Update from another analysis object
func (prices *PriceRange) updatePriceRangeFromPrices(
	other hasPrices, period PricePeriod,
) {
	minUpdated := prices.UpdateMin(other.MinPrice(), true)
	maxUpdated := prices.UpdateMax(other.MaxPrice())
	if minUpdated {
		prices.minPeriod = period
	}

	if maxUpdated {
		prices.maxPeriod = period
	}
}

// Update from another analysis object
func (prices *PriceRange) updatePriceRangeFromOther(other hasPriceRange) {
	minUpdated, maxUpdated := prices.updatePrices(other, false)
	if minUpdated {
		prices.minPeriod = other.MinPeriod()
	}

	if maxUpdated {
		prices.maxPeriod = other.MaxPeriod()
	}
}

type Analysis struct {
	PriceRange
	chance float64
}

func (analysis *Analysis) MinPeriod() PricePeriod {
	return analysis.minPeriod
}

func (analysis *Analysis) MaxPeriod() PricePeriod {
	return analysis.maxPeriod
}

func (analysis *Analysis) Chance() float64 {
	return analysis.chance
}

func (analysis *Analysis) setChance(value float64) {
	analysis.chance = value
}
