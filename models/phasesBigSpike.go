package models

import "golang.org/x/xerrors"

// STEADY DECREASE
type steadyDecrease struct {
	phaseCoreAuto
}

func (phase *steadyDecrease) Name() string {
	return "steady decrease"
}

func (phase *steadyDecrease) PossibleLengths(
	phases []PatternPhase,
) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{1, 2, 3, 4, 5, 6, 7}
}

func (phase *steadyDecrease) BasePriceMultiplier(
	subPeriod int,
) (min float64, max float64) {
	return 0.85, 0.9
}

func (phase *steadyDecrease) SubPeriodPriceMultiplier(int) (min float64, max float64) {
	return -0.05, -0.03
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

func (phase *sharpIncrease) Name() string {
	return "sharp increase"
}

func (phase *sharpIncrease) PossibleLengths([]PatternPhase) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{3}
}

func (phase *sharpIncrease) BasePriceMultiplier(
	phasePeriod int,
) (min float64, max float64) {
	switch {
	case phasePeriod == 0:
		return 0.9, 1.4
	case phasePeriod == 1:
		return 1.4, 2
	case phasePeriod == 2:
		return 2, 6
	default:
		panic(xerrors.New("sharp increase only has 3 price periods"))
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

func (phase *sharpDecrease) PossibleLengths(
	phases []PatternPhase,
) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{2}
}

func (phase *sharpDecrease) BasePriceMultiplier(
	phasePeriod int,
) (min float64, max float64) {
	switch {
	case phasePeriod == 0:
		return 1.4, 2
	case phasePeriod == 1:
		return 0.9, 1.4
	default:
		panic(xerrors.New("sharp decrease only has 2 price periods"))
	}
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

func (phase *randomDecrease) BasePriceMultiplier(int) (min float64, max float64) {
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
