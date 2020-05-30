package models

import "math"

// Generates heat for a possible near-future spike.
func heatSpikePeriod(
	period PricePeriod,
	breakdown *SpikeChanceBreakdown,
	periodMultiplier float64,
) (periodHeat float64) {
	// If the period is beyond the end of the week, we add 0.
	if int(period) < len(breakdown) {
		periodHeat = breakdown[period] * periodMultiplier
	}
	return periodHeat
}

// Generate a heat multiplier based on the likelihood of a spike happening in the
// next 3 periods starting with the current period.
func heatSpikeMultiplier(
	currentPeriod PricePeriod,
	breakdown *SpikeChanceBreakdown,
) (spikeMultiplier float64) {
	nextPeriod := currentPeriod + 1
	twoPeriods := currentPeriod + 2

	// We need this to always increase the score, so start it at 1.
	spikeMultiplier = 1

	// Give bonuses for possible spikes spikes in the next 3 periods. We want a current
	// spike with a below average roll to mostly out-shine a 100% possible potential
	// spike the next period with a higher average. With this current setup, a pattern
	// with a spike 20% below average will generate equal heat as a possible spike
	// with the same average and 100% possibility of happening the same day.
	spikeMultiplier += heatSpikePeriod(currentPeriod, breakdown, 0.4)
	spikeMultiplier += heatSpikePeriod(nextPeriod, breakdown, 0.2)
	spikeMultiplier += heatSpikePeriod(twoPeriods, breakdown, 0.2)
	return spikeMultiplier
}

// Calculate the investment heat for this island.
func (predictor *Predictor) CalcHeat() {
	prediction := predictor.result

	var heatFloat float64
	currentPeriod := predictor.Ticker.CurrentPeriod

	for _, pattern := range prediction.Patterns {
		// The base heat for each pattern will be the avg of the max and guaranteed
		// price times the chance of the pattern
		priceAverage := pattern.Future.MaxPrice() + pattern.Future.GuaranteedPrice()
		baseHeat := float64(priceAverage) / 2 * pattern.chance
		switch pattern.Pattern {
		case BIGSPIKE:
			breakdown := prediction.Spikes.Big().Breakdown()
			baseHeat *= heatSpikeMultiplier(currentPeriod, breakdown)
		case SMALLSPIKE:
			breakdown := prediction.Spikes.Small().Breakdown()
			baseHeat *= heatSpikeMultiplier(currentPeriod, breakdown)
		}
		heatFloat += baseHeat
	}

	prediction.Heat = int(math.Round(heatFloat))
}
