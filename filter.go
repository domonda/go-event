package event

// type Filter interface {
// 	FilterEvent(event interface{}) bool
// }

// type FilterFunc func(event interface{}) bool

// func (f FilterFunc) FilterEvent(event interface{}) bool {
// 	return f(event)
// }

// func FilterSource(source Source, filter Filter) Source {
// 	return SourceFunc(func(handler Handler, eventTypes ...reflect.Type) {
// 		source.Subscribe(
// 			HandlerFunc(func(event interface{}) {
// 				handler.HandleEvent(transformer.TransformEvent(event))
// 			}),
// 			eventTypes...,
// 		)
// 	})
// }
