package models

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPeriod(t *testing.T) {

	weekdays := []time.Weekday{
		time.Monday, time.Monday,
		time.Tuesday, time.Tuesday,
		time.Wednesday, time.Wednesday,
		time.Thursday, time.Thursday,
		time.Friday, time.Friday,
		time.Saturday, time.Saturday,
	}

	timeOfDays := []ToD{
		AM, PM,
		AM, PM,
		AM, PM,
		AM, PM,
		AM, PM,
		AM, PM,
	}

	for i := 0; i < 12; i++ {

		period := PricePeriod(i)

		testWeekday := func(t *testing.T) {
			assert.Equal(t, weekdays[i], period.Weekday(), "weekday")
		}

		testTimeOfDay := func(t *testing.T) {
			assert.Equal(t, timeOfDays[i], period.ToD(), "time of day")
		}

		testPeriod := func(t *testing.T) {
			t.Run("Weekday", testWeekday)
			t.Run("Time of Day", testTimeOfDay)
		}

		t.Run(fmt.Sprint(int(period)), testPeriod)
	}

}
