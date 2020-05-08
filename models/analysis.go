package models

import (
	"sort"
)

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

// Information about the min and max prices over the 12 price periods of the week.
type PriceSeries struct {
	prices

	// We want to implement this as a map so we don't double-add price periods.
	// We're going to use it as a set
	minPeriodsSet map[PricePeriod]interface{}
	maxPeriodsSet map[PricePeriod]interface{}

	minPeriodsCached []PricePeriod
	maxPeriodsCached []PricePeriod
}

func (prices *PriceSeries) createPeriodCache(
	periodSet map[PricePeriod]interface{},
) []PricePeriod {
	newCache := make([]PricePeriod, len(periodSet))
	i := 0
	for key := range periodSet {
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

// The PricePeriods that this minimum guaranteed price might occur. On PotentialWeeks,
// this will always be a single value, but on PotentialPatterns and Predictions, every
// possible day the minimum guaranteed price *might* occur is used.
func (prices *PriceSeries) MinPeriods() []PricePeriod {
	if prices.minPeriodsCached == nil {
		prices.minPeriodsCached = prices.createPeriodCache(prices.minPeriodsSet)
	}
	return prices.minPeriodsCached
}

// The PricePeriods that this maximum potential price might occur. On PotentialWeeks,
// this will always be a single value, but on PotentialPatterns and Predictions, every
// possible day the maximum potential price *might* occur is used.
func (prices *PriceSeries) MaxPeriods() []PricePeriod {
	if prices.maxPeriodsCached == nil {
		prices.maxPeriodsCached = prices.createPeriodCache(prices.maxPeriodsSet)
	}
	return prices.maxPeriodsCached
}

func (prices *PriceSeries) clearPeriods(minUpdated bool, maxUpdated bool) {
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
func (prices *PriceSeries) updatePriceRangeFromPrices(
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

func (prices *PriceSeries) addPeriodsToSet(
	periods []PricePeriod, set map[PricePeriod]interface{},
) {
	for _, pricePeriod := range periods {
		set[pricePeriod] = nil
	}
}

// Update from another analysis object
func (prices *PriceSeries) updatePriceRangeFromOther(other hasPriceRange) {
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

// Price range and the chance of the range occurring. This type is designed to be
// embedded into prediction, pattern, and week objects to give them a common interface
// for fetching price and probability information.
type Analysis struct {
	PriceSeries
	chance float64
}

// The chance from 0.0-1.0 that this week / pattern / price will occur.
func (analysis *Analysis) Chance() float64 {
	return analysis.chance
}

func (analysis *Analysis) setChance(value float64) {
	// In some instances weird float rounding errors result in a -0 value. We're going
	// to flip the signs on this.
	if value == -0 {
		value = 0
	}
	analysis.chance = value
}
