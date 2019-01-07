package event

import (
	"reflect"

	"github.com/domonda/errors"
)

type Transformer interface {
	TransformEvent(source interface{}) (result interface{}, useResult bool, err error)

	SourceEventType() reflect.Type
	ResultEventType() reflect.Type
	IsFilter() bool
	ReturnsError() bool
}

func NewTransformer(transformFunc interface{}) Transformer {
	funcVal := reflect.ValueOf(transformFunc)
	funcType := funcVal.Type()
	if funcType.Kind() != reflect.Func {
		panic(errors.Errorf("transformer must be a function, but is %s", funcType))
	}
	if funcType.NumIn() != 1 {
		panic(errors.Errorf("transformer function must have 1 argument, but has %d", funcType.NumIn()))
	}

	var (
		useResultIndex int
		errResultIndex int
	)

	switch funcType.NumOut() {
	case 1:
		// OK

	case 2:
		switch funcType.Out(1) {
		case reflect.TypeOf(false):
			useResultIndex = 1
		case errors.Type:
			errResultIndex = 1
		default:
			panic(errors.Errorf("transformer function's second return value type must be bool or error, but is %s", funcType.Out(1)))
		}

	case 3:
		if funcType.Out(1) != reflect.TypeOf(false) {
			panic(errors.Errorf("transformer function's second return value type must be bool, but is %s", funcType.Out(1)))
		}
		if funcType.Out(2) != errors.Type {
			panic(errors.Errorf("transformer function's third return value type must be error, but is %s", funcType.Out(2)))
		}
		useResultIndex = 1
		errResultIndex = 2

	default:
		panic(errors.Errorf("transformer function must have 1 to 3 return values, but has %d", funcType.NumOut()))
	}

	return &transformer{funcVal: funcVal, useResultIndex: useResultIndex, errResultIndex: errResultIndex}
}

type transformer struct {
	funcVal        reflect.Value
	useResultIndex int
	errResultIndex int
}

func (t *transformer) TransformEvent(source interface{}) (result interface{}, useResult bool, err error) {
	results := t.funcVal.Call([]reflect.Value{reflect.ValueOf(source)})

	result = results[0].Interface()
	if t.IsFilter() {
		useResult = results[t.useResultIndex].Bool()
	} else {
		useResult = true
	}
	if t.ReturnsError() {
		err, _ = results[t.errResultIndex].Interface().(error)
	}

	return result, useResult, err
}

func (t *transformer) SourceEventType() reflect.Type {
	return t.funcVal.Type().In(0)
}

func (t *transformer) ResultEventType() reflect.Type {
	return t.funcVal.Type().Out(0)
}

func (t *transformer) IsFilter() bool {
	return t.useResultIndex > 0
}

func (t *transformer) ReturnsError() bool {
	return t.errResultIndex > 0
}
