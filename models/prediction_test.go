package models

import (
	"github.com/peake100/turnup-go/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPredictionBadPatternErr(t *testing.T) {
	prediction := &Prediction{
		Patterns: make(Patterns, 0),
	}

	_, err := prediction.Patterns.Get(5)
	assert.EqualError(t, err, errs.ErrPatternStringValue.Error())
}
