package models

// Interface defining a potential object that has a hasSpikeAny
type HasSpike interface {
	HasSpikeAny() bool
	HasSpikeBig() bool
	HasSpikeSmall() bool
}

// Interface defining a potential object that has a hasSpikeAny range
type HasSpikeRange interface {
	HasSpike

	// The first price period any hasSpikeAny could occur.
	SpikeAnyStart() PricePeriod
	// The last price period any hasSpikeAny could occur (inclusive).
	SpikeAnyEnd() PricePeriod

	// The first price period a big hasSpikeAny could occur.
	SpikeBigStart() PricePeriod
	// The last price period a big hasSpikeAny could occur (inclusive).
	SpikeBigEnd() PricePeriod

	// The first price period a small hasSpikeAny could occur.
	SpikeSmallStart() PricePeriod
	// The last price period a small hasSpikeAny could occur (inclusive).
	SpikeSmallEnd() PricePeriod
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

func (spike *Spikes) HasSpikeAny() bool {
	return spike.hasSpikeAny
}

func (spike *Spikes) HasSpikeBig() bool {
	return spike.hasSpikeBig
}

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

func (spike *SpikeRange) SpikeAnyStart() PricePeriod {
	return spike.spikeAnyStart
}

func (spike *SpikeRange) SpikeAnyEnd() PricePeriod {
	return spike.spikeAnyEnd
}

func (spike *SpikeRange) SpikeBigStart() PricePeriod {
	return spike.spikeBigStart
}

func (spike *SpikeRange) SpikeBigEnd() PricePeriod {
	return spike.spikeBigEnd
}

func (spike *SpikeRange) SpikeSmallStart() PricePeriod {
	return spike.spikeSmallStart
}

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

// UpdateSpikeFromPeriod
func (spike *SpikeRange) UpdateSpikeFromPeriod(period PricePeriod, info HasSpike) {
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
