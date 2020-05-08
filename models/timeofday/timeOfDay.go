package timeofday

// Each non-sunday day in animal crossing is divided into two periods where the price is
// set, we are going to call that the time-of-day. The first period will be AM, and the
// second period PM. This is intended to be used as an enum type.
type ToD string

// Returns the phase offset value for the time of day. AM = 0, PM = 1
func (tod ToD) PhaseOffset() int {
	if tod == AM {
		return 0
	}
	return 1
}

const (
	// First price period of a day
	AM ToD = "AM"

	// Second price period of a day
	PM ToD = "PM"
)
