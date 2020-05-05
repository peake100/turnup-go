package models

// There are two decreasing phases in this pattern that use the same price calculations,
// so we are going to make an embeddable type for them.
type smallSpikeDecreasingBase struct {
	phaseCoreAuto
}

func (phase *smallSpikeDecreasingBase) AdjustPriceMultiplier(
	factor float32, isMin bool,
) float32 {
	if isMin {
		// In order to match the EXACT calculations from the game, we need to subtract
		// both 0.02 and 0.03 discreetly, otherwise we end up with a SLIGHTLY different
		// float value that can result in a perice different from what the game would
		// yield.
		return factor - 0.02 - 0.03
	}
	return factor - 0.03
}

func (phase *smallSpikeDecreasingBase) BasePriceMultiplier(int) (
	min float32, max float32,
) {
	return 0.4, 0.9
}

// DECREASING PHASE 1
type smallSpikeDecreasing1 struct {
	smallSpikeDecreasingBase
}

func (phase *smallSpikeDecreasing1) Name() string {
	return "steady decrease"
}

func (phase *smallSpikeDecreasing1) PossibleLengths(
	[]PatternPhase,
) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{0, 1, 2, 3, 4, 5, 6, 7}
}

func (phase *smallSpikeDecreasing1) Duplicate() phaseImplement {
	return &smallSpikeDecreasing1{
		smallSpikeDecreasingBase{
			phase.smallSpikeDecreasingBase.phaseCoreAuto,
		},
	}
}

// INCREASING PHASE
type smallSpikeIncreasing struct {
	phaseCoreAuto
}

func (phase *smallSpikeIncreasing) FinalPriceAdjustment(subPeriod int) int {
	if subPeriod == 2 || subPeriod == 4 {
		return -1
	}
	// For period 3 and 5, we subtract 1 from the total after doing our calculation.
	return 0
}

func (phase *smallSpikeIncreasing) BasePriceMultiplier(
	subPeriod int,
) (min float32, max float32) {
	switch {
	case subPeriod == 0 || subPeriod == 1:
		// Periods 1 and 2 are random between 90% and and 140%.
		return 0.9, 1.4
	default:
		// The rest of the phase periods are random between 140% and 200%
		return 1.4, 2.0
	}
}

func (phase *smallSpikeIncreasing) Name() string {
	return "slight spike"
}

func (phase *smallSpikeIncreasing) PossibleLengths(
	[]PatternPhase,
) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{5}
}

func (phase *smallSpikeIncreasing) Duplicate() phaseImplement {
	return &smallSpikeIncreasing{
		phase.phaseCoreAuto,
	}
}

// DECREASING PHASE 1
type smallSpikeDecreasing2 struct {
	smallSpikeDecreasingBase
}

func (phase *smallSpikeDecreasing2) Name() string {
	return "steady decrease"
}

func (phase *smallSpikeDecreasing2) PossibleLengths(
	phases []PatternPhase,
) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{7 - phases[0].Length()}
}

func (phase *smallSpikeDecreasing2) Duplicate() phaseImplement {
	return &smallSpikeDecreasing2{
		smallSpikeDecreasingBase{
			phase.smallSpikeDecreasingBase.phaseCoreAuto,
		},
	}
}

// Generates a new set of fluctuating phases to branch possible weeks off of.
func smallSpikeProgression(ticker *PriceTicker) []PatternPhase {
	phases := []PatternPhase{
		&patternPhaseAuto{new(smallSpikeDecreasing1)},
		&patternPhaseAuto{new(smallSpikeIncreasing)},
		&patternPhaseAuto{new(smallSpikeDecreasing2)},
	}

	for _, thisPhase := range phases {
		thisPhase.SetTicker(ticker)
	}

	return phases
}
