package models

import (
	"math"
)

// We're going to use this object to store information about our sub period calculation
// so we don't need to store a ton of info
type subPeriodsGenerator struct {
	Ticker           *PriceTicker
	PhaseFull        *patternPhaseAuto
	PricePeriodStart PricePeriod

	// For outside consumption
	LastCompletedSubPeriod int

	// Value cache
	purchasePrice int

	// We're going to cache a few type conversions
	phase       phaseImplement
	compounding phaseCompoundingPrice
	makesFinal  phaseMakesFinalAdjustment
	hasSpike    phaseHasSpike

	// Data fields - these fields will be updated during processing
	// Whether this iterator has been started
	started bool

	subPeriod int
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

func (gen *subPeriodsGenerator) Setup() {
	gen.pricePeriod = gen.PricePeriodStart
	gen.purchasePrice = gen.Ticker.PurchasePrice

	gen.phase = gen.PhaseFull.phaseImplement
	if compounding, ok := gen.phase.(phaseCompoundingPrice) ; ok {
		gen.compounding = compounding
	}
	if makesFinal, ok := gen.phase.(phaseMakesFinalAdjustment) ; ok {
		gen.makesFinal = makesFinal
	}
	if hasSpike, ok := gen.phase.(phaseHasSpike) ; ok {
		gen.hasSpike = hasSpike
	}

	gen.LastCompletedSubPeriod = -1
	gen.started = true
}

func (gen *subPeriodsGenerator) beginThisIteration() {
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

func (gen *subPeriodsGenerator) endThisIteration() {
	// if we are beyond phase one, set the previous price so the compounding
	// calculations have access to it.
	gen.previousPeriodPrice = gen.Ticker.Prices[gen.pricePeriod]
	gen.LastCompletedSubPeriod = gen.subPeriod
	gen.subPeriod++
	gen.pricePeriod++
}

// Calculates the actual multiplier value if we know the price from the previous day.
func (gen *subPeriodsGenerator) calcPhasePeriodHistoricalMultiplier(
	isMin bool,
) (historicMultiplier float32) {
	previousPrice := gen.previousPeriodPrice
	// Un-adjust this price if it has an adjustment (price adjustments
	// never happen in the actual game during compounding phases, but we'll
	// put it here for max compatibility in case that ever changes with an
	// update, it should always be 0 when this code block is executed).
	previousPrice -= gen.finalAdjustment

	// We need to get the most extreme pre-rounded price that could have
	// resulted in the known price. For the max price, this is the price
	// itself. For isMin, this is the number - 1 + the smallest possible
	// float value.
	previousPriceFloat := float32(previousPrice)
	previousPriceFloor := previousPriceFloat - 1
	previousPriceFloat = math.Nextafter32(
		previousPriceFloor, previousPriceFloat,
	)

	// now work out the extreme end of the previous multiplier
	historicMultiplier = previousPriceFloat / float32(gen.purchasePrice)

	var baseMultiplier float32
	if isMin {
		baseMultiplier = gen.baseMultiplierMin
	} else {
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

func (gen *subPeriodsGenerator) calcNookPriceAndWidth(
	priceMultiplier float32,
) (price int, binWidth float64) {
	price = RoundBells(float32(gen.purchasePrice) * priceMultiplier)

	// Convert everything to float64 so our prediction math is more precise.
	binWidth = float64(price) -
		(float64(gen.purchasePrice) * float64(priceMultiplier))

	return price, binWidth
}

func (gen *subPeriodsGenerator) adjustCompounding() {
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
		gen.baseMultiplierMax, true,
	)

	// Same with the historical multipliers
	gen.historicalMultiplierMin = gen.compounding.AdjustPriceMultiplier(
		gen.historicalMultiplierMin, true,
	)
	gen.historicalMultiplierMax = gen.compounding.AdjustPriceMultiplier(
		gen.historicalMultiplierMax, true,
	)

	// We need to update the bin width here, as the likelihood of repeated
	// lower bounds is compounding. To do that we need to know the price for
	// this period.
	var subBinWidthMin float64
	var subBinWidthMax float64
	gen.priceMin, subBinWidthMin = gen.calcNookPriceAndWidth(gen.historicalMultiplierMin)
	gen.priceMax, subBinWidthMax = gen.calcNookPriceAndWidth(gen.historicalMultiplierMax)

	gen.binWidthMin *= subBinWidthMin
	gen.binWidthMax *= subBinWidthMax
}

func (gen *subPeriodsGenerator) buildCurrentPeriod() *PotentialPricePeriod {
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
		prices: prices{
			min: gen.priceMin,
			max: gen.priceMax,

			minChance: minChance,
			midChance: midChance,
			maxChance: maxChance,
		},
		Spike: Spike{
			hasSpikeAny:   isSpike,
			hasSpikeBig:   isBigSpike,
			hasSpikeSmall: isSmallSpike,
		},
		PricePeriod:  gen.pricePeriod,
		PatternPhase: gen.PhaseFull,
	}
}

// Yields the next potential price period.
func (gen *subPeriodsGenerator) Next() *PotentialPricePeriod {
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
		gen.priceMax, gen.binWidthMax = gen.calcNookPriceAndWidth(gen.baseMultiplierMin)
	}

	// If this is a compounded price, we need to adjust all our values by the minimum
	// and maximum sub period multiplier.
	if gen.compounding != nil && gen.subPeriod > 0 {
		gen.adjustCompounding()
	}

	gen.priceMin += gen.finalAdjustment
	gen.priceMax += gen.finalAdjustment

	gen.endThisIteration()

	return gen.buildCurrentPeriod()
}

// This struct can be embedded with an implemented phase to complete the full phase
// implementation and get price period calculations for free
type patternPhaseAuto struct {
	phaseImplement
	// For compounding phases, we have to generate
	potentialPeriods []*PotentialPricePeriod
}

// Calculates the actual multiplier value if we know the price from the previous day.
func (phase *patternPhaseAuto) calcPhasePeriodHistoricalMultiplier(
	purchasePrice int,
	previousPrice int,
	baseMultiplier float32,
	finalAdjustment int,
	isMin bool,
) (historicMultiplier float32) {
	// Un-adjust this price if it has an adjustment (price adjustments
	// never happen in the actual game during compounding phases, but we'll
	// put it here for max compatibility in case that ever changes with an
	// update, it should always be 0 when this code block is executed).
	previousPrice -= finalAdjustment

	// We need to get the most extreme pre-rounded price that could have
	// resulted in the known price. For the max price, this is the price
	// itself. For isMin, this is the number - 1 + the smallest possible
	// float value.
	previousPriceFloat := float32(previousPrice)
	previousPriceFloor := previousPriceFloat - 1
	previousPriceFloat = math.Nextafter32(
		previousPriceFloor, previousPriceFloat,
	)

	// now work out the extreme end of the previous multiplier
	historicMultiplier = previousPriceFloat / float32(purchasePrice)

	// if it is lower than the isMin multiplier or higher than the
	// max multiplier, we need to bring it in line with the
	// possible range
	if (isMin && historicMultiplier < baseMultiplier) ||
		(!isMin && historicMultiplier > baseMultiplier) {
		historicMultiplier = baseMultiplier
	}

	return historicMultiplier
}

func (phase *patternPhaseAuto) calcPhasePeriodPriceCompounding(
	compounding phaseCompoundingPrice,
	currentPrice int,
	binWidth float64,
	baseMultiplier float32,
	purchasePrice int,
	pricePeriod PricePeriod,
	phasePeriod int,
	finalAdjustment int,
	isMin bool,
) (compoundedPrice int, compoundedBinWidth float64) {
	// Get the starting price period for this phase
	compoundedPrice = currentPrice
	pricePeriod = pricePeriod - PricePeriod(phasePeriod)

	historicMultiplier := baseMultiplier

	// If phasePeriod is 0, this loop does not occur
	for i := 0; i < phasePeriod; i++ {
		// If we know the price for the day before, then we can make a more accurate
		// projection of what this period's prices will be by multiplying the real
		// world price by the upper and lower sub-period bounds. Each time we know
		// the real price for the previous price period, we are going to reset the
		// multiplier to
		previousPrice := phase.Ticker().Prices[pricePeriod]
		if previousPrice != 0 {
			historicMultiplier = phase.calcPhasePeriodHistoricalMultiplier(
				purchasePrice,
				previousPrice,
				baseMultiplier,
				finalAdjustment,
				isMin,
			)
		}

		// Alter the base multiplier by the phase period amount.
		baseMultiplier = compounding.AdjustPriceMultiplier(
			baseMultiplier, isMin,
		)

		historicMultiplier = compounding.AdjustPriceMultiplier(
			historicMultiplier, isMin,
		)

		// We need to update the bin width here, as the likelihood of repeated
		// lower bounds is compounding. To do that we need to know the price for
		// this period.
		//
		// Convert everything to float64 so our prediction math is more precise.
		compoundedPrice = RoundBells(float32(purchasePrice) * historicMultiplier)
		subBinWidth := float64(compoundedPrice) -
			(float64(purchasePrice) * float64(historicMultiplier))

		binWidth *= subBinWidth

		pricePeriod++
	}

	return compoundedPrice, binWidth
}

// Calculates the minimum or maximum price of a given period and the bin width (chance)
// of it happening.
func (phase *patternPhaseAuto) calcPhasePeriodPrice(
	baseMultiplier float32,
	purchasePrice int,
	pricePeriod PricePeriod,
	phasePeriod int,
	finalAdjustment int,
	isMin bool,
) (price int, binWidth float64) {
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
	price = RoundBells(float32(purchasePrice) * baseMultiplier)
	binWidth = float64(price) - (float64(purchasePrice) * float64(baseMultiplier))

	// We want to adjust our multiplier based on the price history, so we're going to
	// to keep track of an adjusted multiplier based on island prices. We ALSO need to
	// track the extreme possible ends of this pattern overall to bound this multiplier
	// when accounting for rounding.
	if compounding, ok := phase.phaseImplement.(phaseCompoundingPrice); ok {
		price, binWidth = phase.calcPhasePeriodPriceCompounding(
			compounding,
			price,
			binWidth,
			baseMultiplier,
			purchasePrice,
			pricePeriod,
			phasePeriod,
			finalAdjustment,
			isMin,
		)
	}

	// Make any final adjustment that needs to be made to the price after randomization.
	price += finalAdjustment
	return price, binWidth
}

func (phase *patternPhaseAuto) potentialPrice(
	purchasePrice int, pricePeriod PricePeriod, phasePeriod int,
) *prices {
	baseMinFactor, baseMaxFactor := phase.BasePriceMultiplier(phasePeriod)

	// Check if we need to make a final adjustment to a price
	var finalAdjustment int
	if makesAdjustment, ok := phase.phaseImplement.(phaseMakesFinalAdjustment); ok {
		finalAdjustment = makesAdjustment.FinalPriceAdjustment(phasePeriod)
	}

	minPrice, minWidth := phase.calcPhasePeriodPrice(
		baseMinFactor,
		purchasePrice,
		pricePeriod,
		phasePeriod,
		finalAdjustment,
		true,
	)
	maxPrice, maxWidth := phase.calcPhasePeriodPrice(
		baseMaxFactor,
		purchasePrice,
		pricePeriod,
		phasePeriod,
		finalAdjustment,
		false,
	)

	possibilityCount := maxPrice - minPrice + 1

	// Every  number  that is not the min or max has a width of 1, so the total width
	// of the mid range is all the possible prices - 2
	midWidth := float64(possibilityCount - 2)

	totalWidth := minWidth + midWidth + maxWidth

	// To get the final chances take the min, mid, and max widths and divide them by
	// the total width
	minChance := minWidth / totalWidth
	midChance := midWidth / totalWidth
	maxChance := maxWidth / totalWidth

	result := &prices{
		min: minPrice,
		max: maxPrice,

		minChance: minChance,
		midChance: midChance,
		maxChance: maxChance,
	}

	return result
}

func (phase *patternPhaseAuto) generateSubPeriod(
	period PricePeriod, subPeriod int,
) *PotentialPricePeriod {
	purchasePrice := phase.Ticker().PurchasePrice
	if purchasePrice == 0 {
		// If the purchase price is 0, then it is unknown. We need to compute the prices
		// for both the lowest and highest possible base price. The lowest possible
		// price is 90. Since sell prices are always a percentage of the purchase
		// price, the lower purchase price will always yield the lowest sell price and
		// vice versa.
		purchasePrice = 90
	}

	prices := phase.potentialPrice(purchasePrice, period, subPeriod)
	if phase.Ticker().PurchasePrice == 0 {
		// Now, if no purchase price was supplied, we need to run the numbers again
		// with the highest possible base price to get the max bracket for what we
		// know.
		pricesMax := phase.potentialPrice(110, period, subPeriod)
		prices.max = pricesMax.max
		prices.maxChance = pricesMax.maxChance
		prices.midChance = 1.0 - prices.minChance - prices.maxChance
	}

	var isSpike, isBigSpike, isSmallSpike bool
	if hasSpike, ok := phase.phaseImplement.(phaseHasSpike); ok {
		isSpike, isBigSpike = hasSpike.IsSpike(subPeriod)
		isSmallSpike = isSpike && !isBigSpike
	}

	potentialPeriod := &PotentialPricePeriod{
		prices: *prices,
		PricePeriod:       period,
		Spike: Spike{
			hasSpikeAny:   isSpike,
			hasSpikeBig:   isBigSpike,
			hasSpikeSmall: isSmallSpike,
		},
		PatternPhase: phase,
	}

	phase.potentialPeriods[subPeriod] = potentialPeriod

	return potentialPeriod
}

func (phase *patternPhaseAuto) PotentialPeriod(
	period PricePeriod, subPeriod int,
) *PotentialPricePeriod {
	if phase.potentialPeriods == nil {
		phase.potentialPeriods = make([]*PotentialPricePeriod, phase.MaxLength())
	}

	potentialPeriod := phase.potentialPeriods[subPeriod]
	if potentialPeriod != nil {
		return potentialPeriod
	}

	return phase.generateSubPeriod(period, subPeriod)
}

func (phase *patternPhaseAuto) Duplicate() PatternPhase {
	return &patternPhaseAuto{
		phaseImplement:      phase.phaseImplement.Duplicate(),
		// This cache will need to be created for each phase p
		potentialPeriods: nil,
	}
}
