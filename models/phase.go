package models

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
	BasePriceMultiplier(subPeriod int) (min float64, max float64)

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
	AdjustPriceMultiplier(factor float64, min bool) float64
}

// A phase may implement this interface if a final adjustment to the buying price
// should be made after applying BasePriceMultiplier() and SubPeriodPriceMultiplier().
// In practice only the increasing phase of the Small Spike pattern will need to
// implement this interface.
type phaseMakesFinalAdjustment interface {
	FinalPriceAdjustment(subPeriod int) int
}

type patternPhaseAuto struct {
	phaseImplement
}

func (phase *patternPhaseAuto) calcPhasePeriodPrice(
	baseMultiplier float64,
	purchasePrice, phasePeriod int,
	finalAdjustment int,
	min bool,
) (price int) {
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
	if compounding, ok := phase.phaseImplement.(phaseCompoundingPrice); ok {
		for i := 0; i < phasePeriod; i++ {
			baseMultiplier = compounding.AdjustPriceMultiplier(
				baseMultiplier, min,
			)
		}
	}

	price = RoundBells(float64(purchasePrice) * baseMultiplier)
	price += finalAdjustment
	return price
}

func (phase *patternPhaseAuto) potentialPrice(
	purchasePrice int, phasePeriod int,
) (minPrice int, maxPrice int) {
	baseMinFactor, baseMaxFactor := phase.BasePriceMultiplier(phasePeriod)

	var finalAdjustment int
	if makesAdjustment, ok := phase.phaseImplement.(phaseMakesFinalAdjustment); ok {
		finalAdjustment = makesAdjustment.FinalPriceAdjustment(phasePeriod)
	}

	minPrice = phase.calcPhasePeriodPrice(
		baseMinFactor, purchasePrice, phasePeriod, finalAdjustment, true,
	)
	maxPrice = phase.calcPhasePeriodPrice(
		baseMaxFactor, purchasePrice, phasePeriod, finalAdjustment, false,
	)

	return minPrice, maxPrice
}

func (phase *patternPhaseAuto) PotentialPeriod(
	period PricePeriod, phasePeriod int,
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

	minPrice, maxPrice := phase.potentialPrice(purchasePrice, phasePeriod)
	if phase.Ticker().PurchasePrice == 0 {
		// Now, if no purchase price was supplied, we need to run the numbers again
		// with the highest possible base price to get the max bracket for what we
		// know.
		_, maxPrice = phase.potentialPrice(110, phasePeriod)
	}

	return &PotentialPricePeriod{
		prices: prices{
			min: minPrice,
			max: maxPrice,
		},
		PricePeriod: period,
	}
}

func (phase *patternPhaseAuto) Duplicate() PatternPhase {
	return &patternPhaseAuto{
		phase.phaseImplement.Duplicate(),
	}
}
