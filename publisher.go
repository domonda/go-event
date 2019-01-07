package event

// Publisher has a Publish method to publish events
type Publisher interface {
	// Publish the passed event
	Publish(event interface{}) <-chan error

	PublishAwait(event interface{}) error
}
