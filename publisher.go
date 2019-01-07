package event

// Publisher has methods for publishing events
type Publisher interface {
	// Publish an event without waiting for a result.
	Publish(event interface{})

	// PublishAsync publishes an event asynchronousely
	// using one or more go routines.
	// Exactly one error or nil will be written to
	// the returned channel when the event has been
	// handled by the subsribed handlers.
	// The error can be a combination of multiple
	// errors from multiple event handlers.
	PublishAsync(event interface{}) <-chan error

	// PublishAwait publishes an event and waits
	// for all handlers to return an error or nil.
	// The error can be a combination of multiple
	// errors from multiple event handlers.
	PublishAwait(event interface{}) error
}
