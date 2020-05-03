package turnup

//revive:disable:import-shadowing reason: Disabled for assert := assert.New(), which is
// the preferred method of using multiple asserts in a test.

import (
	"encoding/csv"
	"fmt"
	"github.com/illuscio-dev/turnup-go/models"
	"github.com/illuscio-dev/turnup-go/patterns"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

type priceBracket struct {
	Min int
	Max int
}

type expectedWeek struct {
	Pattern            models.Pattern
	GuaranteedMinPrice int
	MaxPrice           int
	Prices             [12]*priceBracket
}

type expectedPattern struct {
	Chance             float64
	MinGuaranteedPrice int
	MaxPotentialPrice  int
	PossibleWeeks      int
}

type expectedPrediction struct {
	Fluctuating        *expectedPattern
	BigSpike           *expectedPattern
	Decreasing         *expectedPattern
	SmallSpike         *expectedPattern
	PriceCSV           string
	expectedWeekHashes map[string]interface{}
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
	pattern models.Pattern, prices [12]*priceBracket, min int, max int,
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

func potentialWeekKey(pattern models.Pattern, week *models.PotentialWeek) string {
	var priceBrackets [12]*priceBracket
	for i, pricePeriod := range week.PricePeriods {
		periodBracket := &priceBracket{
			Min: pricePeriod.MinPrice(),
			Max: pricePeriod.MaxPrice(),
		}
		priceBrackets[i] = periodBracket
	}

	return makeWeekKey(
		pattern, priceBrackets, week.Analysis().MinPrice(), week.Analysis().MaxPrice(),
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

// We can use this function to test a prediction for a given ticker against our expected
// results
func testPrediction(
	t *testing.T, expected *expectedPrediction, ticker *PriceTicker,
) {
	prediction := Predict(ticker)

	var thisExpected *expectedPattern
	var thisPattern *models.PotentialPattern
	var err error

	testPattern := func(t *testing.T) {
		testPattern(t, thisExpected, thisPattern)
	}

	expectedPatterns := expected.Patterns()

	for _, pattern := range patterns.PATTERNSGAME {

		thisExpected = expectedPatterns[pattern]
		thisPattern, err = prediction.Pattern(pattern)

		assert.NoError(t, err)
		t.Run(pattern.String(), testPattern)
	}

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
	testPatternChance := func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(
			expected.Chance, pattern.Analysis().Chance,
			fmt.Sprintf("%v chance", pattern.Pattern),
		)
	}

	t.Run("chance", testPatternChance)

	testPricePeriodCount := func(t *testing.T) {
		assert := assert.New(t)

		for _, week := range pattern.PotentialWeeks {
			assert.Len(
				week.PricePeriods,
				12,
				"price period count should be 12",
			)
		}
	}

	t.Run("weekly price period count", testPricePeriodCount)

	testPriceMin := func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(
			expected.MinGuaranteedPrice, pattern.Analysis().MinPrice(),
			"minimum guaranteed price",
		)
	}

	t.Run("min price", testPriceMin)

	testPriceMax := func(t *testing.T) {
		assert := assert.New(t)

		assert.Equal(
			expected.MaxPotentialPrice,
			pattern.Analysis().MaxPrice(),
			"max potential price",
		)
	}

	t.Run("max price", testPriceMax)
}

// Tests that we get ALL the correct possibilities with a purchase price of 100 bells
// and no buy price info. This should yield every possible outcome.
//
// We are going to use data from turnip prophet to validate our predictions
func Test100BellPurchase(t *testing.T) {

	ticker := &PriceTicker{
		PreviousPattern: patterns.UNKNOWN,
		PurchasePrice:   100,
	}

	expected := &expectedPrediction{
		Fluctuating: &expectedPattern{
			Chance:             0.35,
			MinGuaranteedPrice: 90,
			MaxPotentialPrice:  140,
			PossibleWeeks:      56,
		},
		BigSpike: &expectedPattern{
			Chance:             0.2625,
			MinGuaranteedPrice: 200,
			MaxPotentialPrice:  600,
			PossibleWeeks:      7,
		},
		Decreasing: &expectedPattern{
			Chance:             0.1375,
			MinGuaranteedPrice: 85,
			MaxPotentialPrice:  90,
			PossibleWeeks:      1,
		},
		SmallSpike: &expectedPattern{
			Chance:             0.25,
			MinGuaranteedPrice: 140,
			MaxPotentialPrice:  200,
			PossibleWeeks:      8,
		},
		PriceCSV: "./zdevelop/tests/100_bell_no_ticker.csv",
	}

	testPrediction(t, expected, ticker)

}

// Test a pattern that results in a single large spike possibility
func Test100BellPurchaseLargeSpike(t *testing.T) {

	ticker := &PriceTicker{
		PreviousPattern: patterns.UNKNOWN,
		PurchasePrice:   100,
	}
	ticker.Prices[0] = 86
	ticker.Prices[1] = 90
	ticker.Prices[2] = 160

	expected := &expectedPrediction{
		Fluctuating: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
		BigSpike: &expectedPattern{
			Chance:             1.0,
			MinGuaranteedPrice: 200,
			MaxPotentialPrice:  600,
			PossibleWeeks:      1,
		},
		Decreasing: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
		SmallSpike: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
	}

	testPrediction(t, expected, ticker)
}

// Test a pattern that results in a single large spike possibility
func Test100BellPurchaseFluctuating(t *testing.T) {

	ticker := &PriceTicker{
		PreviousPattern: patterns.DECREASING,
		PurchasePrice:   100,
	}
	ticker.Prices[0] = 140
	ticker.Prices[1] = 140
	ticker.Prices[2] = 140
	ticker.Prices[3] = 140
	ticker.Prices[4] = 140
	ticker.Prices[5] = 140

	expected := &expectedPrediction{
		Fluctuating: &expectedPattern{
			Chance:             1,
			MinGuaranteedPrice: 90,
			MaxPotentialPrice:  140,
			PossibleWeeks:      2,
		},
		BigSpike: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
		Decreasing: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
		SmallSpike: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
	}

	testPrediction(t, expected, ticker)
}

// Test a pattern that results in a decreasing possibility
func Test100BellPurchaseDecreasing(t *testing.T) {

	ticker := &PriceTicker{
		PreviousPattern: patterns.DECREASING,
		PurchasePrice:   100,
	}
	ticker.Prices[0] = 86
	ticker.Prices[1] = 86
	ticker.Prices[2] = 80
	ticker.Prices[3] = 76
	ticker.Prices[4] = 72
	ticker.Prices[5] = 68
	ticker.Prices[6] = 64
	ticker.Prices[7] = 59

	expected := &expectedPrediction{
		Fluctuating: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
		BigSpike: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
		Decreasing: &expectedPattern{
			Chance:             1,
			MinGuaranteedPrice: 85,
			MaxPotentialPrice:  90,
			PossibleWeeks:      1,
		},
		SmallSpike: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
	}

	testPrediction(t, expected, ticker)
}

// Test a pattern that results in a single large spike possibility
func Test100BellPurchaseSmallSpike(t *testing.T) {

	ticker := &PriceTicker{
		PreviousPattern: patterns.SMALLSPIKE,
		PurchasePrice:   100,
	}
	ticker.Prices[0] = 120
	ticker.Prices[1] = 120
	ticker.Prices[2] = 199

	expected := &expectedPrediction{
		Fluctuating: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
		BigSpike: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
		Decreasing: &expectedPattern{
			Chance:             0,
			MinGuaranteedPrice: 0,
			MaxPotentialPrice:  0,
			PossibleWeeks:      0,
		},
		SmallSpike: &expectedPattern{
			Chance:             1,
			MinGuaranteedPrice: 140,
			MaxPotentialPrice:  200,
			PossibleWeeks:      1,
		},
	}

	testPrediction(t, expected, ticker)
}

// Test getting doing a prediction when you don't know the purchase price.
func TestUnknownBellPurchase(t *testing.T) {

	ticker := &PriceTicker{
		PreviousPattern: patterns.UNKNOWN,
	}

	expected := &expectedPrediction{
		Fluctuating: &expectedPattern{
			Chance:             0.35,
			MinGuaranteedPrice: 81,
			MaxPotentialPrice:  154,
			PossibleWeeks:      56,
		},
		BigSpike: &expectedPattern{
			Chance:             0.2625,
			MinGuaranteedPrice: 180,
			MaxPotentialPrice:  660,
			PossibleWeeks:      7,
		},
		Decreasing: &expectedPattern{
			Chance:             0.1375,
			MinGuaranteedPrice: 77,
			MaxPotentialPrice:  99,
			PossibleWeeks:      1,
		},
		SmallSpike: &expectedPattern{
			Chance:             0.25,
			MinGuaranteedPrice: 126,
			MaxPotentialPrice:  220,
			PossibleWeeks:      8,
		},
	}

	testPrediction(t, expected, ticker)
}
