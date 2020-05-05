package models

import "math"

// A phase is a period of time within a price pattern that follows a single algorithm.
// When making predictions for a given pattern, we will iterate over a set of phases.
//
// A phase is responsible for:
//
//		1. Communicating a set of possible lengths.
//		2. Reporting if it's length has been set in stone when the predictor is
//		   iterating over the phases to set all possible lengths.
//      3. Returning the price range for a given price period within itself.
//      4. Copying itself for spawning a new set of phase length possibilities.
//
// This interface describes the methods necessary to accomplish these four goals,
// and is used by the predictor to map out all possible phase combinations, and get all
// possible price ranges for a given price period.
type PatternPhase interface {
	// The name of the phase
	Name() string

	// The predictor will set the ticker during setup to make it available for
	// calculations. The phase, in turn promises NOT to mutate the ticker.
	SetTicker(ticker *PriceTicker)

	// Returns a list of possible lengths. Should return nil if it cannot yet be
	// determined. 'lengthPass' is a counter of how many times the list of phases has
	// been passed over when computing the possible lengths. For each possible length
	// returned, a new goroutine will be spawned to compute that possibility by making
	// a copy of `phases` and calling 'set length' on this pattern.
	//
	// Should return 'nil' for ``possibilities`` if possibilities cannot be computed
	// for this pass. Should panic if we are calling on a finalized phase.
	PossibleLengths(phases []PatternPhase) (possibilities []int)

	// Sets the length we want to assume for this phase. This does not need be the
	// final length, many phases go through a temp length. This method is called by
	// the predictor when setting up a series of possible phase combinations.
	SetLength(length int)

	// Returns the length set by ``.SetLength()`` for other phases to inspect when
	// making calculations.
	Length() int

	// Whether the value returned by .Length() is the final length.
	IsFinal() bool

	// Returns a potential price bracket for a given day of this Phase. ``period`` is
	// the absolute period for the week, while ``subPeriod`` is the price period
	// relative to the start of this phase, beginning at 0.
	PotentialPeriod(
		period PricePeriod, subPeriod int,
	) *PotentialPricePeriod

	// Creates a duplicate of this phase in the current state. Used for making
	// permutations.
	Duplicate() PatternPhase
}

// There are a few methods of PatternPhase we can implement uniformly, and that we also
// would want access to to implement the unique logic for a given phase. We can embed
// this type in our specific implementations in order to gain access to these methods.
type phaseCoreAuto struct {
	ticker *PriceTicker

	// The current length in price periods for this phase.
	length int
	// Should be incremented every time `PossibleLengths` is called. Used to determine
	// what length calculation pass we are on
	pass int

	// Set when there will be no further possible lengths
	possibilitiesComplete bool
	isFinal               bool
}

// Called by the predictor when setting up a prediction. Sets the real-work pricing
// information.
func (phase *phaseCoreAuto) SetTicker(ticker *PriceTicker) {
	phase.ticker = ticker
}

// Returns the ticker set by SetTicker()
func (phase *phaseCoreAuto) Ticker() *PriceTicker {
	return phase.ticker
}

// How many times IncrementPass() has been called. The intended usage is to track
// how many times the predictor has requested PossibleLengths() while iterating over
// potential lengths.
func (phase *phaseCoreAuto) Pass() int {
	return phase.pass
}

// To be called when PossibleLengths() is called, allows us to track how many times
// the predictor has iterated over this phase looking for possible lengths.
func (phase *phaseCoreAuto) IncrementPass() {
	phase.pass++
}

// The length of this phase. This length may or may not be the final length, as some
// phases go through multiple temporary lengths before the calculation is done.
func (phase *phaseCoreAuto) Length() int {
	return phase.length
}

// Called by the predictor to set a potential length for this pattern
func (phase *phaseCoreAuto) SetLength(length int) {
	phase.length = length
	// If there are not going to be any more possibilities then this is the final
	// length, not a temp length. This phase is finalized.
	if phase.possibilitiesComplete {
		phase.isFinal = true
	}
}

// If true, there the current length is the final length
func (phase *phaseCoreAuto) IsFinal() bool {
	return phase.isFinal
}

// After this is called, IsFinal() will return true
func (phase *phaseCoreAuto) PossibilitiesComplete() {
	phase.possibilitiesComplete = true
}

// These are the methods we need to implement in order to be wrapped by
// phaseCoreAuto, which has default implementations for reporting prices.
//
// NOTE: many of these methods can be implemented for free by embedding phaseCoreAuto
// in a phase implementation.
type phaseImplement interface {
	// IMPLEMENTED BY phaseCoreAuto. See for descriptions.
	SetTicker(ticker *PriceTicker)
	SetLength(length int)
	Length() int
	Ticker() *PriceTicker
	IsFinal() bool

	// MUST BE IMPLEMENTED UNIQUELY
	//
	// See descriptions in PatternPhase interface above.
	Name() string
	PossibleLengths(phases []PatternPhase) (possibilities []int)

	// Returns the min and max base price multipliers for a given subPeriod within the
	// phase. This function is the real point of this interface. By implementing this
	// function, patternPhaseAuto can do most of the math for us to calculate the price
	// range.
	BasePriceMultiplier(subPeriod int) (min float32, max float32)

	// Create a copy of this object
	Duplicate() phaseImplement
}

// If this is a price phase that gradually improves or degrades, return the
// min and max factor by which this occurs. Only needs to be implemented by phases
// that loose or gain a percentage of the previous period's price while the phase
// is active
type phaseCompoundingPrice interface {
	// Applies the phase period multiplier to the current min and max factors. Will be
	// called once for each phase period after the first.
	//
	// We need to implement this as an addition rather than a value return because some
	// sub period multipliers are created from adding multiple floats on each iteration,
	// so returning a single value to add to the base factor results in floating point
	// rounding errors.
	//
	// `Min` is set to true when this is the minimum factor.
	AdjustPriceMultiplier(factor float32, isMin bool) float32
}

// A phase may implement this interface if a final adjustment to the buying price
// should be made after applying BasePriceMultiplier() and SubPeriodPriceMultiplier().
// In practice only the increasing phase of the Small HasSpikeAny pattern will need to
// implement this interface.
type phaseMakesFinalAdjustment interface {
	FinalPriceAdjustment(subPeriod int) int
}

// A phase can implement this method if it has a price hasSpikeAny in it's bounds. Returns
// whether a given sub period is a spiked price, and whether it's a large hasSpikeAny or small
// hasSpikeAny. NOTE: Small spikes are defined as the peak price day for the small hasSpikeAny
// pattern, along with the day to either side, as the potential prices for that day are
// just one bell less than the peak price itself.
type phaseHasSpike interface {
	IsSpike(subPeriod int) (isSpike bool, isBig bool)
}

type patternPhaseAuto struct {
	phaseImplement
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

func (phase *patternPhaseAuto) PotentialPeriod(
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

	return &PotentialPricePeriod{
		prices:       *prices,
		PricePeriod:  period,
		PatternPhase: phase,

		Spike: Spike{
			hasSpikeAny:   isSpike,
			hasSpikeBig:   isBigSpike,
			hasSpikeSmall: isSmallSpike,
		},
	}
}

func (phase *patternPhaseAuto) Duplicate() PatternPhase {
	return &patternPhaseAuto{
		phase.phaseImplement.Duplicate(),
	}
}
