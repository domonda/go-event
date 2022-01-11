package event

// NopStream is an event stream that implements Publisher and Subscribable
// but does actually nothing (nop = no operation).
type NopStream struct {
	subscribable
}

// MewNopStream returns a new NopStream.
func MewNopStream(subscribeTo ...Subscribable) *NopStream {
	return new(NopStream)
}

// Publish is a dummy.
func (stream *NopStream) Publish(event interface{}) error {
	return nil
}

// PublishAwait is a dummy.
func (stream *NopStream) PublishAsync(event interface{}) <-chan error {
	errChan := make(chan error, 1)
	errChan <- nil
	return errChan
}

// PublishAwait is a dummy.
func (stream *NopStream) PublishAwait(event interface{}) error {
	return nil
}
