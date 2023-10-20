package model

import "github.com/ostcar/timer/sticky"

// Event is something that can happen in the bietrunde.
type Event = sticky.Event[Model]

// GetEvent returns an empty event.
func GetEvent(eventType string) Event {
	switch eventType {
	default:
		return nil
	}
}
