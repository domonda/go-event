package event

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTransformer(t *testing.T) {
	var (
		transformer Transformer
		transformed interface{}
		useResult   bool
		err         error
	)

	transformer = NewTransformer(func(i int) float64 {
		return float64(i)
	})
	assert.Equal(t, reflect.TypeOf(int(666)), transformer.SourceEventType())
	assert.Equal(t, reflect.TypeOf(float64(666)), transformer.ResultEventType())
	assert.False(t, transformer.IsFilter())

	transformed, useResult, err = transformer.TransformEvent(int(666))
	assert.NoError(t, err)
	assert.True(t, useResult)
	assert.Exactly(t, float64(666), transformed)

	transformer = NewTransformer(func(i int) (float64, bool) {
		return float64(i), true
	})
	assert.Equal(t, reflect.TypeOf(int(666)), transformer.SourceEventType())
	assert.Equal(t, reflect.TypeOf(float64(666)), transformer.ResultEventType())
	assert.True(t, transformer.IsFilter())

	transformed, useResult, err = transformer.TransformEvent(int(666))
	assert.NoError(t, err)
	assert.True(t, useResult)
	assert.Exactly(t, float64(666), transformed)

	transformer = NewTransformer(func(i int) (float64, bool, error) {
		return 0, false, nil
	})
	transformed, useResult, err = transformer.TransformEvent(int(666))
	assert.NoError(t, err)
	assert.False(t, useResult)

	invalidTransformerFuncs := []interface{}{
		"not a function",
		func() {},
		func() (i int) { return },
		func() (b bool) { return },
		func(int, int) (i int) { return },
		func(int, int) (b bool) { return },
		func(int) {},
		func(int) (r0, r1 int) { return },
		func(int) (r0, r1 int, b bool) { return },
		func(int) (r0 int, r1 error, r2 bool) { return },
	}

	for _, transformerFunc := range invalidTransformerFuncs {
		assert.Panics(t, func() {
			NewTransformer(transformerFunc)
		})
	}

	transformer = NewTransformer(func(i int) (float64, bool, error) {
		return 0, false, errors.New("resultError")
	})
	transformed, useResult, err = transformer.TransformEvent(int(666))
	assert.EqualError(t, err, "resultError")
	assert.False(t, useResult)

	transformer = NewTransformer(func(i int) (int, bool, error) {
		return -1, true, errors.New("resultError")
	})
	transformed, useResult, err = transformer.TransformEvent(int(666))
	assert.EqualError(t, err, "resultError")
	assert.True(t, useResult)
	assert.Exactly(t, int(-1), transformed)
}
