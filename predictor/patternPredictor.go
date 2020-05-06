package predictor

import (
	"github.com/peake100/turnup-go/models"
)

type patternPredictor struct {
	// Info
	Ticker *models.PriceTicker
	Pattern models.PricePattern

	result *models.PotentialPattern
}


// Makes a duplicate of the current phase pattern to be a new possibility
func (predictor *patternPredictor) duplicatePhasePattern(
	patternPhases []models.PatternPhase,
) []models.PatternPhase {

	dupedPhases := make([]models.PatternPhase, len(patternPhases))
	for i, phase := range patternPhases {
		var branchPhase models.PatternPhase

		// If the phase is finalized, we don't need to copy it. It's values will not
		// be changing.
		if phase.IsFinal() {
			branchPhase = phase
		} else {
			branchPhase = phase.Duplicate()
		}

		dupedPhases[i] = branchPhase
	}
	return dupedPhases
}

// Once a particular phase pattern is fully computed, this function build the potential
// prices for each price period. Returns nil if this pattern is impossible given the
// ticker's real-world values
func (predictor *patternPredictor) potentialWeekFromPhasePattern(
	patternPhases []models.PatternPhase, ticker *models.PriceTicker,
) *models.PotentialWeek {
	result := new(models.PotentialWeek)

	// The current week's price period
	var pricePeriod models.PricePeriod
	// The current sub period of the phase
	var phasePeriod int

	// Loop through each phase of the pattern
	for _, thisPhase := range patternPhases {

		// Loop through the number of periods in this phase's length.
		for phasePeriod = 0; phasePeriod < thisPhase.Length(); phasePeriod++ {
			// Get the projected price for this price period, according to the phase.
			potentialPeriod := thisPhase.PotentialPeriod(pricePeriod, phasePeriod)

			if !potentialPeriod.IsValidPrice(ticker.Prices[pricePeriod]) {
				return nil
			}

			result.PricePeriods = append(result.PricePeriods, potentialPeriod)

			// We want to find the highest minimum for this potential week and use that
			// as the week's guaranteed minimum
			result.Analysis().Update(potentialPeriod, true)
			result.UpdateSpikeFromPeriod(potentialPeriod.PricePeriod, potentialPeriod)

			// Increment the overall price period
			pricePeriod++
		}

	}

	return result
}

// Launches the goroutine for a new permutation of a price pattern.
func (predictor *patternPredictor) computeBranch(
	thisPossibleLength int,
	possibilityIndex int,
	phaseIndex int,
	allPossibleLengths []int,
	patternPhases []models.PatternPhase,
) {
	var newBranch []models.PatternPhase
	if possibilityIndex < len(allPossibleLengths)-1 {
		// duplicate our current pattern so we can set the possible length
		// for this phase
		newBranch = predictor.duplicatePhasePattern(patternPhases)
	} else {
		// If this is the last possible length, we can just re-use the current
		// branch rather than making a new duplicate and throwing away our
		// current
		newBranch = patternPhases
	}

	// set the branch phases' length to this possibility
	newBranch[phaseIndex].SetLength(thisPossibleLength)
	// Tell the work group we are adding a new branch we need to wait for
	// Launch the branch
	predictor.branchPhases(newBranch)
}

// Takes an array of pattern phases and recursively works through all un-computed
// possible phase length patterns.
func (predictor *patternPredictor) branchPhases(
	patternPhases []models.PatternPhase,
) {
	ticker := predictor.Ticker

	// To figure out the pattern for a week, we need to find all the possible lengths
	// for each phase, then make a copy of the phase pattern with that possibility
	// set to be re-iterated over in a new routine. We continue until all possibilities
	// in all goroutines have reported they are finalized.
	//
	// There is no variance in the price pattern of each phase, only in how long the
	// phase lasts. So if we have all possible combinations of phase lengths, then we
	// have all possible price patterns.
	for phaseIndex, phase := range patternPhases {

		// If this phase has it's final length set, we can continue to the next phase.
		if phase.IsFinal() {
			continue
		}

		// Get all the possible lengths for this phase,
		possibleLengths := phase.PossibleLengths(patternPhases)
		// If the possibilities are nil, then this phase is waiting for more
		// information, and we should continue to the next phase
		if possibleLengths == nil {
			continue
		}

		// Otherwise we need to create a new possible pattern branch for each phase
		// length and branch off of it.
		for i, phaseLength := range possibleLengths {
			predictor.computeBranch(
				phaseLength,
				i,
				phaseIndex,
				possibleLengths,
				patternPhases,
			)
		}
	}

	// If we make it all the way through than we have hit a fully formed possible phase
	// pattern! Now we can compute the possible prices and return them as the result
	potentialWeek := predictor.potentialWeekFromPhasePattern(patternPhases, ticker)
	// If we get nil back, then this week cannot be the result of this ticker
	if potentialWeek == nil {
		return
	}

	result := predictor.result
	// Otherwise, add the result
	result.PotentialWeeks = append(result.PotentialWeeks, potentialWeek)
	result.Analysis().Update(potentialWeek.Analysis(), false)
	result.UpdateSpikeFromRange(potentialWeek)
}

// Calculate all the possible phase permutations for a given price pattern.
func (predictor *patternPredictor) Predict() *models.PotentialPattern {
	result := &models.PotentialPattern{
		Pattern:        predictor.Pattern,
	}
	predictor.result = result

	ticker := predictor.Ticker

	// Get the base phase progression of this pattern
	patternPhases := predictor.Pattern.PhaseProgression(ticker)
	predictor.branchPhases(patternPhases)

	return result
}
