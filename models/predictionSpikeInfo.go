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

type HasSpikeChance interface {
	HasSpikeRange
	Chance() float64
	Breakdown() *SpikeChanceBreakdown
}

type SpikeChance struct {
	SpikeRange
	chance    float64
	breakdown *SpikeChanceBreakdown
}

func (spike *SpikeChance) Chance() float64 {
	return spike.chance
}

func (spike *SpikeChance) Breakdown() *SpikeChanceBreakdown {
	return spike.breakdown
}

func (spike *SpikeChance) updatePeriodDensity(
	update HasSpikeRange,
	period PricePeriod,
	weekChance float64,
) {
	if update.Has() && period >= update.Start() && period <= update.End() {
		spike.breakdown[period] += weekChance
	}
}

// A probability heat-map of when a price spike might occur.
type SpikeChancesAll struct {
	small *SpikeChance
	big   *SpikeChance
	any   *SpikeChance
}

func (spikes *SpikeChancesAll) Big() HasSpikeChance {
	return spikes.big
}

func (spikes *SpikeChancesAll) Small() HasSpikeChance {
	return spikes.small
}

func (spikes *SpikeChancesAll) Any() HasSpikeChance {
	return spikes.any
}

// Converts from HasSpikeChancesAll to HasSpikeRangesAll
func (spikes *SpikeChancesAll) SpikeRangeAll() *SpikeRangeAll {
	// Extract the embedded types and rewrap them
	return &SpikeRangeAll{
		big:   &spikes.big.SpikeRange,
		small: &spikes.small.SpikeRange,
		any:   &spikes.any.SpikeRange,
	}
}

func (spikes *SpikeChancesAll) updateRanges(info *SpikeRangeAll) {
	spikes.any.updateFromRange(info.any)
	spikes.big.updateFromRange(info.big)
	spikes.small.updateFromRange(info.small)
}

// We will updatePrices the density from the potential weeks.
func (spikes *SpikeChancesAll) updateDensities(
	updateWeek *PotentialWeek,
) {
	// The idea behind this heatmap is simple: take the bin width of a given potential
	// week, and add it to a running tally of each price period a spike occurs on that
	// week for. We need to run this AFTER all of the chances are normalized for every
	// pattern, as the total likelihood for any spike may be under 1.

	// If there is no spike, abort.
	update := updateWeek.Spikes
	weekChance := updateWeek.Chance()

	if !update.any.Has() {
		return
	}

	start := update.any.Start()
	end := update.any.End()

	for period := start; period <= end; period++ {
		if update.any.has {
			spikes.any.updatePeriodDensity(update.any, period, weekChance)
		}
		if update.big.has {
			spikes.big.updatePeriodDensity(update.big, period, weekChance)
		}
		if update.small.has {
			spikes.small.updatePeriodDensity(update.small, period, weekChance)
		}
	}
}
