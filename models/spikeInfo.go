package models

// Interface defining a potential object that has a spike of a given type
type HasSpike interface {
	// Whether the object has the potential for a Big Spike pattern
	Has() bool
}

// Interface defining a potential object that has a hasSpikeAny range
type HasSpikeRange interface {
	HasSpike

	// The first price period a big hasSpikeAny could occur.
	Start() PricePeriod
	// The last price period a big hasSpikeAny could occur (inclusive).
	End() PricePeriod
}

// Implementation of HasSpike
type Spike struct {
	// Whether this is a big or small hasSpikeAny.
	has bool
}

// Whether the object has the potential for a Big Spike pattern
func (spike *Spike) Has() bool {
	return spike.has
}

// Implementation of HasSpikeRange
type SpikeRange struct {
	Spike
	start PricePeriod
	end   PricePeriod
}

// The first price period any spike pattern could occur.
func (spike *SpikeRange) Start() PricePeriod {
	return spike.start
}

// The last price period any spike pattern could occur.
func (spike *SpikeRange) End() PricePeriod {
	return spike.end
}

func (spike *SpikeRange) updateSpikeFromPeriod(period PricePeriod, info HasSpike) {
	if !info.Has() {
		return
	}

	if !spike.Has() || period < spike.start {
		spike.start = period
	}
	if period > spike.end {
		spike.end = period
	}
	spike.has = true
}

func (spike *SpikeRange) updateFromRange(info HasSpikeRange) {
	if !info.Has() {
		return
	}

	if !spike.has || info.Start() < spike.start {
		spike.start = info.Start()
	}
	if info.End() > spike.end {
		spike.end = info.End()
	}

	spike.has = true
}

// Shared interface for SpikeHasAll, SpikeRangeAll, and SpikeChancesAll
type HasSpikeAll interface {
	Big() HasSpike
	Small() HasSpike
	Any() HasSpike
}

// Shared interface for SpikeRangeAll, and SpikeChancesAll
type HasSpikeRangeAll interface {
	Big() HasSpikeRange
	Small() HasSpikeRange
	Any() HasSpikeRange
}

type SpikeHasAll struct {
	big   *Spike
	small *Spike
	any   *Spike
}

func (spikes *SpikeHasAll) Big() HasSpike {
	return spikes.big
}

func (spikes *SpikeHasAll) Small() HasSpike {
	return spikes.small
}

func (spikes *SpikeHasAll) Any() HasSpike {
	return spikes.any
}

type SpikeRangeAll struct {
	big   *SpikeRange
	small *SpikeRange
	any   *SpikeRange
}

func (spike *SpikeRangeAll) Big() HasSpikeRange {
	return spike.big
}

func (spike *SpikeRangeAll) Small() HasSpikeRange {
	return spike.small
}

func (spike *SpikeRangeAll) Any() HasSpikeRange {
	return spike.any
}

func (spike *SpikeRangeAll) updateSpikeFromPeriod(period PricePeriod, info HasSpikeAll) {
	spike.any.updateSpikeFromPeriod(period, info.Any())
	spike.big.updateSpikeFromPeriod(period, info.Big())
	spike.small.updateSpikeFromPeriod(period, info.Small())
}

// Update From Range
func (spike *SpikeRangeAll) updateSpikeFromRange(info HasSpikeRangeAll) {
	spike.big.updateFromRange(info.Big())
	spike.small.updateFromRange(info.Small())
	spike.any.updateFromRange(info.Any())
}
