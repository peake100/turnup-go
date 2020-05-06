package models

import "golang.org/x/xerrors"

// We only need to implement a single phase for this, since the whole week follows one
// pattern.
type decreasingPattern struct {
	phaseCoreAuto
}

func (phase *decreasingPattern) Name() string {
	return "whomp whomp"
}

func (phase *decreasingPattern) PossibleLengths([]PatternPhase) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{12}
}

func (phase *decreasingPattern) MaxLength() int {
	return 12
}

func (phase *decreasingPattern) BasePriceMultiplier(int) (min float32, max float32) {
	return 0.85, 0.90
}

func (phase *decreasingPattern) AdjustPriceMultiplier(
	factor float32, isMin bool,
) float32 {
	if isMin {
		return factor - 0.05
	}
	return factor - 0.03
}

func (phase *decreasingPattern) Duplicate() phaseImplement {
	// We should never need to duplicate a decreasing stage because there is only ont
	// price pattern for the decreasing pattern. We'll set it to panic if it get's
	// accessed
	panic(xerrors.New("decreasing phase should never be duplicated"))
}

// Generates a new set of decreasing phases to branch possible weeks off of.
func decreasingProgression(ticker *PriceTicker) []PatternPhase {
	phases := []PatternPhase{
		&patternPhaseAuto{phaseImplement: new(decreasingPattern)},
	}

	for _, thisPhase := range phases {
		thisPhase.SetTicker(ticker)
	}

	return phases
}
