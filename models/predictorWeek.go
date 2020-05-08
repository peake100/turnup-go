package models

// Handles doing predictions for a potential phase permutation of a week once the phase
// pattern is finalized
type weekPredictor struct {
	Ticker        *PriceTicker
	Pattern       PricePattern
	PatternPhases []PatternPhase

	// Value cache
	// The probability weight of the price pattern given last week's pattern
	patternWeight           float64
	patternPermutationCount int

	// Variables
	// The total bin width for this week
	binWidth float64
	// Set to true if there are any known prices in the ticker
	pricesKnown bool

	result *PotentialWeek
}

func (predictor *weekPredictor) increaseBinWidth(amount float64) {
	predictor.binWidth += amount
}

// Come up with a score for how likely this period is to match the ticker. We call this
// score the bin "width"
func (predictor *weekPredictor) addPeriodBinWidth(
	pricePeriod PricePeriod,
	knownPrice int,
) {
	// If this price is unknown (0) then we can't make probability estimates with it.
	if knownPrice == 0 {
		return
	}

	// Otherwise remember that we know a price
	predictor.pricesKnown = true

	// Get the min and max prices for this period
	prices := predictor.result.Prices[pricePeriod]

	// Get the number of possible bell values (how many sides on this
	// dice?). We need to add one since this is an inclusive range
	periodRange := prices.MaxPrice() - prices.MinPrice() + 1

	// Now compute the likelihood of any particular price in this bracket
	// occurring divided by the total number of prices. For many combinations
	// the minimum and maximum prices are far less likely to occur because of how
	// the price math is implemented. We divide by the period range to get the
	// likelihood that this price would occur in this range relative to other
	// ranges.
	priceChance := prices.PriceChance(knownPrice)
	periodWidth := 0.0
	if priceChance != 0.0 {
		periodWidth = priceChance / float64(periodRange)
	}

	// Weight it by the likelihood of this pattern occurring in the first
	// place
	periodWidth *= predictor.patternWeight

	// Add it to the total likelihood of this week permutation happening
	predictor.increaseBinWidth(periodWidth)
}

func (predictor *weekPredictor) buildWeek() {
	ticker := predictor.Ticker
	result := predictor.result

	// The current week's price period
	var pricePeriod PricePeriod
	// The current sub period of the phase
	var phasePeriod int

	// Loop through each phase of the pattern
	for _, thisPhase := range predictor.PatternPhases {

		// Loop through the number of periods in this phase's length.
		for phasePeriod = 0; phasePeriod < thisPhase.Length(); phasePeriod++ {
			// Get the projected price for this price period, according to the phase.
			potentialPeriod := thisPhase.PotentialPeriod(pricePeriod, phasePeriod)

			// If this is not a valid price, we set the result to nil and stop making
			// predictions
			knownPrice := ticker.Prices[pricePeriod]
			if !potentialPeriod.IsValidPrice(knownPrice) {
				predictor.result = nil
				return
			}

			result.Prices = append(result.Prices, potentialPeriod)

			// We want to find the highest minimum for this potential week and use that
			// as the week's guaranteed minimum
			result.updatePriceRangeFromPrices(potentialPeriod, pricePeriod)
			result.Spikes.updateSpikeFromPeriod(
				potentialPeriod.PricePeriod, potentialPeriod.Spikes,
			)

			// Now get the probability width that this week will happen
			predictor.addPeriodBinWidth(pricePeriod, knownPrice)

			// Increment the overall price period
			pricePeriod++
		}
	}
}

func (predictor *weekPredictor) finalizeWidth() {
	// If we had known prices, then we have a more informed bin width, and can return.
	// We check this flag rather than for a bin width of 0, as patterns that COULD
	// happen, but are vanishingly unlikely will have an effective width of 0.s
	if !predictor.pricesKnown {
		predictor.binWidth = predictor.patternWeight
	}

	// Now weight each week by the number of possible weeks for this pattern. As we
	// knock out possible phase combinations for a pattern, the likelihood of this
	// pattern goes down.
	predictor.binWidth /= float64(predictor.patternPermutationCount)

	// Use this bin width as our chance for now.
	predictor.result.chance = predictor.binWidth
}

func (predictor *weekPredictor) setup() {
	predictor.result = &PotentialWeek{
		Analysis: new(Analysis),
		Spikes:   &SpikeRangeAll{
			big:   new(SpikeRange),
			small: new(SpikeRange),
			any:   new(SpikeRange),
		},
	}
	predictor.patternWeight = predictor.Pattern.BaseChance(
		predictor.Ticker.PreviousPattern,
	)
	predictor.patternPermutationCount = predictor.Pattern.PermutationCount()
}

func (predictor *weekPredictor) Predict() (
	potentialWeek *PotentialWeek, binWidth float64,
) {
	predictor.setup()
	predictor.buildWeek()
	if predictor.result != nil {
		predictor.finalizeWidth()
	}
	return predictor.result, predictor.binWidth
}
