package turnup

//revive:disable:import-shadowing reason: Disabled for assert := assert.New(), which is
// the preferred method of using multiple asserts in a test.

import (
	"github.com/peake100/turnup-go/errs"
	"github.com/peake100/turnup-go/models"
	"github.com/peake100/turnup-go/models/patterns"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Tests that we get ALL the correct possibilities with a purchase price of 100 bells
// and no buy price info. This should yield every possible outcome.
//
// We are going to use data from turnip prophet to validate our predictions
func Test100BellPurchase(t *testing.T) {

	ticker := NewPriceTicker(100, patterns.UNKNOWN, 0)

	expected := &expectedPrediction{
		Prices: PriceRange{
			Min:        10,
			Guaranteed: 85,
			Max:        600,
		},
		PricesFuture: PriceRange{
			Min:        10,
			Guaranteed: 85,
			Max:        600,
		},
		Fluctuating: &expectedPattern{
			Chance: 0.35,
			Prices: PriceRange{
				Min:        40,
				Guaranteed: 90,
				Max:        140,
			},
			PricesFuture: PriceRange{
				Min:        40,
				Guaranteed: 90,
				Max:        140,
			},
			PossibleWeeks: 56,
			MinPricePeriods: []models.PricePeriod{
				2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
			},
			GuaranteedPricePeriods: []models.PricePeriod{
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
			},
			MaxPricePeriods: []models.PricePeriod{
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
			},
		},
		BigSpike: &expectedPattern{
			Chance: 0.2625,
			Prices: PriceRange{
				Min:        40,
				Guaranteed: 200,
				Max:        600,
			},
			PricesFuture: PriceRange{
				Min:        40,
				Guaranteed: 200,
				Max:        600,
			},
			PossibleWeeks: 7,
			Spike: expectedSpike{
				Small:      false,
				SmallStart: 0,
				SmallEnd:   0,
				Big:        true,
				BigStart:   3,
				BigEnd:     9,
			},
			MinPricePeriods:        []models.PricePeriod{6, 7, 8, 9, 10, 11},
			GuaranteedPricePeriods: []models.PricePeriod{3, 4, 5, 6, 7, 8, 9},
			MaxPricePeriods:        []models.PricePeriod{3, 4, 5, 6, 7, 8, 9},
		},
		Decreasing: &expectedPattern{
			Chance: 0.1375,
			Prices: PriceRange{
				Min:        30,
				Guaranteed: 85,
				Max:        90,
			},
			PricesFuture: PriceRange{
				Min:        30,
				Guaranteed: 85,
				Max:        90,
			},
			PossibleWeeks:          1,
			MinPricePeriods:        []models.PricePeriod{11},
			GuaranteedPricePeriods: []models.PricePeriod{0},
			MaxPricePeriods:        []models.PricePeriod{0},
		},
		SmallSpike: &expectedPattern{
			Chance: 0.25,
			Prices: PriceRange{
				Min:        10,
				Guaranteed: 140,
				Max:        200,
			},
			PricesFuture: PriceRange{
				Min:        10,
				Guaranteed: 140,
				Max:        200,
			},
			PossibleWeeks: 8,
			Spike: expectedSpike{
				Small:      true,
				SmallStart: 2,
				SmallEnd:   11,
				Big:        false,
				BigStart:   0,
				BigEnd:     0,
			},
			MinPricePeriods:        []models.PricePeriod{6, 11},
			GuaranteedPricePeriods: []models.PricePeriod{3, 4, 5, 6, 7, 8, 9, 10},
			MaxPricePeriods:        []models.PricePeriod{3, 4, 5, 6, 7, 8, 9, 10},
		},
		Spike: expectedSpike{
			Small:      true,
			SmallStart: 2,
			SmallEnd:   11,
			Big:        true,
			BigStart:   3,
			BigEnd:     9,
		},
		MinPricePeriods:        []models.PricePeriod{6, 11},
		GuaranteedPricePeriods: []models.PricePeriod{0},
		MaxPricePeriods:        []models.PricePeriod{3, 4, 5, 6, 7, 8, 9},
		PriceCSV:               "./zdevelop/tests/100_bell_no_ticker.csv",
	}

	testPrediction(t, expected, ticker)

}

// Test a pattern that results in a single large spike possibility
func Test100BellPurchaseBigSpike(t *testing.T) {

	ticker := NewPriceTicker(100, patterns.UNKNOWN, 0)
	ticker.Prices[0] = 86
	ticker.Prices[1] = 90
	ticker.Prices[2] = 160

	expected := &expectedPrediction{
		Prices: PriceRange{
			Min:        40,
			Guaranteed: 200,
			Max:        600,
		},
		PricesFuture: PriceRange{
			Min:        40,
			Guaranteed: 200,
			Max:        600,
		},
		Fluctuating: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		BigSpike: &expectedPattern{
			Chance: 1.0,
			Prices: PriceRange{
				Min:        40,
				Guaranteed: 200,
				Max:        600,
			},
			PricesFuture: PriceRange{
				Min:        40,
				Guaranteed: 200,
				Max:        600,
			},
			PossibleWeeks: 1,
			Spike: expectedSpike{
				Small:      false,
				SmallStart: 0,
				SmallEnd:   0,
				Big:        true,
				BigStart:   3,
				BigEnd:     3,
			},
			MinPricePeriods:        []models.PricePeriod{6, 7, 8, 9, 10, 11},
			GuaranteedPricePeriods: []models.PricePeriod{3},
			MaxPricePeriods:        []models.PricePeriod{3},
		},
		Decreasing: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		SmallSpike: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		Spike: expectedSpike{
			Small:      false,
			SmallStart: 0,
			SmallEnd:   0,
			Big:        true,
			BigStart:   3,
			BigEnd:     3,
		},
		MinPricePeriods:        []models.PricePeriod{6, 7, 8, 9, 10, 11},
		GuaranteedPricePeriods: []models.PricePeriod{3},
		MaxPricePeriods:        []models.PricePeriod{3},
	}

	testPrediction(t, expected, ticker)
}

// Test a pattern that results in a single large spike possibility
func Test100BellPurchaseFluctuating(t *testing.T) {

	ticker := NewPriceTicker(100, patterns.DECREASING, 0)
	ticker.Prices[0] = 140
	ticker.Prices[1] = 140
	ticker.Prices[2] = 140
	ticker.Prices[3] = 140
	ticker.Prices[4] = 140
	ticker.Prices[5] = 140

	expected := &expectedPrediction{
		Prices: PriceRange{
			Min:        40,
			Guaranteed: 90,
			Max:        140,
		},
		PricesFuture: PriceRange{
			Min:        40,
			Guaranteed: 140,
			Max:        140,
		},
		Fluctuating: &expectedPattern{
			Chance: 1,
			Prices: PriceRange{
				Min:        40,
				Guaranteed: 90,
				Max:        140,
			},
			PricesFuture: PriceRange{
				Min:        40,
				Guaranteed: 140,
				Max:        140,
			},
			PossibleWeeks:          2,
			MinPricePeriods:        []models.PricePeriod{8, 11},
			GuaranteedPricePeriods: []models.PricePeriod{0, 1, 2, 3, 4, 5, 8, 9},
			MaxPricePeriods:        []models.PricePeriod{0, 1, 2, 3, 4, 5, 8, 9},
		},
		BigSpike: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		Decreasing: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		SmallSpike: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		MinPricePeriods:        []models.PricePeriod{8, 11},
		GuaranteedPricePeriods: []models.PricePeriod{0, 1, 2, 3, 4, 5, 8, 9},
		MaxPricePeriods:        []models.PricePeriod{0, 1, 2, 3, 4, 5, 8, 9},
	}

	testPrediction(t, expected, ticker)
}

// Test a pattern that results in a decreasing possibility
func Test100BellPurchaseDecreasing(t *testing.T) {

	ticker := NewPriceTicker(100, patterns.DECREASING, 0)
	ticker.Prices[0] = 86
	ticker.Prices[1] = 82
	ticker.Prices[2] = 78
	ticker.Prices[3] = 74
	ticker.Prices[4] = 70
	ticker.Prices[5] = 66
	ticker.Prices[6] = 62
	ticker.Prices[7] = 58

	expected := &expectedPrediction{
		Prices: PriceRange{
			Min:        37,
			Guaranteed: 85,
			Max:        90,
		},
		PricesFuture: PriceRange{
			Min:        37,
			Guaranteed: 58,
			Max:        58,
		},
		Fluctuating: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		BigSpike: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		Decreasing: &expectedPattern{
			Chance: 1,
			Prices: PriceRange{
				Min:        37,
				Guaranteed: 85,
				Max:        90,
			},
			PricesFuture: PriceRange{
				Min:        37,
				Guaranteed: 58,
				Max:        58,
			},
			PossibleWeeks:          1,
			MinPricePeriods:        []models.PricePeriod{11},
			GuaranteedPricePeriods: []models.PricePeriod{0},
			MaxPricePeriods:        []models.PricePeriod{0},
		},
		SmallSpike: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		MinPricePeriods:        []models.PricePeriod{11},
		GuaranteedPricePeriods: []models.PricePeriod{0},
		MaxPricePeriods:        []models.PricePeriod{0},
	}

	testPrediction(t, expected, ticker)
}

// Test a pattern that results in a single large spike possibility
func Test100BellPurchaseSmallSpike(t *testing.T) {

	ticker := NewPriceTicker(100, patterns.SMALLSPIKE, 0)
	ticker.Prices[0] = 120
	ticker.Prices[1] = 120
	ticker.Prices[2] = 199

	expected := &expectedPrediction{
		Prices: PriceRange{
			Min:        10,
			Guaranteed: 140,
			Max:        200,
		},
		PricesFuture: PriceRange{
			Min:        10,
			Guaranteed: 199,
			Max:        200,
		},
		Fluctuating: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		BigSpike: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		Decreasing: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		SmallSpike: &expectedPattern{
			Chance: 1,
			Prices: PriceRange{
				Min:        10,
				Guaranteed: 140,
				Max:        200,
			},
			PricesFuture: PriceRange{
				Min:        10,
				Guaranteed: 199,
				Max:        200,
			},
			PossibleWeeks: 1,
			Spike: expectedSpike{
				Small:      true,
				SmallStart: 2,
				SmallEnd:   4,
				Big:        false,
				BigStart:   0,
				BigEnd:     0,
			},
			MinPricePeriods:        []models.PricePeriod{11},
			GuaranteedPricePeriods: []models.PricePeriod{3},
			MaxPricePeriods:        []models.PricePeriod{3},
		},
		Spike: expectedSpike{
			Small:      true,
			SmallStart: 2,
			SmallEnd:   4,
			Big:        false,
			BigStart:   0,
			BigEnd:     0,
		},
		MinPricePeriods:        []models.PricePeriod{11},
		GuaranteedPricePeriods: []models.PricePeriod{3},
		MaxPricePeriods:        []models.PricePeriod{3},
	}

	testPrediction(t, expected, ticker)
}

// Test getting doing a prediction when you don't know the purchase price.
func TestUnknownBellPurchase(t *testing.T) {

	ticker := NewPriceTicker(0, patterns.UNKNOWN, 0)

	expected := &expectedPrediction{
		Prices: PriceRange{
			Min:        9,
			Guaranteed: 77,
			Max:        660,
		},
		PricesFuture: PriceRange{
			Min:        9,
			Guaranteed: 77,
			Max:        660,
		},
		Fluctuating: &expectedPattern{
			Chance: 0.35,
			Prices: PriceRange{
				Min:        36,
				Guaranteed: 81,
				Max:        154,
			},
			PricesFuture: PriceRange{
				Min:        36,
				Guaranteed: 81,
				Max:        154,
			},
			PossibleWeeks: 56,
			MinPricePeriods: []models.PricePeriod{
				2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
			},
			GuaranteedPricePeriods: []models.PricePeriod{
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
			},
			MaxPricePeriods: []models.PricePeriod{
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
			},
		},
		BigSpike: &expectedPattern{
			Chance: 0.2625,
			Prices: PriceRange{
				Min:        36,
				Guaranteed: 180,
				Max:        660,
			},
			PricesFuture: PriceRange{
				Min:        36,
				Guaranteed: 180,
				Max:        660,
			},
			PossibleWeeks: 7,
			Spike: expectedSpike{
				Small:      false,
				SmallStart: 0,
				SmallEnd:   0,
				Big:        true,
				BigStart:   3,
				BigEnd:     9,
			},
			MinPricePeriods:        []models.PricePeriod{6, 7, 8, 9, 10, 11},
			GuaranteedPricePeriods: []models.PricePeriod{3, 4, 5, 6, 7, 8, 9},
			MaxPricePeriods:        []models.PricePeriod{3, 4, 5, 6, 7, 8, 9},
		},
		Decreasing: &expectedPattern{
			Chance: 0.1375,
			Prices: PriceRange{
				Min:        27,
				Guaranteed: 77,
				Max:        99,
			},
			PricesFuture: PriceRange{
				Min:        27,
				Guaranteed: 77,
				Max:        99,
			},
			PossibleWeeks:          1,
			MinPricePeriods:        []models.PricePeriod{11},
			GuaranteedPricePeriods: []models.PricePeriod{0},
			MaxPricePeriods:        []models.PricePeriod{0},
		},
		SmallSpike: &expectedPattern{
			Chance: 0.25,
			Prices: PriceRange{
				Min:        9,
				Guaranteed: 126,
				Max:        220,
			},
			PricesFuture: PriceRange{
				Min:        9,
				Guaranteed: 126,
				Max:        220,
			},
			PossibleWeeks: 8,
			Spike: expectedSpike{
				Small:      true,
				SmallStart: 2,
				SmallEnd:   11,
				Big:        false,
				BigStart:   0,
				BigEnd:     0,
			},
			MinPricePeriods:        []models.PricePeriod{6, 11},
			GuaranteedPricePeriods: []models.PricePeriod{3, 4, 5, 6, 7, 8, 9, 10},
			MaxPricePeriods:        []models.PricePeriod{3, 4, 5, 6, 7, 8, 9, 10},
		},
		Spike: expectedSpike{
			Small:      true,
			SmallStart: 2,
			SmallEnd:   11,
			Big:        true,
			BigStart:   3,
			BigEnd:     9,
		},
		MinPricePeriods:        []models.PricePeriod{6, 11},
		GuaranteedPricePeriods: []models.PricePeriod{0},
		MaxPricePeriods:        []models.PricePeriod{3, 4, 5, 6, 7, 8, 9},
	}

	testPrediction(t, expected, ticker)
}

// Test submitting an impossible price pattern
func TestImpossiblePattern(t *testing.T) {
	assert := assert.New(t)

	ticker := NewPriceTicker(0, patterns.UNKNOWN, 0)
	ticker.Prices[0] = 10

	result, err := Predict(ticker)
	assert.Nil(result, "result nil")
	assert.EqualError(
		err,
		errs.ErrImpossibleTickerPrices.Error(),
		"impossible prices error",
	)
}

func TestMultiplePossibleMatches(t *testing.T) {
	ticker := NewPriceTicker(100, patterns.DECREASING, 0)
	ticker.Prices[0] = 86
	ticker.Prices[1] = 82

	expected := &expectedPrediction{
		Prices: PriceRange{
			Guaranteed: 85,
			Min:        20,
			Max:        600,
		},
		PricesFuture: PriceRange{
			Guaranteed: 82,
			Min:        20,
			Max:        600,
		},
		Fluctuating: &expectedPattern{
			Chance:                 0.0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		BigSpike: &expectedPattern{
			Chance: 0.6725,
			Prices: PriceRange{
				Min:        40,
				Guaranteed: 200,
				Max:        600,
			},
			PricesFuture: PriceRange{
				Min:        40,
				Guaranteed: 200,
				Max:        600,
			},
			PossibleWeeks: 6,
			Spike: expectedSpike{
				Small:      false,
				SmallStart: 0,
				SmallEnd:   0,
				Big:        true,
				BigStart:   4,
				BigEnd:     9,
			},
			MinPricePeriods:        []models.PricePeriod{7, 8, 9, 10, 11},
			GuaranteedPricePeriods: []models.PricePeriod{4, 5, 6, 7, 8, 9},
			MaxPricePeriods:        []models.PricePeriod{4, 5, 6, 7, 8, 9},
		},
		Decreasing: &expectedPattern{
			Chance: 0.0872,
			Prices: PriceRange{
				Min:        31,
				Guaranteed: 85,
				Max:        90,
			},
			PricesFuture: PriceRange{
				Min:        31,
				Guaranteed: 82,
				Max:        82,
			},
			PossibleWeeks:          1,
			MinPricePeriods:        []models.PricePeriod{11},
			GuaranteedPricePeriods: []models.PricePeriod{0},
			MaxPricePeriods:        []models.PricePeriod{0},
		},
		SmallSpike: &expectedPattern{
			Chance: 0.2404,
			Prices: PriceRange{
				Min:        20,
				Guaranteed: 140,
				Max:        200,
			},
			PricesFuture: PriceRange{
				Min:        20,
				Guaranteed: 140,
				Max:        200,
			},
			PossibleWeeks: 6,
			Spike: expectedSpike{
				Small:      true,
				SmallStart: 4,
				SmallEnd:   11,
				Big:        false,
				BigStart:   0,
				BigEnd:     0,
			},
			MinPricePeriods:        []models.PricePeriod{11},
			GuaranteedPricePeriods: []models.PricePeriod{5, 6, 7, 8, 9, 10},
			MaxPricePeriods:        []models.PricePeriod{5, 6, 7, 8, 9, 10},
		},
		Spike: expectedSpike{
			Small:      true,
			SmallStart: 4,
			SmallEnd:   11,
			Big:        true,
			BigStart:   4,
			BigEnd:     9,
		},
		MinPricePeriods:        []models.PricePeriod{11},
		GuaranteedPricePeriods: []models.PricePeriod{0},
		MaxPricePeriods:        []models.PricePeriod{4, 5, 6, 7, 8, 9},
	}

	testPrediction(t, expected, ticker)
}

// We have special logic for when there is an INCREDIBLY unlikely price patterns. This
// test will trigger it because the actual chances of this pattern occurring are 1 in
// several billion (the bin width comes out to 0)
func Test100BellPurchaseUnlikelyLowerBoundPattern(t *testing.T) {

	ticker := NewPriceTicker(100, patterns.SMALLSPIKE, 0)
	ticker.Prices[0] = 85
	ticker.Prices[1] = 80
	ticker.Prices[2] = 75
	ticker.Prices[3] = 70
	ticker.Prices[4] = 65
	ticker.Prices[5] = 60
	ticker.Prices[6] = 55
	ticker.Prices[7] = 50
	ticker.Prices[8] = 45
	ticker.Prices[9] = 40
	ticker.Prices[10] = 35
	ticker.Prices[11] = 30

	expected := &expectedPrediction{
		Prices: PriceRange{
			Min:        30,
			Guaranteed: 85,
			Max:        90,
		},
		PricesFuture: PriceRange{
			Min:        30,
			Guaranteed: 30,
			Max:        30,
		},
		Fluctuating: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		BigSpike: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		Decreasing: &expectedPattern{
			Chance: 1,
			Prices: PriceRange{
				Min:        30,
				Guaranteed: 85,
				Max:        90,
			},
			PricesFuture: PriceRange{
				Min:        30,
				Guaranteed: 30,
				Max:        30,
			},
			PossibleWeeks:          1,
			MinPricePeriods:        []models.PricePeriod{11},
			GuaranteedPricePeriods: []models.PricePeriod{0},
			MaxPricePeriods:        []models.PricePeriod{0},
		},
		SmallSpike: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		MinPricePeriods:        []models.PricePeriod{11},
		GuaranteedPricePeriods: []models.PricePeriod{0},
		MaxPricePeriods:        []models.PricePeriod{0},
	}

	testPrediction(t, expected, ticker)
}

// We tested the lower bound of a compounding pattern last test, lets try the upper
// bound this time
func Test100BellPurchaseUnlikelyUpperBoundPattern(t *testing.T) {

	ticker := NewPriceTicker(100, patterns.SMALLSPIKE, 0)
	ticker.Prices[0] = 90
	ticker.Prices[1] = 87
	ticker.Prices[2] = 84
	ticker.Prices[3] = 82
	ticker.Prices[4] = 79
	ticker.Prices[5] = 76
	ticker.Prices[6] = 73
	ticker.Prices[7] = 70
	ticker.Prices[8] = 67
	ticker.Prices[9] = 64
	ticker.Prices[10] = 61
	ticker.Prices[11] = 58

	expected := &expectedPrediction{
		Prices: PriceRange{
			Min:        55,
			Guaranteed: 85,
			Max:        90,
		},
		PricesFuture: PriceRange{
			Min:        58,
			Guaranteed: 58,
			Max:        58,
		},
		Fluctuating: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		BigSpike: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		Decreasing: &expectedPattern{
			Chance: 1,
			Prices: PriceRange{
				Min:        55,
				Guaranteed: 85,
				Max:        90,
			},
			PricesFuture: PriceRange{
				Min:        58,
				Guaranteed: 58,
				Max:        58,
			},
			PossibleWeeks:          1,
			MinPricePeriods:        []models.PricePeriod{11},
			GuaranteedPricePeriods: []models.PricePeriod{0},
			MaxPricePeriods:        []models.PricePeriod{0},
		},
		SmallSpike: &expectedPattern{
			Chance:                 0,
			PossibleWeeks:          0,
			MinPricePeriods:        []models.PricePeriod{},
			GuaranteedPricePeriods: []models.PricePeriod{},
			MaxPricePeriods:        []models.PricePeriod{},
		},
		MinPricePeriods:        []models.PricePeriod{11},
		GuaranteedPricePeriods: []models.PricePeriod{0},
		MaxPricePeriods:        []models.PricePeriod{0},
	}

	testPrediction(t, expected, ticker)
}
