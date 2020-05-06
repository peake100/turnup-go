package models

import "math"

// This struct can be embedded with an implemented phase to complete the full phase
// implementation and get price period calculations for free
type patternPhaseAuto struct {
	phaseImplement
	// We are going to cache potential price period info here so we only need to
	// generate it once. It's a mapping of subperiod -> potential price period
	potentialSubPeriods map[int]*potentialPhaseSubPeriod
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

	potentialSubPeriod := &potentialPhaseSubPeriod{
		prices:       *prices,
		Spike: Spike{
			hasSpikeAny:   isSpike,
			hasSpikeBig:   isBigSpike,
			hasSpikeSmall: isSmallSpike,
		},
		PatternPhase: phase,
	}

	phase.potentialSubPeriods[subPeriod] = potentialSubPeriod

	return &PotentialPricePeriod{
		potentialPhaseSubPeriod: potentialSubPeriod,
		PricePeriod:             period,
	}
}

func (phase *patternPhaseAuto) PotentialPeriod(
	period PricePeriod, subPeriod int,
) *PotentialPricePeriod {
	if phase.potentialSubPeriods == nil {
		phase.potentialSubPeriods = make(map[int]*potentialPhaseSubPeriod)
	}

	if potentialSubPeriod, ok := phase.potentialSubPeriods[subPeriod] ; ok {
		return &PotentialPricePeriod{
			potentialPhaseSubPeriod: potentialSubPeriod,
			PricePeriod:             period,
		}
	}

	return phase.generateSubPeriod(period, subPeriod)
}

func (phase *patternPhaseAuto) Duplicate() PatternPhase {
	return &patternPhaseAuto{
		phaseImplement:      phase.phaseImplement.Duplicate(),
		potentialSubPeriods: phase.potentialSubPeriods,
	}
}
