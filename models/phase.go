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

	// Returns a potential price bracket for a given day of this phase. ``period`` is
	// the absolute period for the week, while ``subPeriod`` is the price period
	// relative to the start of this phase, beginning at 0.
	PotentialPeriod(
		period PricePeriod, subPeriod int,
	) *PotentialPricePeriod

	// Creates a duplicate of this phase in the current state. Used for making
	// permutations.
	Duplicate() PatternPhase
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

	// The max length this phase can be, used to pre-cache price period values.
	MaxLength() int

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
// In practice only the increasing phase of the Small Spike pattern will need to
// implement this interface.
type phaseMakesFinalAdjustment interface {
	FinalPriceAdjustment(subPeriod int) int
}

// A phase can implement this method if it has a price spike in it's bounds. Returns
// whether a given sub period is a spiked price, and whether it's a large spike or small
// spike. NOTE: Small spikes are defined as the peak price day for the small spike
// pattern, along with the day to either side, as the potential prices for that day are
// just one bell less than the peak price itself.
type phaseHasSpike interface {
	IsSpike(subPeriod int) (isSpike bool, isBig bool)
}
