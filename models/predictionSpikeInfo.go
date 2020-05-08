package models

import (
	"github.com/peake100/turnup-go/models/timeofday"
	"github.com/peake100/turnup-go/values"
	"time"
)

type SpikeChanceBreakdown [values.PricePeriodCount]float64

// Return the spike chance for a given Weekday + time of day
func (spikes *SpikeChanceBreakdown) ForDay(
	weekday time.Weekday, tod timeofday.ToD,
) (chance float64, err error) {
	pricePeriod, err := PricePeriodFromDay(weekday, tod)
	if err != nil {
		return 0, err
	}

	return spikes[pricePeriod], nil
}

// Return the spike chance for a given time. The ticker does not contain any information
// about dates, so it is assumed that the time passed in to spikeTime is for the week
// that the density describes.
func (spikes *SpikeChanceBreakdown) ForTime(
	spikeTime time.Time,
) (chance float64, err error) {
	pricePeriod, err := PricePeriodFromTime(spikeTime)
	if err != nil {
		return 0, err
	}

	return spikes[pricePeriod], nil
}

// A probability heat-map of when a price spike might occur.
type SpikeChances struct {
	SpikeRange
	// The probability distribution of a small spike by day.
	SmallBreakdown SpikeChanceBreakdown
	// The overall chance of a small spike.
	SmallChance float64

	// The probability distribution of a big spike by day
	BigBreakdown SpikeChanceBreakdown
	// The overall chance of a big spike.
	BigChance float64

	// The probability distribution of any spike by day
	AnyBreakdown SpikeChanceBreakdown
	// The overall chance of any spike.
	AnyChance float64
}

func (density *SpikeChances) updateSpikePeriod(
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
		density.SmallBreakdown[period] += weekChance
		containsSpike = true
	}

	// Add chance to big density if this is a big spike.
	if hasBig && period >= bigStart && period <= bigEnd {
		density.BigBreakdown[period] += weekChance
		containsSpike = true
	}

	if containsSpike {
		// Add to total density for both
		density.AnyBreakdown[period] += weekChance
	}
}

// We will updatePrices the density from the potential weeks.
func (density *SpikeChances) UpdateSpikeDensity(
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
