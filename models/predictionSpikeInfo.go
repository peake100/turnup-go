package models

import "github.com/peake100/turnup-go/values"

// A probability heat-map of when a price spike might occur.
type SpikePeriodDensity struct {
	SpikeRange
	// The probability distribution of a spike by day
	SmallDensity [values.PricePeriodCount]float64
	SmallChance  float64

	BigDensity [values.PricePeriodCount]float64
	BigChance  float64

	AnyDensity [values.PricePeriodCount]float64
	AnyChance  float64
}

func (density *SpikePeriodDensity) updateSpikePeriod(
	update HasSpikeRange,
	period PricePeriod,
	weekChance float64,
) {

	hasSmall := update.HasSpikeSmall()
	smallStart := update.SpikeSmallStart()
	smallEnd := update.SpikeSmallStart()

	hasBig := update.HasSpikeBig()
	bigStart := update.SpikeBigStart()
	bigEnd := update.SpikeBigEnd()

	// Add chance to small density if this is a small spike.
	containsSpike := false
	if hasSmall && period >= smallStart && period <= smallEnd {
		density.SmallDensity[period] += weekChance
		containsSpike = true
	}

	// Add chance to big density if this is a big spike.
	if hasBig && period >= bigStart && period <= bigEnd {
		density.BigDensity[period] += weekChance
		containsSpike = true
	}

	if containsSpike {
		// Add to total density for both
		density.AnyDensity[period] += weekChance
	}
}

// We will updatePrices the density from the potential weeks.
func (density *SpikePeriodDensity) UpdateSpikeDensity(
	updateWeek *PotentialWeek,
) {
	// The idea behind this heatmap is simple: take the bin width of a given potential
	// week, and add it to a running tally of each price period a spike occurs on that
	// week for. We need to run this AFTER all of the chances are normalized for every
	// pattern, as the total likelihood for any spike may be under 1.

	// If there is no spike, abort.
	update := updateWeek.Spikes
	weekChance := updateWeek.Chance()

	if !update.HasSpikeAny() {
		return
	}

	start := update.SpikeAnyStart()
	end := update.SpikeAnyEnd()

	for period := start; period <= end; period++ {
		density.updateSpikePeriod(update, period, weekChance)
	}
}
