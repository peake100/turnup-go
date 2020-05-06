package predictor

import (
	"github.com/peake100/turnup-go/errs"
	"github.com/peake100/turnup-go/models"
	"github.com/peake100/turnup-go/models/patterns"
	"sync"
)

type Predictor struct {
	// The price ticker to use for this prediction
	Ticker *models.PriceTicker

	// The prediction result
	result *models.Prediction
}

// Makes a duplicate of the current phase pattern to be a new possibility
func (predictor *Predictor) duplicatePhasePattern(
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
func (predictor *Predictor) potentialWeekFromPhasePattern(
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

// Takes an array of pattern phases and recursively works through all un-computed
// possible phase length patterns.
func (predictor *Predictor) branchWeeks(
	ticker *models.PriceTicker,
	patternPhases []models.PatternPhase,
	weeksWorkSync *weekPredictionSync,
) {
	// Remove this mutations work counter off of the wait group.
	defer weeksWorkSync.WaitGroup.Done()

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
			predictor.launchPossibleLengthRoutine(
				ticker,
				phaseLength,
				i,
				phaseIndex,
				possibleLengths,
				patternPhases,
				weeksWorkSync,
			)
		}

		// Once we have started all the possible branches for this phase, we can return
		// and let them continue branching
		return
	}

	// If we make it all the way through than we have hit a fully formed possible phase
	// pattern! Now we can compute the possible prices and return them as the result
	potentialWeek := predictor.potentialWeekFromPhasePattern(patternPhases, ticker)
	// If we get nil back, then this week cannot be the result of this ticker
	if potentialWeek == nil {
		return
	}

	// Otherwise, report the week to the results channel
	weeksWorkSync.ResultChan <- potentialWeek
}

// Launches the goroutine for a new permutation of a price pattern.
func (predictor *Predictor) launchPossibleLengthRoutine(
	ticker *models.PriceTicker,
	thisPossibleLength int,
	possibilityIndex int,
	phaseIndex int,
	allPossibleLengths []int,
	patternPhases []models.PatternPhase,
	weeksWorkSync *weekPredictionSync,
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
	weeksWorkSync.WaitGroup.Add(1)
	// Launch the branch
	go predictor.branchWeeks(ticker, newBranch, weeksWorkSync)
}

// Calculate all the possible phase permutations for a given price pattern.
func (predictor *Predictor) predictPattern(
	ticker *models.PriceTicker,
	pattern models.PricePattern,
	patternWorkSync *patternsPredictionSync,
) {
	defer patternWorkSync.WaitGroup.Done()

	// Get the base phase progression of this pattern
	patternPhases := pattern.PhaseProgression(ticker)

	// Set up our goroutine sync objects. Each time a new routine is started for a
	// possible week we are going to increment WaitGroup.
	//
	// Finished week predictions will be reported to ResultChan.
	weeksWorkSync := &weekPredictionSync{
		ResultChan: make(chan *models.PotentialWeek, 1000),
		WaitGroup:  new(sync.WaitGroup),
	}

	weeksWorkSync.WaitGroup.Add(1)
	go predictor.branchWeeks(ticker, patternPhases, weeksWorkSync)

	// Set up our result object
	resultPattern := &models.PotentialPattern{
		Pattern: pattern,
	}

	weeksWorkSync.WaitGroup.Wait()
	close(weeksWorkSync.ResultChan)

	for potentialWeek := range weeksWorkSync.ResultChan {
		resultPattern.PotentialWeeks = append(
			resultPattern.PotentialWeeks, potentialWeek,
		)
		resultPattern.Analysis().Update(potentialWeek.Analysis(), false)
		resultPattern.UpdateSpikeFromRange(potentialWeek)
	}

	patternWorkSync.ResultChan <- resultPattern
}

func (predictor *Predictor) Predict() (*models.Prediction, error) {
	result := new(models.Prediction)
	predictor.result = result

	currentWeek := predictor.Ticker

	patternWorkSync := &patternsPredictionSync{
		ResultChan: make(chan *models.PotentialPattern, len(patterns.PATTERNSGAME)),
		WaitGroup:  new(sync.WaitGroup),
	}

	for _, pattern := range patterns.PATTERNSGAME {
		patternWorkSync.WaitGroup.Add(1)
		go predictor.predictPattern(currentWeek, pattern, patternWorkSync)
	}

	patternWorkSync.WaitGroup.Wait()
	close(patternWorkSync.ResultChan)

	validPrices := false
	for potentialPattern := range patternWorkSync.ResultChan {
		if len(potentialPattern.PotentialWeeks) > 0 {
			validPrices = true
		}
		result.Patterns = append(result.Patterns, potentialPattern)
		result.Analysis().Update(potentialPattern.Analysis(), false)
		result.UpdateSpikeFromRange(potentialPattern)
	}

	// If there are no possible price patterns based on this ticker, return an error
	if !validPrices {
		return nil, errs.ErrImpossibleTickerPrices
	}
	predictor.calculateChances(currentWeek, result)

	return result, nil
}
