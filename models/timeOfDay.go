package models

// Each non-sunday day in animal crossing is divided into two periods where the price is
// set, we are going to call that the time-of-day. The first period will be AM, and the
// second period PM.

type ToD string

// Returns the phase offset value for the time of day. AM = 0, PM = 1
func (tod ToD) PhaseOffset() int {
	if tod == AM {
		return 0
	} else {
		return 1
	}
}

const (
	AM ToD = "AM"
	PM ToD = "PM"
)
