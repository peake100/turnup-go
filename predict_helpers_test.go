package turnup

//revive:disable:import-shadowing reason: Disabled for assert := assert.New(), which is
// the preferred method of using multiple asserts in a test.

import (
	"encoding/csv"
	"fmt"
	"github.com/peake100/turnup-go/models"
	"github.com/peake100/turnup-go/models/patterns"
	"github.com/peake100/turnup-go/values"
	"github.com/stretchr/testify/assert"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

type expectedSpike struct {
	Small      bool
	SmallStart models.PricePeriod
	SmallEnd   models.PricePeriod

	Big      bool
	BigStart models.PricePeriod
	BigEnd   models.PricePeriod
}

type priceBracket struct {
	Min int
	Max int
}

type expectedWeek struct {
	Pattern            models.PricePattern
	GuaranteedMinPrice int
	MaxPrice           int
	Prices             [values.PricePeriodCount]*priceBracket
}

type expectedPattern struct {
	Chance             float64
	MinGuaranteedPrice int
	MaxPotentialPrice  int
	PossibleWeeks      int
	Spike              expectedSpike
	MinPricePeriods    []models.PricePeriod
	MaxPricePeriods    []models.PricePeriod
}

type expectedPrediction struct {
	Fluctuating        *expectedPattern
	BigSpike           *expectedPattern
	Decreasing         *expectedPattern
	SmallSpike         *expectedPattern
	PriceCSV           string
	Spike              expectedSpike
	expectedWeekHashes map[string]interface{}
	MinGuaranteedPrice int
	MaxPotentialPrice  int
	MinPricePeriods    []models.PricePeriod
	MaxPricePeriods    []models.PricePeriod
}

func (expected *expectedPrediction) Patterns() []*expectedPattern {
	return []*expectedPattern{
		expected.Fluctuating,
		expected.BigSpike,
		expected.Decreasing,
		expected.SmallSpike,
	}
}

func parsePriceRecordPattern(week *expectedWeek, dataString string) {
	var err error
	week.Pattern, err = models.PatternFromString(dataString)
	if err != nil {
		panic(fmt.Sprintf("cannot parse pattern %v", dataString))
	}
}

func parsePriceRecordPeriod(week *expectedWeek, dataString string, pricePeriod int) {
	prices := strings.Split(dataString, "-")
	minPrice, err := strconv.Atoi(prices[0])
	if err != nil {
		panic(fmt.Sprintf("cannot parse min from %v", prices[0]))
	}
	maxPrice, err := strconv.Atoi(prices[1])
	if err != nil {
		panic(fmt.Sprintf("cannot parse max from %v", prices[1]))
	}

	week.Prices[pricePeriod] = &priceBracket{
		Min: minPrice,
		Max: maxPrice,
	}
}

func parseWeekPriceBound(dataString string) int {
	price, err := strconv.Atoi(dataString)
	if err != nil {
		panic(fmt.Sprintf("cannot parse week price from %v", dataString))
	}
	return price
}

func parsePriceRecord(
	record []string,
) (week *expectedWeek) {
	week = new(expectedWeek)

	for column, dataString := range record {
		switch {
		case column == 0:
			parsePriceRecordPattern(week, dataString)
		case column > 0 && column < 13:
			pricePeriod := column - 1
			parsePriceRecordPeriod(week, dataString, pricePeriod)
		case column == 13:
			week.GuaranteedMinPrice = parseWeekPriceBound(dataString)
		case column == 14:
			week.MaxPrice = parseWeekPriceBound(dataString)
		}
	}

	return week
}

// We're going to make a unique string for a weekly price pattern, which we can add to
// a dict like a set.
func makeWeekKey(
	pattern models.PricePattern,
	prices [values.PricePeriodCount]*priceBracket,
	min int,
	max int,
) string {
	key := pattern.String()
	for i, price := range prices {
		key = fmt.Sprintf("%v-%v:(min:%v,max:%v)", key, i, price.Min, price.Max)
	}
	key = fmt.Sprintf("%v-min:%v-max:%v", key, min, max)
	return key
}

// load price data from csv
func loadPriceData(t *testing.T, csvPath string) map[string]interface{} {
	result := make(map[string]interface{})

	fileReader, err := os.Open(csvPath)
	if err != nil {
		panic("error opening price csv")
	}
	defer func() {
		_ = fileReader.Close()
	}()

	data := csv.NewReader(fileReader)
	data.TrimLeadingSpace = true

	for {
		record, err := data.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic("error reading price record")
		}
		week := parsePriceRecord(record)
		key := makeWeekKey(
			week.Pattern, week.Prices, week.GuaranteedMinPrice, week.MaxPrice,
		)
		t.Logf("expected price pattern: %v\n", key)
		// Add the key to the map
		result[key] = nil
	}

	return result
}

func potentialWeekKey(pattern models.PricePattern, week *models.PotentialWeek) string {
	var priceBrackets [values.PricePeriodCount]*priceBracket
	for i, pricePeriod := range week.Prices {
		periodBracket := &priceBracket{
			Min: pricePeriod.MinPrice(),
			Max: pricePeriod.MaxPrice(),
		}
		priceBrackets[i] = periodBracket
	}

	return makeWeekKey(
		pattern, priceBrackets, week.MinPrice(), week.MaxPrice(),
	)
}

func testPriceData(
	t *testing.T,
	expected *expectedPrediction,
	prediction *models.Prediction,
) {
	// We have a dict with unique keys for every expected price pattern. We're going to
	// go through the actual price patterns recorded and make what we hope will be
	// a matching key and see if it is part of our hash table. If it is, we'll remove
	// it. If not, we'll add am error to the test.
	//
	// If the map is not empty by the end, then we did not have a pattern we expected,
	// and will report the error.
	for _, pattern := range prediction.Patterns {
		for _, week := range pattern.PotentialWeeks {
			weekKey := potentialWeekKey(pattern.Pattern, week)
			if _, ok := expected.expectedWeekHashes[weekKey]; !ok {
				t.Errorf("unexpected price pattern: %v", weekKey)
				continue
			}
			// Delete this key so we know we got it.
			delete(expected.expectedWeekHashes, weekKey)
		}
	}

	for key := range expected.expectedWeekHashes {
		t.Errorf("pattern not generated: %v", key)
	}
}

func testExpectedSpikeAnyHasSpike(
	t *testing.T,
	expected *expectedSpike,
	predicted models.HasSpikeRange,
) {
	assert := assert.New(t)
	assert.True(predicted.HasSpikeAny(), "has any spike")

	var expectedStart models.PricePeriod
	var expectedEnd models.PricePeriod

	if expected.Big {
		expectedStart = expected.BigStart
		expectedEnd = expected.BigEnd
	}

	if expected.Small {
		if !expected.Big || expected.SmallStart < expectedStart {
			expectedStart = expected.SmallStart
		}
		if !expected.Big || expected.SmallEnd > expectedEnd {
			expectedEnd = expected.SmallEnd
		}
	}

	assert.Equal(
		expectedStart,
		predicted.SpikeAnyStart(),
		"start for any spike",
	)

	assert.Equal(
		expectedEnd,
		predicted.SpikeAnyEnd(),
		"end for any spike",
	)
}

func testExpectedSpikeAnyNoSpike(
	t *testing.T,
	predicted models.HasSpikeRange,
) {
	assert := assert.New(t)

	assert.False(predicted.HasSpikeAny(), "does not have any spike")
	assert.Equal(
		models.PricePeriod(0),
		predicted.SpikeAnyStart(),
		"no spike start",
	)
	assert.Equal(
		models.PricePeriod(0),
		predicted.SpikeAnyEnd(),
		"no spike end",
	)
}

func testExpectedSpikeAny(
	t *testing.T,
	expected *expectedSpike,
	predicted models.HasSpikeRange,
) {

	if expected.Big || expected.Small {
		testExpectedSpikeAnyHasSpike(t, expected, predicted)
	} else {
		testExpectedSpikeAnyNoSpike(t, predicted)
	}
}

func testExpectedSpike(
	t *testing.T,
	expected *expectedSpike,
	predicted models.HasSpikeRange,
) {
	assert := assert.New(t)

	assert.Equal(expected.Big, predicted.HasSpikeBig(), "has big spike")
	assert.Equal(
		expected.Small, predicted.HasSpikeSmall(), "has small spike",
	)

	assert.Equal(
		expected.BigStart, predicted.SpikeBigStart(), "big spike start",
	)
	assert.Equal(
		expected.BigEnd, predicted.SpikeBigEnd(), "big spike end",
	)

	assert.Equal(
		expected.SmallStart,
		predicted.SpikeSmallStart(),
		"small spike start",
	)
	assert.Equal(
		expected.SmallEnd, predicted.SpikeSmallEnd(), "big spike end",
	)

	testExpectedSpikeAny(t, expected, predicted)
}

func testSpikesDensity(
	t *testing.T, prediction *models.Prediction,
) {
	assert := assert.New(t)

	bigSpike, _ := prediction.Patterns.Get(models.BIGSPIKE)
	smallSpike, _ := prediction.Patterns.Get(models.SMALLSPIKE)

	assert.Equal(
		bigSpike.Chance(),
		prediction.Spikes.BigChance,
		"big spike chance equals pattern",
	)

	assert.Equal(
		smallSpike.Chance(),
		prediction.Spikes.SmallChance,
		"small spike chance equals pattern",
	)

	assert.Equal(
		bigSpike.Chance()+smallSpike.Chance(),
		prediction.Spikes.AnyChance,
		"total spike chance equals big + small",
	)

	var bigSpikeTotal, smallSpikeTotal, anySpikeTotal float64

	for i := 0; i < values.PricePeriodCount; i++ {
		smallChancePeriod := prediction.Spikes.SmallBreakdown[i]
		bigChancePeriod := prediction.Spikes.BigBreakdown[i]
		anyChancePeriod := prediction.Spikes.AnyBreakdown[i]

		bigSpikeTotal += bigChancePeriod
		smallSpikeTotal += smallChancePeriod
		anySpikeTotal += anyChancePeriod

		assert.Equal(
			smallChancePeriod+bigChancePeriod,
			anyChancePeriod,
			fmt.Sprintf("any chance for period %v equals small + big", i),
		)
	}

	// There are going to be some floating point errors when we add up all the floats
	// for the density map, check that we are within an acceptable bound (less than)
	// 0.05%
	bigVariance := math.Abs(bigSpikeTotal - bigSpike.Chance())
	assert.Less(bigVariance, 0.0005, "big spike density total")

	smallVariance := math.Abs(smallSpikeTotal - smallSpike.Chance())
	assert.Less(smallVariance, 0.0005, "small spike density total")

	anyVariance := math.Abs(anySpikeTotal - (smallSpike.Chance() + bigSpike.Chance()))
	assert.Less(anyVariance, 0.0005, "any spike density total")
}

// We can use this function to test a prediction for a given ticker against our expected
// results
func testPrediction(
	t *testing.T, expected *expectedPrediction, ticker *models.PriceTicker,
) {
	prediction, err := Predict(ticker)
	assert.NoError(t, err, "prices are not possible")
	if err != nil {
		t.FailNow()
	}

	var thisExpected *expectedPattern
	var thisPattern *models.PotentialPattern

	testPattern := func(t *testing.T) {
		testPattern(t, thisExpected, thisPattern)
	}

	expectedPatterns := expected.Patterns()

	for _, pattern := range patterns.PATTERNSGAME {

		thisExpected = expectedPatterns[pattern]
		thisPattern, err = prediction.Patterns.Get(pattern)

		assert.NoError(t, err)
		t.Run(pattern.String(), testPattern)
	}

	testSpike := func(t *testing.T) {
		testExpectedSpike(t, &expected.Spike, &prediction.Spikes)
	}
	t.Run("spike_info", testSpike)

	testMinPrice := func(t *testing.T) {
		assert.Equal(t, expected.MinGuaranteedPrice, prediction.MinPrice())
	}
	t.Run("min guaranteed price", testMinPrice)

	testMaxPrice := func(t *testing.T) {
		assert.Equal(t, expected.MaxPotentialPrice, prediction.MaxPrice())
	}
	t.Run("max potential price", testMaxPrice)

	testMinPeriods := func(t *testing.T) {
		assert.Equal(t, expected.MinPricePeriods, prediction.MinPeriods())
	}
	t.Run("min price periods", testMinPeriods)

	testMaxPeriods := func(t *testing.T) {
		assert.Equal(t, expected.MaxPricePeriods, prediction.MaxPeriods())
	}
	t.Run("max price periods", testMaxPeriods)

	testSpikeDensity := func(t *testing.T) {
		testSpikesDensity(t, prediction)
	}
	t.Run("spikes density", testSpikeDensity)

	if expected.PriceCSV != "" {
		expected.expectedWeekHashes = loadPriceData(t, expected.PriceCSV)
		testPrices := func(t *testing.T) {
			testPriceData(t, expected, prediction)
		}

		t.Run("price_data_check", testPrices)
	}
}

// Test the expected values of a ticker against the actual result
func testPattern(
	t *testing.T, expected *expectedPattern, pattern *models.PotentialPattern,
) {
	testPatternPermutations := func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(
			expected.PossibleWeeks, len(pattern.PotentialWeeks),
			fmt.Sprintf("%v permutations", pattern.Pattern),
		)
	}

	t.Run("permutation_count", testPatternPermutations)

	testPatternChance := func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(
			expected.Chance, pattern.Chance(),
			fmt.Sprintf("%v chance", pattern.Pattern),
		)
	}

	t.Run("chance", testPatternChance)

	testPricePeriodCount := func(t *testing.T) {
		assert := assert.New(t)

		for _, week := range pattern.PotentialWeeks {
			assert.Len(
				week.Prices,
				values.PricePeriodCount,
				"price period count should be 12",
			)
		}
	}

	t.Run("weekly price period count", testPricePeriodCount)

	testPriceMin := func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(
			expected.MinGuaranteedPrice, pattern.MinPrice(),
			"minimum guaranteed price",
		)
	}

	t.Run("min price", testPriceMin)

	testPriceMax := func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(
			expected.MaxPotentialPrice,
			pattern.MaxPrice(),
			"max potential price",
		)
	}

	t.Run("max price", testPriceMax)

	testSpikeInfo := func(t *testing.T) {
		testExpectedSpike(t, &expected.Spike, pattern.Spikes)
	}

	t.Run("spike info", testSpikeInfo)

	testMinPricePeriods := func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(expected.MinPricePeriods, pattern.MinPeriods())
	}

	t.Run("min price periods", testMinPricePeriods)

	testMaxPricePeriods := func(t *testing.T) {
		assert := assert.New(t)
		assert.Equal(expected.MaxPricePeriods, pattern.MaxPeriods())
	}

	t.Run("max price periods", testMaxPricePeriods)
}