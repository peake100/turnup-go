package models

// This struct can be embedded with an implemented phase to complete the full phase
// implementation and get price period calculations for free
type patternPhaseAuto struct {
	phaseImplement
	// For compounding phases, we have to generate
	potentialPeriods []*PotentialPricePeriod

	// whether or not the purchase price is known
	purchasePriceKnown bool

	// Generator objects that calculate the next period's price on demand.
	pricePeriodGen    *phasePeriodGenerator
	pricePeriodGenMax *phasePeriodGenerator
}

func (phase *patternPhaseAuto) setup(period PricePeriod, subPeriod int) {
	phase.potentialPeriods = make([]*PotentialPricePeriod, phase.MaxLength())
	phase.purchasePriceKnown = phase.Ticker().PurchasePrice != 0

	phaseStartPeriod := period - PricePeriod(subPeriod)
	purchasePrice := phase.Ticker().PurchasePrice
	if !phase.purchasePriceKnown {
		purchasePrice = 90
	}

	phase.pricePeriodGen = &phasePeriodGenerator{
		Ticker:           phase.Ticker(),
		PurchasePrice:    purchasePrice,
		PhaseFull:        phase,
		PricePeriodStart: phaseStartPeriod,
	}
	phase.pricePeriodGen.Setup()

	// If the purchase price is unknown we need to get the possibilities twice, once
	// for the min price and once for the max price.
	if !phase.purchasePriceKnown {
		phase.pricePeriodGenMax = &phasePeriodGenerator{
			Ticker:           phase.Ticker(),
			PurchasePrice:    110,
			PhaseFull:        phase,
			PricePeriodStart: phaseStartPeriod,
		}
		phase.pricePeriodGenMax.Setup()
	}
}

func (phase *patternPhaseAuto) PotentialPeriod(
	period PricePeriod, subPeriod int,
) *PotentialPricePeriod {
	// Set up our cache if needed
	if phase.potentialPeriods == nil {
		phase.setup(period, subPeriod)
	}

	// See if this period is in our cache
	potentialPeriod := phase.potentialPeriods[subPeriod]
	if potentialPeriod != nil {
		// If it is, return it.
		return potentialPeriod
	}

	// If not, we are going to generate price periods through our price generator until
	// we get to the period we want.
	for i := phase.pricePeriodGen.LastCompletedSubPeriod + 1; i <= subPeriod; i++ {
		potentialPeriod = phase.pricePeriodGen.Next()

		// If the price is unknown get the next max possibility and adjust the max
		// values.
		if !phase.purchasePriceKnown {
			potentialPeriodMax := phase.pricePeriodGenMax.Next()
			potentialPeriod.maxPrice = potentialPeriodMax.maxPrice
			potentialPeriod.maxChance = potentialPeriodMax.maxChance
		}
		phase.potentialPeriods[i] = potentialPeriod
	}

	return potentialPeriod
}

func (phase *patternPhaseAuto) Duplicate() PatternPhase {
	return &patternPhaseAuto{
		phaseImplement: phase.phaseImplement.Duplicate(),
		// This cache will need to be created for each phase p
		potentialPeriods: nil,
	}
}
