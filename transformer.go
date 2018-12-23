package event

import (
	"reflect"

	"github.com/domonda/errors"
)

type Transformer interface {
	TransformEvent(source interface{}) (result interface{}, useResult bool)
	SourceEventType() reflect.Type
	ResultEventType() reflect.Type
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
	if funcType.NumOut() != 1 && funcType.NumOut() != 2 {
		panic(errors.Errorf("transformer function must have 1 or 2 return values, but has %d", funcType.NumOut()))
	}
	if funcType.NumOut() == 2 && funcType.Out(1) != reflect.TypeOf(false) {
		panic(errors.Errorf("transformer function's second return value type must be bool, but is %T", funcType.Out(1)))
	}
	return &transformer{funcVal}
}

type transformer struct {
	funcVal reflect.Value
}

func (t *transformer) TransformEvent(source interface{}) (result interface{}, useResult bool) {
	results := t.funcVal.Call([]reflect.Value{reflect.ValueOf(source)})
	useResult = true
	if len(results) == 2 {
		useResult = results[1].Bool()
	}
	return results[0].Interface(), useResult
}

func (t *transformer) SourceEventType() reflect.Type {
	return t.funcVal.Type().In(0)
}

func (t *transformer) ResultEventType() reflect.Type {
	return t.funcVal.Type().Out(0)
}
