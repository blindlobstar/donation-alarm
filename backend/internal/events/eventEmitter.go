package events

type EventEmitter interface {
	Publish(event any, name string) error
}
