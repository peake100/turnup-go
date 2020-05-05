package models

import (
	"github.com/peake100/turnup-go/errs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPredictionBadPatternErr(t *testing.T) {
	prediction := new(Prediction)

	_, err := prediction.Pattern(5)
	assert.EqualError(t, err, errs.ErrPatternStringValue.Error())
}