package models

import (
	"sort"
)

type hasPriceRange interface {
	HasPrices
	MinPeriods() []PricePeriod
	GuaranteedPeriods() []PricePeriod
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
	pricesVal

	future        bool
	currentPeriod PricePeriod
	currentPrice  int

	// We want to implement this as a map so we don't double-add price periods.
	// We're going to use it as a set
	minPeriodsSet        map[PricePeriod]interface{}
	guaranteedPeriodsSet map[PricePeriod]interface{}
	maxPeriodsSet        map[PricePeriod]interface{}

	minPeriodsCached        []PricePeriod
	guaranteedPeriodsCached []PricePeriod
	maxPeriodsCached        []PricePeriod
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

// The price periods that the absolute minimum price might occur
func (prices *PriceSeries) MinPeriods() []PricePeriod {
	if prices.minPeriodsCached == nil {
		prices.minPeriodsCached = prices.createPeriodCache(prices.minPeriodsSet)
	}
	return prices.minPeriodsCached
}

// The PricePeriods that this minimum guaranteed price might occur. On PotentialWeeks,
// this will always be a single value, but on PotentialPatterns and Predictions, every
// possible day the minimum guaranteed price *might* occur is used.
func (prices *PriceSeries) GuaranteedPeriods() []PricePeriod {
	if prices.guaranteedPeriodsCached == nil {
		prices.guaranteedPeriodsCached = prices.createPeriodCache(
			prices.guaranteedPeriodsSet,
		)
	}
	return prices.guaranteedPeriodsCached
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

func (prices *PriceSeries) clearPeriods(
	guaranteedUpdated bool, maxUpdated bool, minUpdated bool,
) {
	// If the value was updated, we have a new min/max, so we need to clear the
	// map
	if minUpdated {
		prices.minPeriodsSet = make(map[PricePeriod]interface{})
		prices.minPeriodsCached = nil
	}
	if guaranteedUpdated {
		prices.guaranteedPeriodsSet = make(map[PricePeriod]interface{})
		prices.guaranteedPeriodsCached = nil
	}
	if maxUpdated {
		prices.maxPeriodsSet = make(map[PricePeriod]interface{})
		prices.maxPeriodsCached = nil
	}
}

func (prices *PriceSeries) checkFuture(
	otherIn HasPrices, period PricePeriod,
) (skip bool, otherOut HasPrices) {
	// If this is a future-only price range, do not update if we are not on or past
	// the current period.
	if prices.future && !(period >= prices.currentPeriod) {
		return true, nil
	}

	// Next if this IS the same price period as the current price, then the current
	// price needs to be used for the min, max, and guaranteed of this perriod
	if prices.future && period == prices.currentPeriod && prices.currentPrice != 0 {
		otherOut = &pricesVal{
			minPrice:        prices.currentPrice,
			guaranteedPrice: prices.currentPrice,
			maxPrice:        prices.currentPrice,
		}
	} else {
		otherOut = otherIn
	}

	return false, otherOut
}

// Update from another analysis object
func (prices *PriceSeries) updatePriceRangeFromPrices(
	other HasPrices, period PricePeriod,
) {
	skip, other := prices.checkFuture(other, period)
	if skip {
		return
	}

	minUpdated := prices.updateMin(other.MinPrice())
	guaranteedUpdated := prices.updateGuaranteed(
		other.GuaranteedPrice(), true,
	)
	maxUpdated := prices.updateMax(other.MaxPrice())
	prices.clearPeriods(guaranteedUpdated, maxUpdated, minUpdated)

	// Now add the price period to the set if it was updated OR if it's equal to our
	// current value, as that means it's another high or low point
	if minUpdated || other.MinPrice() == prices.minPrice {
		prices.minPeriodsSet[period] = nil
	}

	if guaranteedUpdated || other.GuaranteedPrice() == prices.guaranteedPrice {
		prices.guaranteedPeriodsSet[period] = nil
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
	guaranteedUpdated, maxUpdated, minUpdated := prices.updatePrices(
		other, false,
	)
	prices.clearPeriods(guaranteedUpdated, maxUpdated, minUpdated)

	// Now add the price period to the set if it was updated OR if it's equal to our
	// current value, as that means it's another high or low point
	if minUpdated || other.MinPrice() == prices.minPrice {
		prices.addPeriodsToSet(other.MinPeriods(), prices.minPeriodsSet)
	}

	if guaranteedUpdated || other.GuaranteedPrice() == prices.guaranteedPrice {
		prices.addPeriodsToSet(other.GuaranteedPeriods(), prices.guaranteedPeriodsSet)
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
	// Contains information about a future price series
	Future PriceSeries
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

func NewAnalysis(ticker *PriceTicker) *Analysis {
	return &Analysis{
		PriceSeries: PriceSeries{},
		// We need to set up the future price series for this.
		Future: PriceSeries{
			future:        true,
			currentPeriod: ticker.CurrentPeriod,
			currentPrice:  ticker.Prices[ticker.CurrentPeriod],
		},
	}
}
