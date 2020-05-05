package models

// STEADY DECREASE
type steadyDecrease struct {
	phaseCoreAuto
}

func (phase *steadyDecrease) Name() string {
	return "steady decrease"
}

func (phase *steadyDecrease) PossibleLengths([]PatternPhase) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{1, 2, 3, 4, 5, 6, 7}
}

func (phase *steadyDecrease) BasePriceMultiplier(int) (min float32, max float32) {
	return 0.85, 0.9
}

func (phase *steadyDecrease) AdjustPriceMultiplier(factor float32, isMin bool) float32 {
	if isMin {
		return factor - 0.05
	}
	return factor - 0.03
}

func (phase *steadyDecrease) Duplicate() phaseImplement {
	return &steadyDecrease{
		phase.phaseCoreAuto,
	}
}

// SHARP INCREASE
type sharpIncrease struct {
	phaseCoreAuto
}

func (phase *sharpIncrease) IsSpike(subPeriod int) (isSpike bool, isBig bool) {
	if subPeriod == 2 {
		return true, true
	}
	return false, false
}

func (phase *sharpIncrease) Name() string {
	return "sharp increase"
}

func (phase *sharpIncrease) PossibleLengths([]PatternPhase) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{3}
}

func (phase *sharpIncrease) BasePriceMultiplier(
	subPeriod int,
) (min float32, max float32) {
	switch {
	case subPeriod == 0:
		return 0.9, 1.4
	case subPeriod == 1:
		return 1.4, 2
	default:
		return 2, 6
	}
}

func (phase *sharpIncrease) Duplicate() phaseImplement {
	return &sharpIncrease{
		phase.phaseCoreAuto,
	}
}

// SHARP DECREASE
type sharpDecrease struct {
	phaseCoreAuto
}

func (phase *sharpDecrease) Name() string {
	return "sharp decrease"
}

func (phase *sharpDecrease) PossibleLengths([]PatternPhase) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{2}
}

func (phase *sharpDecrease) BasePriceMultiplier(
	subPeriod int,
) (min float32, max float32) {
	if subPeriod == 0 {
		return 1.4, 2
	}

	return 0.9, 1.4
}

func (phase *sharpDecrease) Duplicate() phaseImplement {
	return &sharpDecrease{
		phase.phaseCoreAuto,
	}
}

// RANDOM DECREASE
type randomDecrease struct {
	phaseCoreAuto
}

func (phase *randomDecrease) Name() string {
	return "random low"
}

func (phase *randomDecrease) PossibleLengths(
	phases []PatternPhase,
) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{12 - phases[0].Length() - 5}
}

func (phase *randomDecrease) BasePriceMultiplier(int) (min float32, max float32) {
	return 0.4, 0.9
}

func (phase *randomDecrease) Duplicate() phaseImplement {
	return &randomDecrease{
		phase.phaseCoreAuto,
	}
}

// Generates a new set of fluctuating phases to branch possible weeks off of.
func bigSpikeProgression(ticker *PriceTicker) []PatternPhase {
	phases := []PatternPhase{
		&patternPhaseAuto{new(steadyDecrease)},
		&patternPhaseAuto{new(sharpIncrease)},
		&patternPhaseAuto{new(sharpDecrease)},
		&patternPhaseAuto{new(randomDecrease)},
	}

	for _, thisPhase := range phases {
		thisPhase.SetTicker(ticker)
	}

	return phases
}
