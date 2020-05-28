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

	// The closer the possible spike is, the more heat it generates.
	spikeMultiplier += heatSpikePeriod(currentPeriod, breakdown, 2)
	spikeMultiplier += heatSpikePeriod(nextPeriod, breakdown, 1)
	spikeMultiplier += heatSpikePeriod(twoPeriods, breakdown, 0.5)
	return spikeMultiplier
}

// Calculate the investment heat for this island.
func (predictor *Predictor) CalcHeat() {
	prediction := predictor.result

	var heatFloat float64
	currentPeriod := predictor.Ticker.CurrentPeriod

	for _, pattern := range prediction.Patterns {
		baseHeat := float64(pattern.Future.maxPrice-100) * pattern.chance
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
