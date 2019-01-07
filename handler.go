package event

import (
	"encoding/json"
	"io"
	"reflect"

	"github.com/domonda/errors"
)

// Handler has a method to handle an event of type interface{}
type Handler interface {
	// HandleEvent handles an event
	HandleEvent(event interface{}) error
}

// HandlerFunc implements Handler for a function pointer
type HandlerFunc func(event interface{}) error

func (f HandlerFunc) HandleEvent(event interface{}) error { return f(event) }

// HandlerFuncNoError implements Handler for a function pointer
type HandlerFuncNoError func(event interface{})

func (f HandlerFuncNoError) HandleEvent(event interface{}) error {
	f(event)
	return nil
}

// RepublishHandler wraps a Publisher as a Handler
// by calling publisher.Publish(event) in every
// HandleEvent method call.
func RepublishHandler(publisher Publisher) Handler {
	return HandlerFunc(func(event interface{}) error {
		err := publisher.Publish(event)
		return <-err
	})
}

// ChanHandler wraps a channel as an event Handler.
// HandleEvent calls will write the passed events to the channel.
func ChanHandler(handlerChan chan<- interface{}) Handler {
	return HandlerFunc(func(event interface{}) error {
		handlerChan <- event
		return nil
	})
}

// WriteJSONHandler returns a Handler that pretty prints
// every event as JSON followed by a newline to writer.
func WriteJSONHandler(writer io.Writer) Handler {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	return HandlerFunc(func(event interface{}) error {
		err := encoder.Encode(event)
		if err != nil {
			return err
		}
		_, err = writer.Write([]byte{'\n'})
		return err
	})
}

// TypeHandler
func TypeHandler(handlerFuncOrChan interface{}) (eventType reflect.Type, handler Handler) {
	handlerVal := reflect.ValueOf(handlerFuncOrChan)
	handlerType := handlerVal.Type()

	switch handlerType.Kind() {
	case reflect.Func:
		if handlerType.NumIn() != 1 {
			panic(errors.Errorf("event handler function must have 1 argument, but has %d", handlerType.NumIn()))
		}
		if handlerType.NumOut() > 1 {
			panic(errors.Errorf("event handler function must not have more than one error result, but has %d", handlerType.NumOut()))
		}
		if handlerType.NumOut() == 1 && handlerType.Out(0) != errors.Type {
			panic(errors.Errorf("event handler function must return an error, but returns %s", handlerType.Out(0)))
		}
		eventType = handlerType.In(0)
		returnsError := handlerType.NumOut() == 1
		handler = HandlerFunc(func(event interface{}) error {
			results := handlerVal.Call([]reflect.Value{reflect.ValueOf(event)})
			if !returnsError {
				return nil
			}
			err, _ := results[0].Interface().(error)
			return err
		})

	case reflect.Chan:
		eventType = handlerType.Elem()
		handler = HandlerFunc(func(event interface{}) error {
			handlerVal.Send(reflect.ValueOf(event))
			return nil
		})

	default:
		panic(errors.Errorf("event handler is not a function or channel but '%T'", handlerFuncOrChan))
	}

	return eventType, handler
}
