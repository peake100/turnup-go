package models

type patternPredictor struct {
	// Info
	Ticker  *PriceTicker
	Pattern PricePattern

	// The total probability width of this pattern
	binWidth float64

	result *PotentialPattern
}

func (predictor *patternPredictor) increaseBinWidth(amount float64) {
	predictor.binWidth += amount
}

// Makes a duplicate of the current phase pattern to be a new possibility
func (predictor *patternPredictor) duplicatePhasePattern(
	patternPhases []PatternPhase,
) []PatternPhase {

	dupedPhases := make([]PatternPhase, len(patternPhases))
	for i, phase := range patternPhases {
		var branchPhase PatternPhase

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
func (predictor *patternPredictor) addWeekFromFinalizedPhases(
	patternPhases []PatternPhase,
) {
	thisWeekPredictor := &weekPredictor{
		Ticker:        predictor.Ticker,
		Pattern:       predictor.Pattern,
		PatternPhases: patternPhases,
	}

	potentialWeek, binWidth := thisWeekPredictor.Predict()
	if potentialWeek == nil {
		return
	}

	result := predictor.result

	// Otherwise, add the result and updatePrices all of our pattern's stats
	result.PotentialWeeks = append(result.PotentialWeeks, potentialWeek)
	result.updatePriceRangeFromOther(potentialWeek)
	result.Spikes.updateSpikeFromRange(potentialWeek.Spikes)
	predictor.increaseBinWidth(binWidth)
}

// Creates a new branch and calculates it's possible permutations
func (predictor *patternPredictor) computeBranch(
	thisPossibleLength int,
	possibilityIndex int,
	phaseIndex int,
	allPossibleLengths []int,
	patternPhases []PatternPhase,
) {
	var newBranch []PatternPhase
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
	patternPhases []PatternPhase,
) {
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

		// We don't have a complete branch if we are permutating possibilities, so
		// return
		return
	}

	// If we make it all the way through than we have hit a fully formed possible phase
	// pattern! Now we can compute the possible prices and return them as the result
	predictor.addWeekFromFinalizedPhases(patternPhases)
}

func (predictor *patternPredictor) setup() {
	predictor.result = &PotentialPattern{
		Analysis: new(Analysis),
		Pattern:  predictor.Pattern,
		Spikes:   new(SpikeRange),
	}
}

// Calculate all the possible phase permutations for a given price pattern.
func (predictor *patternPredictor) Predict() (
	result *PotentialPattern, binWidth float64,
) {
	predictor.setup()

	// Get the base phase progression of this pattern
	patternPhases := predictor.Pattern.PhaseProgression(predictor.Ticker)
	predictor.branchPhases(patternPhases)

	// Store the total width in the analysis object for now
	predictor.result.chance = predictor.binWidth
	return predictor.result, predictor.binWidth
}
