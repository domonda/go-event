package event

type Publisher interface {
	Publish(event interface{})
}
