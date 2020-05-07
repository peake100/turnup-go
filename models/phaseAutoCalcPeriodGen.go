package models

import (
	"math"
)

// In order to get the sub period price values for a given phase, we have to know
// information from all the previous phases to take in historical data and properly
// updatePrices compounded price multipliers.
//
// If we were to make these calculations on demand, we would end up with an algorithm
// whose execution scaled exponentially as the sub price period increased, as each
// period would need to re-calculate all the information from the period before it.
//
// We COULD solve this by pre-computing all the price periods for a phase in a single
// loop, then returning them when asked. But this would mean that if the first period of
// a phase is a bad match for the ticker, we would have done up to 6 price period
// calculations we would then just throw away
//
// Enter the phasePeriodGenerator. Acts as an iterator-like object that will yield
// the next PotentialPricePeriod when it's .Next() method is called. Internally, it
// keeps all the data we would have in a loop in order to calculate the next price
// period, thus allowing us to NOT re-calculate all past sup-periods when a new one
// is needed, AND not calculate more price periods than strictly necessary for a given
// ticker.
type phasePeriodGenerator struct {
	Ticker           *PriceTicker
	PurchasePrice    int
	PhaseFull        *patternPhaseAuto
	PricePeriodStart PricePeriod

	// For outside consumption
	LastCompletedSubPeriod int

	// We're going to cache a few type conversions
	phase       phaseImplement
	compounding phaseCompoundingPrice
	makesFinal  phaseMakesFinalAdjustment
	hasSpike    phaseHasSpike

	// Data fields - these fields will be updated during processing
	// Whether this iterator has been started
	started bool

	subPeriod   int
	pricePeriod PricePeriod

	baseMultiplierMin float32
	baseMultiplierMax float32

	historicalMultiplierMin float32
	historicalMultiplierMax float32

	finalAdjustment int

	previousPeriodPrice int

	priceMin int
	priceMax int

	binWidthMin float64
	binWidthMax float64
}

func (gen *phasePeriodGenerator) Setup() {
	gen.pricePeriod = gen.PricePeriodStart

	gen.phase = gen.PhaseFull.phaseImplement
	if compounding, ok := gen.phase.(phaseCompoundingPrice); ok {
		gen.compounding = compounding
	}
	if makesFinal, ok := gen.phase.(phaseMakesFinalAdjustment); ok {
		gen.makesFinal = makesFinal
	}
	if hasSpike, ok := gen.phase.(phaseHasSpike); ok {
		gen.hasSpike = hasSpike
	}

	gen.LastCompletedSubPeriod = -1
	gen.started = true
}

func (gen *phasePeriodGenerator) beginThisIteration() {
	// Set the purchase price multiplier for this iteration. If this is a compounding
	// price we only want to do this on the first iteration, as the multipliers
	// will adjust themselves starting in sub-period index 1.
	if gen.compounding == nil || gen.subPeriod == 0 {
		gen.baseMultiplierMin, gen.baseMultiplierMax = gen.phase.BasePriceMultiplier(
			gen.subPeriod,
		)
		// If we're compounding, we want to adjust our multiplier based on the price
		// history, so we're going to to keep track of an adjusted multiplier based on
		// island prices. We ALSO need to track the extreme possible ends of this
		// pattern overall to bound this multiplier when accounting for rounding.
		gen.historicalMultiplierMin = gen.baseMultiplierMin
		gen.historicalMultiplierMax = gen.baseMultiplierMax
	}

	// set the final adjustment that needs to be made this pass
	if gen.makesFinal != nil {
		gen.finalAdjustment = gen.makesFinal.FinalPriceAdjustment(gen.subPeriod)
	}
}

func (gen *phasePeriodGenerator) endThisIteration() {
	// if we are beyond phase one, set the previous price so the compounding
	// calculations have access to it.
	gen.previousPeriodPrice = gen.Ticker.Prices[gen.pricePeriod]
	gen.LastCompletedSubPeriod = gen.subPeriod
	gen.subPeriod++
	gen.pricePeriod++
}

// Calculates the actual multiplier value if we know the price from the previous day.
func (gen *phasePeriodGenerator) calcPhasePeriodHistoricalMultiplier(
	isMin bool,
) (historicMultiplier float32) {
	previousPrice := gen.previousPeriodPrice
	// Un-adjust this price if it has an adjustment (price adjustments
	// never happen in the actual game during compounding phases, but we'll
	// put it here for max compatibility in case that ever changes with an
	// updatePrices, it should always be 0 when this code block is executed).
	previousPrice -= gen.finalAdjustment

	// We need to get the most extreme pre-rounded price that could have
	// resulted in the known price. For the max price, this is the price
	// itself. For isMin, this is the number - 1 + the smallest possible
	// float value.

	previousPriceFloat := float32(previousPrice)
	if isMin {
		previousPriceFloor := previousPriceFloat - 1
		previousPriceFloat = math.Nextafter32(
			previousPriceFloor, previousPriceFloat,
		)
	}

	// now work out the extreme end of the previous multiplier
	historicMultiplier = previousPriceFloat / float32(gen.PurchasePrice)

	var baseMultiplier float32
	if isMin {
		// Because of annoying floating point errors here, we're going to add or
		// subtract a very small bit to the higher or lower bound to give us a little
		// leeway, just a single floating point step is enough
		historicMultiplier = math.Nextafter32(
			historicMultiplier, historicMultiplier - 0.001,
		)
		baseMultiplier = gen.baseMultiplierMin
	} else {
		historicMultiplier = math.Nextafter32(
			historicMultiplier, historicMultiplier + 0.001,
		)
		baseMultiplier = gen.baseMultiplierMax
	}

	// if it is lower than the isMin multiplier or higher than the
	// max multiplier, we need to bring it in line with the
	// possible range
	if (isMin && historicMultiplier < baseMultiplier) ||
		(!isMin && historicMultiplier > baseMultiplier) {
		historicMultiplier = baseMultiplier
	}

	return historicMultiplier
}

func (gen *phasePeriodGenerator) calcNookPriceAndWidth(
	priceMultiplier float32,
) (price int, binWidth float64) {
	price = RoundBells(float32(gen.PurchasePrice) * priceMultiplier)

	// Convert everything to float64 so our prediction math is more precise.
	binWidth = float64(price) -
		(float64(gen.PurchasePrice) * float64(priceMultiplier))

	return price, binWidth
}

func (gen *phasePeriodGenerator) adjustCompounding() {
	// COMPOUNDING FACTORS
	// My first instinct was to calculate this periods rate factor by doing this:
	//		baseMultiplier + (phasePeriod * subPeriodMultiplier)
	//
	// However, if we examine the game logic here:
	//
	// https://gist.github.com/
	// Treeki/85be14d297c80c8b3c0a76375743325b#file-turnipprices-cpp-L320
	//
	// ...we see that the game itself adds the subPeriodMultiplier to the baseMultiplier
	// while looping through each price period.
	//
	// IN A PERFECT MATHEMATICAL WORLD these operations would be equivalent, but in
	// practice we introduce subtle floating point errors that can result in our bell
	// prices being off-by-one from the game. Therefore, we need to exactly imitate the
	// game logic during this calculation.

	// If we know the price for the day before, then we can make a more accurate
	// projection of what this period's prices will be by multiplying the real
	// world price by the upper and lower sub-period bounds. Each time we know
	// the real price for the previous price period, we are going to reset the
	// multiplier to
	if gen.previousPeriodPrice != 0 {
		gen.historicalMultiplierMin = gen.calcPhasePeriodHistoricalMultiplier(
			true,
		)
		gen.historicalMultiplierMax = gen.calcPhasePeriodHistoricalMultiplier(
			false,
		)
	}

	// Alter the base multiplier by the phase period amount.
	gen.baseMultiplierMin = gen.compounding.AdjustPriceMultiplier(
		gen.baseMultiplierMin, true,
	)
	gen.baseMultiplierMax = gen.compounding.AdjustPriceMultiplier(
		gen.baseMultiplierMax, false,
	)

	// Same with the historical multipliers
	gen.historicalMultiplierMin = gen.compounding.AdjustPriceMultiplier(
		gen.historicalMultiplierMin, true,
	)
	gen.historicalMultiplierMax = gen.compounding.AdjustPriceMultiplier(
		gen.historicalMultiplierMax, false,
	)

	// We need to updatePrices the bin width here, as the likelihood of repeated
	// lower bounds is compounding. To do that we need to know the price for
	// this period.
	var subBinWidthMin float64
	var subBinWidthMax float64
	gen.priceMin, subBinWidthMin = gen.calcNookPriceAndWidth(
		gen.historicalMultiplierMin,
	)
	gen.priceMax, subBinWidthMax = gen.calcNookPriceAndWidth(
		gen.historicalMultiplierMax,
	)

	gen.binWidthMin *= subBinWidthMin
	gen.binWidthMax *= subBinWidthMax
}

func (gen *phasePeriodGenerator) buildCurrentPeriod() *PotentialPricePeriod {
	possibilityCount := gen.priceMax - gen.priceMin + 1

	// Every  number  that is not the min or max has a width of 1, so the total width
	// of the mid range is all the possible prices - 2
	midWidth := float64(possibilityCount - 2)
	minWidth := gen.binWidthMin
	maxWidth := gen.binWidthMax

	totalWidth := minWidth + midWidth + maxWidth

	// To get the final chances take the min, mid, and max widths and divide them by
	// the total width
	minChance := minWidth / totalWidth
	midChance := midWidth / totalWidth
	maxChance := maxWidth / totalWidth

	var isSpike, isBigSpike, isSmallSpike bool
	if gen.hasSpike != nil {
		isSpike, isBigSpike = gen.hasSpike.IsSpike(gen.subPeriod)
		isSmallSpike = isSpike && !isBigSpike
	}

	return &PotentialPricePeriod{
		prices: &prices{
			minPrice: gen.priceMin,
			maxPrice: gen.priceMax,

			minChance: minChance,
			midChance: midChance,
			maxChance: maxChance,
		},
		Spikes: &Spikes{
			hasSpikeAny:   isSpike,
			hasSpikeBig:   isBigSpike,
			hasSpikeSmall: isSmallSpike,
		},
		PricePeriod:  gen.pricePeriod,
		PatternPhase: gen.PhaseFull,
	}
}

// Yields the next potential price period.
func (gen *phasePeriodGenerator) Next() *PotentialPricePeriod {
	// Do some set up work to start this iteration
	gen.beginThisIteration()

	// BIN WIDTH
	// We want to figure out the chance width BEFORE we make the final adjustment. The
	// adjustment always gets made uniformly, so its really the chance of the
	// pre-adjusted max and isMin we need to compute
	//
	// The chance width of a price is the rounded price minus the non-rounded extreme
	// price value. Prices in the middle of a range will always have a bin width of 1.
	//
	// This is important for figuring out the likelihood we are in a pattern. If we
	// have a purchase price of 100 bells, and a buy price of 90 bells in a price period
	// where the random multiplier is between 0.9 and 100, we know the chance of 90
	// bells occurring is essentially 0, since the random float generator would have to
	// return EXACTLY 0.9 out of many millions of possible values.
	if gen.compounding == nil || gen.subPeriod == 0 {
		gen.priceMin, gen.binWidthMin = gen.calcNookPriceAndWidth(gen.baseMultiplierMin)
		gen.priceMax, gen.binWidthMax = gen.calcNookPriceAndWidth(gen.baseMultiplierMax)
	}

	// If this is a compounded price, we need to adjust all our values by the minimum
	// and maximum sub period multiplier.
	if gen.compounding != nil && gen.subPeriod > 0 {
		gen.adjustCompounding()
	}

	gen.priceMin += gen.finalAdjustment
	gen.priceMax += gen.finalAdjustment

	potentialPeriod := gen.buildCurrentPeriod()

	gen.endThisIteration()

	return potentialPeriod
}
