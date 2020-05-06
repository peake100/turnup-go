package models

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
