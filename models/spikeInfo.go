package models

// Interface defining a potential object that has a hasSpikeAny
type HasSpike interface {
	// Whether the object has the potential for a Big Spike pattern
	HasSpikeBig() bool
	// Whether the object has the potential for a Small Spike pattern
	HasSpikeSmall() bool
	// Whether the object has the potential for a any spike pattern
	HasSpikeAny() bool
}

// Interface defining a potential object that has a hasSpikeAny range
type HasSpikeRange interface {
	HasSpike

	// The first price period a big hasSpikeAny could occur.
	SpikeBigStart() PricePeriod
	// The last price period a big hasSpikeAny could occur (inclusive).
	SpikeBigEnd() PricePeriod

	// The first price period a small hasSpikeAny could occur.
	SpikeSmallStart() PricePeriod
	// The last price period a small hasSpikeAny could occur (inclusive).
	SpikeSmallEnd() PricePeriod

	// The first price period any hasSpikeAny could occur.
	SpikeAnyStart() PricePeriod
	// The last price period any hasSpikeAny could occur (inclusive).
	SpikeAnyEnd() PricePeriod
}

// Implementation of HasSpikeAny
type Spikes struct {
	// Whether this is a big or small hasSpikeAny.
	hasSpikeAny bool
	// Whether this is a big hasSpikeAny. If small hasSpikeAny, is false.
	hasSpikeBig bool
	// Whether this is a big hasSpikeAny. If small hasSpikeAny, is false.
	hasSpikeSmall bool
}

// Whether the object has the potential for a Big Spike pattern
func (spike *Spikes) HasSpikeAny() bool {
	return spike.hasSpikeAny
}

// Whether the object has the potential for a Big Spike pattern
func (spike *Spikes) HasSpikeBig() bool {
	return spike.hasSpikeBig
}

// Whether the object has the potential for a Small Spike pattern
func (spike *Spikes) HasSpikeSmall() bool {
	return spike.hasSpikeSmall
}

// Implementation of HasSpikeRange
type SpikeRange struct {
	Spikes
	spikeAnyStart PricePeriod
	spikeAnyEnd   PricePeriod

	spikeBigStart PricePeriod
	spikeBigEnd   PricePeriod

	spikeSmallStart PricePeriod
	spikeSmallEnd   PricePeriod
}

// The first price period any spike pattern could occur.
func (spike *SpikeRange) SpikeAnyStart() PricePeriod {
	return spike.spikeAnyStart
}

// The last price period any spike pattern could occur.
func (spike *SpikeRange) SpikeAnyEnd() PricePeriod {
	return spike.spikeAnyEnd
}

// The first price period a big spike could occur.
func (spike *SpikeRange) SpikeBigStart() PricePeriod {
	return spike.spikeBigStart
}

// The last price period a big spike could occur.
func (spike *SpikeRange) SpikeBigEnd() PricePeriod {
	return spike.spikeBigEnd
}

// The first price period a small spike could occur.
func (spike *SpikeRange) SpikeSmallStart() PricePeriod {
	return spike.spikeSmallStart
}

// The last price period a small spike could occur.
func (spike *SpikeRange) SpikeSmallEnd() PricePeriod {
	return spike.spikeSmallEnd
}

func (spike *SpikeRange) updateSpikeFromPeriodAny(period PricePeriod, info HasSpike) {
	if !info.HasSpikeAny() {
		return
	}

	if !spike.hasSpikeAny || period < spike.spikeAnyStart {
		spike.spikeAnyStart = period
	}
	if period > spike.spikeAnyEnd {
		spike.spikeAnyEnd = period
	}
	spike.hasSpikeAny = true
}

func (spike *SpikeRange) updateSpikeFromPeriodSmall(period PricePeriod, info HasSpike) {
	if !info.HasSpikeSmall() {
		return
	}

	if !spike.hasSpikeSmall || period < spike.spikeSmallStart {
		spike.spikeSmallStart = period
	}

	if period > spike.spikeSmallEnd {
		spike.spikeSmallEnd = period
	}
	spike.hasSpikeSmall = true
}

func (spike *SpikeRange) updateSpikeFromPeriodBig(period PricePeriod, info HasSpike) {
	if !info.HasSpikeBig() {
		return
	}

	if !spike.hasSpikeBig || period < spike.spikeBigStart {
		spike.spikeBigStart = period
	}

	if period > spike.spikeBigEnd {
		spike.spikeBigEnd = period
	}
	spike.hasSpikeBig = true
}

func (spike *SpikeRange) updateSpikeFromPeriod(period PricePeriod, info HasSpike) {
	spike.updateSpikeFromPeriodAny(period, info)
	spike.updateSpikeFromPeriodSmall(period, info)
	spike.updateSpikeFromPeriodBig(period, info)
}

func (spike *SpikeRange) updateSpikeFromRangeSmall(info HasSpikeRange) {
	if !info.HasSpikeSmall() {
		return
	}

	if !spike.hasSpikeSmall || info.SpikeSmallStart() < spike.spikeSmallStart {
		spike.spikeSmallStart = info.SpikeSmallStart()
	}
	if info.SpikeSmallEnd() > spike.spikeSmallEnd {
		spike.spikeSmallEnd = info.SpikeSmallEnd()
	}

	spike.hasSpikeSmall = true
}

func (spike *SpikeRange) updateSpikeFromRangeBig(info HasSpikeRange) {
	if !info.HasSpikeBig() {
		return
	}

	if !spike.hasSpikeBig || info.SpikeBigStart() < spike.spikeBigStart {
		spike.spikeBigStart = info.SpikeBigStart()
	}
	if info.SpikeBigEnd() > spike.spikeBigEnd {
		spike.spikeBigEnd = info.SpikeBigEnd()
	}

	spike.hasSpikeBig = true
}

func (spike *SpikeRange) updateSpikeFromRangeAny(info HasSpikeRange) {
	if !info.HasSpikeAny() {
		return
	}

	if info.HasSpikeAny() {
		if !spike.hasSpikeAny || info.SpikeAnyStart() < spike.spikeAnyStart {
			spike.spikeAnyStart = info.SpikeAnyStart()
		}
		if info.SpikeAnyEnd() > spike.spikeAnyEnd {
			spike.spikeAnyEnd = info.SpikeAnyEnd()
		}

		spike.hasSpikeAny = true
	}
}

// Update From Range
func (spike *SpikeRange) updateSpikeFromRange(info HasSpikeRange) {
	spike.updateSpikeFromRangeSmall(info)
	spike.updateSpikeFromRangeBig(info)
	spike.updateSpikeFromRangeAny(info)
}
