package models

// FLUCTUATING ///////////////////////

// Every increasing phase follows the same bell-price formula, so we will implement
// the price range values once in a base type and embed that in our specific phases
type increasingPhaseBase struct {
	phaseCoreAuto
}

func (phase *increasingPhaseBase) BasePriceMultiplier(int) (min float32, max float32) {
	return 0.9, 1.4
}

func (phase *increasingPhaseBase) DuplicateBase() increasingPhaseBase {
	return increasingPhaseBase{
		phase.phaseCoreAuto,
	}
}

// Every decreasing phase follows the same bell-price formula, so we will implement
// the price range values once in a base type and embed that in our specific phases
type decreasingPhaseBase struct {
	phaseCoreAuto
}

func (phase *decreasingPhaseBase) BasePriceMultiplier(int) (min float32, max float32) {
	return 0.6, 0.8
}

func (phase *decreasingPhaseBase) AdjustPriceMultiplier(
	factor float32, isMin bool,
) float32 {
	if isMin {
		return factor - 0.1
	}
	return factor - 0.04
}

func (phase *decreasingPhaseBase) DuplicateBase() decreasingPhaseBase {
	return decreasingPhaseBase{
		phase.phaseCoreAuto,
	}
}

// INCREASING PHASE 1
type increasing1 struct {
	increasingPhaseBase
}

func (phase *increasing1) Name() string {
	return "mild increase"
}

func (phase *increasing1) PossibleLengths([]PatternPhase) (possibilities []int) {
	// We only are going to call this possibility once, so we can finalize it
	phase.PossibilitiesComplete()
	return []int{0, 1, 2, 3, 4, 5, 6}
}

func (phase *increasing1) MaxLength() int {
	return 6
}

func (phase *increasing1) Duplicate() phaseImplement {
	return &increasing1{
		phase.DuplicateBase(),
	}
}

// DECREASING PHASE 1
type decreasing1 struct {
	decreasingPhaseBase
}

func (phase *decreasing1) Name() string {
	return "mild decrease"
}

func (phase *decreasing1) PossibleLengths([]PatternPhase) (possibilities []int) {
	// We only are going to call this possibility once, so we can finalize it
	phase.PossibilitiesComplete()
	return []int{2, 3}
}

func (phase *decreasing1) MaxLength() int {
	return 3
}

func (phase *decreasing1) Duplicate() phaseImplement {
	return &decreasing1{
		phase.DuplicateBase(),
	}
}

// INCREASING PHASE 2
type increasing2 struct {
	increasingPhaseBase
}

func (phase *increasing2) Name() string {
	return "mild increase"
}

func (phase *increasing2) PossibleLengths(
	phases []PatternPhase,
) (possibilities []int) {
	phase.IncrementPass()

	switch {

	case phase.Pass() == 1:
		// On the first pass, we return a temporary length of 7 - the length of
		// increasing phase 1.
		return []int{7 - phases[0].Length()}

	case phases[4].IsFinal():
		// The next time we will have enough information to give possibilities is when
		// increasing phase 3 has a length value, since we need it to do our
		// computation.
		//
		// Once we have it we subtract increasing phase 3's length from our temp
		// length.
		//
		// After we return, we are done on the final pass.
		phase.PossibilitiesComplete()
		return []int{phase.Length() - phases[4].Length()}

	default:
		// Otherwise we are waiting for increasing phase 3 to resolve, return no
		// possibilities, but report we are not done.
		return nil
	}

}

func (phase *increasing2) MaxLength() int {
	return 7
}

func (phase *increasing2) Duplicate() phaseImplement {
	return &increasing2{
		phase.DuplicateBase(),
	}
}

// DECREASING PHASE 2
type decreasing2 struct {
	decreasingPhaseBase
}

func (phase *decreasing2) Name() string {
	return "mild decrease"
}

func (phase *decreasing2) PossibleLengths(
	phases []PatternPhase,
) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{5 - phases[1].Length()}
}

func (phase *decreasing2) MaxLength() int {
	return 3
}

func (phase *decreasing2) Duplicate() phaseImplement {
	return &decreasing2{
		phase.DuplicateBase(),
	}
}

// INCREASING PHASE 3
type increasing3 struct {
	increasingPhaseBase
}

func (phase *increasing3) Name() string {
	return "mild increase"
}

func (phase *increasing3) PossibleLengths(
	phases []PatternPhase,
) (possibilities []int) {
	// This phase is a random length between 0 and the temp length of Increasing
	// phase 2 - 1
	minDays := 0
	maxDays := phases[2].Length() - 1

	for i := minDays; i <= maxDays; i++ {
		possibilities = append(possibilities, i)
	}

	phase.PossibilitiesComplete()
	return possibilities
}

func (phase *increasing3) MaxLength() int {
	return 6
}

func (phase *increasing3) Duplicate() phaseImplement {
	return &increasing3{
		phase.DuplicateBase(),
	}
}

// Generates a new set of fluctuating phases to branch possible weeks off of.
func fluctuatingProgression(ticker *PriceTicker) []PatternPhase {
	phases := []PatternPhase{
		&patternPhaseAuto{phaseImplement: new(increasing1)},
		&patternPhaseAuto{phaseImplement: new(decreasing1)},
		&patternPhaseAuto{phaseImplement: new(increasing2)},
		&patternPhaseAuto{phaseImplement: new(decreasing2)},
		&patternPhaseAuto{phaseImplement: new(increasing3)},
	}

	for _, thisPhase := range phases {
		thisPhase.SetTicker(ticker)
	}

	return phases
}
