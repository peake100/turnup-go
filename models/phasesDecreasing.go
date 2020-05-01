package models

// We only need to implement a single phase for this, since the whole week follows one
// pattern.
type decreasingPattern struct {
	phaseCoreAuto
}

func (phase *decreasingPattern) Name() string {
	return "whomp whomp"
}

func (phase *decreasingPattern) PossibleLengths(
	phases []PatternPhase,
) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{12}
}

func (phase *decreasingPattern) BasePriceMultiplier(subPeriod int) (min float64, max float64) {
	return 0.85, 0.90
}

func (phase *decreasingPattern) SubPeriodPriceMultiplier() (min float64, max float64) {
	return -0.05, -0.03
}

func (phase *decreasingPattern) Duplicate() phaseImplement {
	return &decreasingPattern{
		phase.phaseCoreAuto,
	}
}

// Generates a new set of decreasing phases to branch possible weeks off of.
func decreasingProgression(ticker *PriceTicker) []PatternPhase {
	phases := []PatternPhase{
		&patternPhaseAuto{new(decreasingPattern)},
	}

	for _, thisPhase := range phases {
		thisPhase.SetTicker(ticker)
	}

	return phases
}
