package models

import "sort"

type hasPriceRange interface {
	hasPrices
	MinPeriods() []PricePeriod
	MaxPeriods() []PricePeriod
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

	// We want to implement this as a map so we don't double-add price periods.
	// We're going to use it as a set
	minPeriodsSet map[PricePeriod]interface{}
	maxPeriodsSet map[PricePeriod]interface{}

	minPeriodsCached[]PricePeriod
	maxPeriodsCached[]PricePeriod
}

func (prices *PriceRange) createPeriodCache(
	periodSet map[PricePeriod]interface{},
) []PricePeriod {
	newCache := make([]PricePeriod, len(periodSet))
	i := 0
	for key, _ := range periodSet {
		newCache[i] = key
		i++
	}

	// Sort the slice. Maps are not sorted so we have to sort them here.
	sort.SliceStable(
		newCache,
		func(i, j int) bool {
			return newCache[i] < newCache[j]
		},
	)

	return newCache
}

func (prices *PriceRange) MinPeriods() []PricePeriod {
	if prices.minPeriodsCached == nil {
		prices.minPeriodsCached = prices.createPeriodCache(prices.minPeriodsSet)
	}
	return prices.minPeriodsCached
}

func (prices *PriceRange) MaxPeriods() []PricePeriod {
	if prices.maxPeriodsCached == nil {
		prices.maxPeriodsCached = prices.createPeriodCache(prices.maxPeriodsSet)
	}
	return prices.maxPeriodsCached
}

func (prices *PriceRange) clearPeriods(minUpdated bool, maxUpdated bool) {
	// If the value was updated, we have a new min/max, so we need to clear the
	// map
	if minUpdated {
		prices.minPeriodsSet = make(map[PricePeriod]interface{})
		prices.minPeriodsCached = nil
	}
	if maxUpdated {
		prices.maxPeriodsSet = make(map[PricePeriod]interface{})
		prices.maxPeriodsCached = nil
	}
}

// Update from another analysis object
func (prices *PriceRange) updatePriceRangeFromPrices(
	other hasPrices, period PricePeriod,
) {
	minUpdated := prices.updateMin(other.MinPrice(), true)
	maxUpdated := prices.updateMax(other.MaxPrice())
	prices.clearPeriods(minUpdated, maxUpdated)

	// Now add the price period to the set if it was updated OR if it's equal to our
	// current value, as that means it's another high or low point
	if minUpdated || other.MinPrice() == prices.minPrice {
		prices.minPeriodsSet[period] = nil
	}

	if maxUpdated || other.MaxPrice() == prices.maxPrice {
		prices.maxPeriodsSet[period] = nil
	}
}

func (prices *PriceRange) addPeriodsToSet(
	periods []PricePeriod, set map[PricePeriod]interface{},
) {
	for _, pricePeriod := range periods {
		set[pricePeriod] = nil
	}
}

// Update from another analysis object
func (prices *PriceRange) updatePriceRangeFromOther(other hasPriceRange) {
	minUpdated, maxUpdated := prices.updatePrices(other, false)
	prices.clearPeriods(minUpdated, maxUpdated)

	// Now add the price period to the set if it was updated OR if it's equal to our
	// current value, as that means it's another high or low point
	if minUpdated || other.MinPrice() == prices.minPrice {
		prices.addPeriodsToSet(other.MinPeriods(), prices.minPeriodsSet)
	}

	if maxUpdated || other.MaxPrice() == prices.maxPrice {
		prices.addPeriodsToSet(other.MaxPeriods(), prices.maxPeriodsSet)
	}
}

type Analysis struct {
	PriceRange
	chance float64
}

func (analysis *Analysis) Chance() float64 {
	return analysis.chance
}

func (analysis *Analysis) setChance(value float64) {
	analysis.chance = value
}
