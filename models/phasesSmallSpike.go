package models

// There are two decreasing phases in this pattern that use the same price calculations,
// so we are going to make an embeddable type for them.
type smallSpikeDecreasingBase struct {
	phaseCoreAuto
}

func (phase *smallSpikeDecreasingBase) BasePriceMultiplier(subPeriod int) (min float64, max float64) {
	return 0.4, 0.9
}

func (phase *smallSpikeDecreasingBase) SubPeriodPriceMultiplier() (min float64, max float64) {
	return -0.03 - 0.02, -0.03
}

// DECREASING PHASE 1
type smallSpikeDecreasing1 struct {
	smallSpikeDecreasingBase
}

func (phase *smallSpikeDecreasing1) Name() string {
	return "steady decrease"
}

func (phase *smallSpikeDecreasing1) PossibleLengths(
	phases []PatternPhase,
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

func (phase *smallSpikeIncreasing) Name() string {
	return "steady increase"
}

func (phase *smallSpikeIncreasing) PossibleLengths(
	phases []PatternPhase,
) (possibilities []int) {
	phase.PossibilitiesComplete()
	return []int{5}
}

func (phase *smallSpikeIncreasing) SubPeriodPriceMultiplier() (min float64, max float64) {
	return 0, 0
}

func (phase *smallSpikeIncreasing) PotentialPeriod(
	period PricePeriod, phasePeriod int,
) *PotentialPricePeriod {
	minFactor, maxFactor := 1.4, 2.0
	var priceAdjustment int

	switch {
	case phasePeriod == 0 || phasePeriod == 1:
		minFactor, maxFactor = 0.9, 1.4
	case phasePeriod == 2 || phasePeriod == 4:
		priceAdjustment = -1
	case phasePeriod > 4:
		panic("steady increase only has 5 sub periods")
	}
	
	minPrice := RoundBells(float64(phase.ticker.PurchasePrice) * minFactor) +
		priceAdjustment

	maxPrice := RoundBells(float64(phase.ticker.PurchasePrice) * maxFactor) +
		priceAdjustment
	
	return &PotentialPricePeriod{
		prices:      prices{
			min: minPrice,
			max: maxPrice,
		},
		PricePeriod: 0,
	}
}

func (phase *smallSpikeIncreasing) Duplicate() PatternPhase {
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
		new(smallSpikeIncreasing),
		&patternPhaseAuto{new(smallSpikeDecreasing2)},
	}

	for _, thisPhase := range phases {
		thisPhase.SetTicker(ticker)
	}

	return phases
}
