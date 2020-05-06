package predictor

import (
	"github.com/peake100/turnup-go/models"
	"math"
)

type hasAnalysis interface {
	Analysis() *models.Analysis
}

// Converts the chance on an Analysis() method from chance width to absolute chance
// with a precision of 4 digits (XX.XX%).
func (predictor *Predictor) setChanceFromWidth(item hasAnalysis, totalWidth float64) {
	chance := item.Analysis().Chance / totalWidth
	// round to 4 digits (xx.xx%)
	chance = math.Round(chance*10000) / 10000
	item.Analysis().Chance = chance
}

// If we are in a price pattern which is VANISHINGLY unlikely, to the point that the
// effective chance is 0, we are going to base the likelihood of each pattern off it's
// base chance and the number of permutations we can eliminate.
func (predictor *Predictor) fallBackToPatternCount(
	prediction *models.Prediction, ticker *models.PriceTicker,
) (totalWidth float64) {
	for _, potentialPattern := range prediction.Patterns {
		potentialMatches := len(potentialPattern.PotentialWeeks)
		maxPermutations := potentialPattern.Pattern.PermutationCount()

		// Pattern chance is the number of active possibilities / the number of
		// eliminated possibilities, weighted by the base chance
		patternChance :=
			float64(potentialMatches) /
				float64(maxPermutations) *
				potentialPattern.Pattern.BaseChance(ticker.PreviousPattern)

		totalWidth += patternChance
		potentialPattern.Analysis().Chance = patternChance
	}

	return totalWidth
}

// Calculate the chances of each price pattern permutation once they have been
// calculated.
func (predictor *Predictor) calculateChances(
	ticker *models.PriceTicker, prediction *models.Prediction,
) {
	// We are going to calculate the likelihood that a bell price in the ticker came
	// from a given range by looping through the price periods we have data for and
	// examining the likelihood that the results came from the one possible phase combo
	// or another.
	//
	// Imagine we have two dice, a 6-sided die and an 20-sided die. We know that one of
	// these dies was rolled and the result was 5. Originally, there was a 50/50 chance
	// for what die was picked, but now we know the number was a 5, it becomes MORE
	// LIKELY the result came from the 6-sided die, since the chances of rolling a 5
	// on a six sided die are 1-in-6, while the chances for rolling a 5 on a 20-sided
	// die are 1-in-20.
	//
	// We can compute the chance this was a 6-sided dire by adding 1/20 to 1/6 and then
	// dividing 1/6 by the answer:
	//
	// 1/6 / (1/6 + 1/20) = 77% chance this was the result of a d6, which means a 23%
	// chance this was a d20.
	//
	// We are going to apply the same logic to bell price brackets. If we have two
	// possible patterns, and one has a bound  of 85-90 bells, and the other a bound of
	// 90 - 140 bells, then we know it is much more likely that we are in the pattern
	// with the bound of 85-90 bells when we have a price of 90 bells.
	//
	// We will weight these chances with the base chance for the pattern this week.
	//
	// These calculations are done during the generation of potential phase patterns and
	// price periods, allowing us to not need to loop through all the results
	// calculating them here.

	// There's a potential that total width is going to be zero if we're in a super rare
	// price pattern with a price that only has a 1-in-a-billion chance of happening.
	// If that's the case we want to use the number of existing weeks for a pattern
	// divided by the number of possible weeks.
	totalWidth := predictor.totalWidth
	if totalWidth == 0 {
		totalWidth = predictor.fallBackToPatternCount(prediction, ticker)
		predictor.totalWidth = totalWidth
	}

	// Now we can go through and figure out the final chance for each entry using our
	// total chance units
	for _, potentialPattern := range prediction.Patterns {
		// Otherwise refactor the individual chances by dividing their width by the
		// total width.
		predictor.setChanceFromWidth(potentialPattern, totalWidth)
		for _, week := range potentialPattern.PotentialWeeks {
			predictor.setChanceFromWidth(week, totalWidth)
		}
	}

	// And we're done! Phew!
}
