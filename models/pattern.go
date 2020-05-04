package models

import (
	"github.com/peake100/turnup-go/errs"
	"strings"
)

type PricePattern int

func (pattern PricePattern) String() string {
	return [5]string{
		"FLUCTUATING",
		"BIG SPIKE",
		"DECREASING",
		"SMALL SPIKE",
		"UNKNOWN",
	}[pattern]
}

// Because we are including unknown, we need to allow for the 0 position of every
// row to be for if the previous week is unknown. We can actually make an informed
// decision of how likely each pattern is if the previous chance is unknown, since
// the overall aggregate of the likelihoods for a pattern are not uniform.
//
// Lets treat each percentage chance as a 'likelihood unit.' The Small spike pattern
// gets 35 units for being 35% likely after a Random pattern, 25 units for being 25%
// likely after a Large spike, etc.
//
// We end up with the following scores:
//
//	  Random: 140
//    Large Spike: 105
//    Decreasing: 55
//    Small Spike: 100
//
// Now we can divide each by 400 (the total units) to see how likely each pattern is
// *on average*.
//
//	  Random: 35%
//    Large Spike: 26.25%
//    Decreasing: 13.75%
//    Small Spike: 25%
//
// We end up with a probability matrix that looks like this, where the left column is
// last week's pattern and the top row is the likelihood of the pattern for this week.
// ==================================================================
//             | Fluctuating |   Big Spike | Decreasing | Small Spike
// ------------------------------------------------------------------
// Fluctuating | 20.00%      |      30.00% |     15.00% |      35.00%
// ------------------------------------------------------------------
// Big Spike   | 50.00%      |      05.00% |     20.00% |      25.00%
// ------------------------------------------------------------------
// Decreasing  | 25.00%      |      45.00% |     05.00% |      25.00%
// ------------------------------------------------------------------
// Small Spike | 45.00%      |      25.00% |     15.00% |      15.00%
// ------------------------------------------------------------------
// Unknown     | 35.00%      |      26.25% |     13.75% |      25.00%
// ------------------------------------------------------------------
//
// Lets turn this into a matrix on which we can use the pattern indexes to do a lookup
var initialChanceMatrix = [5][4]float64{
	// FLUCTUATING
	{0.20, 0.30, 0.15, 0.35},
	// BIG SPIKE
	{0.50, 0.05, 0.20, 0.25},
	// DECREASING
	{0.25, 0.45, 0.05, 0.25},
	// SMALL SPIKE
	{0.45, 0.25, 0.15, 0.15},
	// Unknown
	{0.35, 0.2625, 0.1375, 0.25},
}

// Returns a the chance of this pattern occurring based on the pattern from last week
func (pattern PricePattern) BaseChance(previous PricePattern) float64 {
	if pattern == UNKNOWN {
		panic(errs.ErrUnknownBaseChanceInvalid)
	}
	// Do the lookup
	return initialChanceMatrix[previous][pattern]
}

// Returns a new set of phase definitions that can be used to calculate the possible
// price values for a week.
func (pattern PricePattern) PhaseProgression(ticker *PriceTicker) []PatternPhase {
	switch {
	case pattern == FLUCTUATING:
		return fluctuatingProgression(ticker)
	case pattern == BIGSPIKE:
		return bigSpikeProgression(ticker)
	case pattern == DECREASING:
		return decreasingProgression(ticker)
	case pattern == SMALLSPIKE:
		return smallSpikeProgression(ticker)
	case pattern == UNKNOWN:
		panic(errs.ErrUnknownPhasesInvalid)
	default:
		panic(errs.ErrBadPatternIndex)
	}
}

// The total possible phase combinations for this pattern. Can be used to determine
// actual chance of this pattern once possibilities have been removed by a ticker.
func (pattern PricePattern) PermutationCount() int {
	return [4]int{56, 7, 1, 8}[pattern]
}

const (
	FLUCTUATING PricePattern = 0
	BIGSPIKE    PricePattern = 1
	DECREASING  PricePattern = 2
	SMALLSPIKE  PricePattern = 3
	UNKNOWN     PricePattern = 4
)

// An array of the possible patterns in index order
var PATTERNS = [5]PricePattern{FLUCTUATING, BIGSPIKE, DECREASING, SMALLSPIKE, UNKNOWN}

// All the valid patterns in the game. Unknown is not a valid pattern, and only
// one we need include because of incomplete game information
// var PATTERNSGAME = [4]models.PricePattern{FLUCTUATING, DECREASING, SMALLSPIKE, BIGSPIKE}
var PATTERNSGAME = [4]PricePattern{FLUCTUATING, BIGSPIKE, DECREASING, SMALLSPIKE}

// Returns a pattern from a string: The following values are valid. The four names are:
//
//		1. Fluctuating
//  	2. Big Spike
//  	3. Decreasing
//  	4. Small Spike
//  	5. Unknown
//
// Incoming values are upper-cased, and spaces are removed before evaluating, so for
// big spike, all of the following would be handled without error for Big Spike:
//
//		- BIGSPIKE
//		- bigspike
//		- BIG SPIKE
//		- Big Spike
//		- big spike
//		- etc.
func PatternFromString(value string) (PricePattern, error) {
	value = strings.ToUpper(value)
	value = strings.Replace(value, " ", "", -1)
	pattern, ok := map[string]PricePattern{
		"FLUCTUATING": FLUCTUATING,
		"BIGSPIKE":    BIGSPIKE,
		"DECREASING":  DECREASING,
		"SMALLSPIKE":  SMALLSPIKE,
		"UNKNOWN":     UNKNOWN,
	}[value]

	if !ok {
		return 5, errs.ErrPatternStringValue
	}

	return pattern, nil
}
