package events

type EventHandler interface {
	Handle(event any) error
}
