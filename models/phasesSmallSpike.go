package models

// There are two decreasing phases in this pattern that use the same price calculations,
// so we are going to make an embeddable type for them.
type smallSpikeDecreasingBase struct {
	phaseCoreAuto
}

func (phase *smallSpikeDecreasingBase) PotentialPeriod(
	period PricePeriod, phasePeriod int,
) *PotentialPricePeriod {
	minFactor, maxFactor := 0.4, 0.9

	for i := 0; i < phasePeriod; i++ {
		minFactor -= 0.03
		minFactor -= 0.02

		maxFactor -= 0.03
	}

	minPrice := RoundBells(float64(phase.ticker.PurchasePrice) * minFactor)
	maxPrice := RoundBells(float64(phase.ticker.PurchasePrice) * maxFactor)

	return &PotentialPricePeriod{
		prices: prices{
			min: minPrice,
			max: maxPrice,
		},
		PricePeriod: period,
	}
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

func (phase *smallSpikeDecreasing1) Duplicate() PatternPhase {
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

func (phase *smallSpikeIncreasing) PotentialPeriod(
	period PricePeriod, phasePeriod int,
) *PotentialPricePeriod {
	// We need to implement this from scratch for this phase, because 2 of the price
	// periods subtract one bell from the final price, which happens during no other
	// phase of the game, so we cannot use our base class.
	//
	// Turnip prophet gets a minimum value of 139 bells for a 100 bell purchase price
	// during this spike, but look at the source code here:
	//
	// https://gist.github.com/
	// Treeki/85be14d297c80c8b3c0a76375743325b#file-turnipprices-cpp-L329
	//
	// I am almost certain the 140% value is the lowest the price can go on this day,
	// and that turnip prophet is wrong. I am going to leave this as is for now and
	// revisit as needed later
	minFactor, maxFactor := 1.4, 2.0
	var priceAdjustment int

	switch {
	case phasePeriod == 0 || phasePeriod == 1:
		// Periods 1 and 2 are random between 90% and and 140% so alter the base rates.
		minFactor, maxFactor = 0.9, 1.4
	case phasePeriod == 2 || phasePeriod == 4:
		// For period 3 and 5, we subtract 1 from the total after doing our calculation.
		priceAdjustment = -1
	case phasePeriod > 4:
		panic("steady increase only has 5 sub periods")
	}

	minPrice := RoundBells(float64(phase.ticker.PurchasePrice)*minFactor) +
		priceAdjustment

	maxPrice := RoundBells(float64(phase.ticker.PurchasePrice)*maxFactor) +
		priceAdjustment

	return &PotentialPricePeriod{
		prices: prices{
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

func (phase *smallSpikeDecreasing2) Duplicate() PatternPhase {
	return &smallSpikeDecreasing2{
		smallSpikeDecreasingBase{
			phase.smallSpikeDecreasingBase.phaseCoreAuto,
		},
	}
}

// Generates a new set of fluctuating phases to branch possible weeks off of.
func smallSpikeProgression(ticker *PriceTicker) []PatternPhase {
	phases := []PatternPhase{
		new(smallSpikeDecreasing1),
		new(smallSpikeIncreasing),
		new(smallSpikeDecreasing2),
	}

	for _, thisPhase := range phases {
		thisPhase.SetTicker(ticker)
	}

	return phases
}
