package errs

import "errors"

var ErrUnknownBaseChanceInvalid = errors.New(
	"'UNKNOWN' is not an in-game pattern and does not have a base chance",
)

var ErrUnknownPhasesInvalid = errors.New(
	"'UNKNOWN' is not an in-game pattern and does not have a phase progression'",
)

var ErrBadPatternIndex = errors.New("pattern index value must be 0-4")

var ErrPatternStringValue = errors.New("could not parse pattern from string")

var ErrNoSundayPricePeriod = errors.New("there are no price periods on sunday")

var ErrImpossibleTickerPrices = errors.New(
	"could not generate possibilities because ticker prices are impossible",
)
