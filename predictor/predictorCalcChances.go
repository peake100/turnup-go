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

func (predictor *Predictor) calculatePotentialWeekWidth(
	week *models.PotentialWeek,
	pattern models.PricePattern,
	knownPricePeriods []models.PricePeriod,
	ticker *models.PriceTicker,
) float64 {
	patternWeight := pattern.BaseChance(ticker.PreviousPattern)
	patternPermutationCount := pattern.PermutationCount()

	var weekWidth float64

	for _, pricePeriod := range knownPricePeriods {
		// Get the min and max prices for this period
		prices := week.PricePeriods[pricePeriod]

		// Get the number of possible bell values (how many sides on this
		// dice?). We need to add one since this is an inclusive range
		periodRange := prices.MaxPrice() - prices.MinPrice() + 1

		// Now compute the likelihood of any particular price in this bracket
		// occurring divided by the total number of prices. For many combinations
		// the minimum and maximum prices are far less likely to occur because of how
		// the price math is implemented. We divide by the period range to get the
		// likelihood that this price would occur in this range relative to other
		// ranges.
		knownPrice := ticker.Prices[pricePeriod]
		priceChance := prices.PriceChance(knownPrice)
		periodWidth := 0.0
		if priceChance != 0.0 {
			periodWidth = priceChance / float64(periodRange)
		}

		// Weight it by the likelihood of this pattern occurring in the first
		// place
		periodWidth *= patternWeight

		// Add it to the total likelihood of this week permutation happening
		weekWidth += periodWidth
	}

	// if there is no known price data, we will just use the pattern chance for
	// this week divided by the number of possible patterns. This gives less
	// weight to patterns who have had permutations eliminated vs those who
	// have not.
	if len(knownPricePeriods) == 0 {
		weekWidth = patternWeight
	}

	// Now weight each week by the number of possible weeks. As we knock out possible
	// phase combinations for a pattern, the likelihood of this pattern goes down.
	weekWidth /= float64(patternPermutationCount)

	return weekWidth
}

func (predictor *Predictor) calculatePatternWidth(
	potentialPattern *models.PotentialPattern,
	knownPricePeriods []models.PricePeriod,
	ticker *models.PriceTicker,
) float64 {

	var patternWidth float64
	for _, week := range potentialPattern.PotentialWeeks {
		weekChance := predictor.calculatePotentialWeekWidth(
			week,
			potentialPattern.Pattern,
			knownPricePeriods,
			ticker,
		)

		// Add this week's chance to the pattern chance
		patternWidth += weekChance
		// Save it as an interim step
		week.Analysis().Chance = weekChance
	}

	return patternWidth
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

	// First lets build a list of the price periods we have data for
	var knownPricePeriods []models.PricePeriod

	for pricePeriodIndex, price := range ticker.Prices {
		pricePeriod := models.PricePeriod(pricePeriodIndex)
		// If the price is 0, then it is unknown
		if price == 0 {
			continue
		}

		knownPricePeriods = append(knownPricePeriods, pricePeriod)
	}

	// We need to store the total chance width of this price space (1/6 + 1/20 in the
	// example above).
	var totalWidth float64

	// Now let's loop through out patterns and start assigning chances.
	for _, potentialPattern := range prediction.Patterns {
		patternWidth := predictor.calculatePatternWidth(
			potentialPattern, knownPricePeriods, ticker,
		)
		// Add the pattern's chance to the total chance units
		totalWidth += patternWidth
		// Remember it as an interim step
		potentialPattern.Analysis().Chance = patternWidth
	}

	// There's a potential that total width is going to be zero if we're in a super rare
	// price pattern with a price that only has a 1-in-a-billion chance of happening.
	// If that's the case we want to use the number of existing weeks for a pattern
	// divided by the number of possible weeks.
	if totalWidth == 0 {
		totalWidth = predictor.fallBackToPatternCount(prediction, ticker)
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
