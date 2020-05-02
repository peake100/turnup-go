package errs

import "errors"

var ErrPeriodOutOfIndex = errors.New("price phases must be between 0 and 11")

var ErrUnknownBaseChanceInvalid = errors.New(
	"'UNKNOWN' is not an in-game pattern and does not have a base chance",
)

var ErrUnknownPhasesInvalid = errors.New(
	"'UNKNOWN' is not an in-game pattern and does not have a phase progression'",
)

var ErrPhaseLengthFinalized = errors.New(
	"trying to fetch possible lengths on a finalized price pattern phase",
)

var ErrBadPatternIndex = errors.New("pattern index value must be 0-4")

var ErrPatternStringValue = errors.New("could not parse pattern from string")

var ErrNoSundayPricePeriod = errors.New("there are no price periods on sunday")
