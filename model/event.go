package model

import (
	"fmt"
	"time"

	"github.com/ostcar/sticky"
)

// Event is something that can happen in the bietrunde.
type Event = sticky.Event[Model]

// GetEvent returns an empty event.
func GetEvent(eventType string) Event {
	switch eventType {
	case eventBieterCreate{}.Name():
		return &eventBieterCreate{}
	case eventBieterUpdate{}.Name():
		return &eventBieterUpdate{}
	case eventBieterDelete{}.Name():
		return &eventBieterDelete{}
	default:
		return nil
	}
}

type eventBieterCreate struct {
	ID int `json:"id"`
}

func (e eventBieterCreate) Name() string {
	return "bieter-create"
}

func (e eventBieterCreate) Validate(model Model) error {
	if _, ok := model.Bieter[e.ID]; ok {
		return fmt.Errorf("Bieter id is not unique")
	}

	if model.State != StateRegistration {
		return fmt.Errorf("Registrierung nicht möglich")
	}

	return nil
}

func (e eventBieterCreate) Execute(model Model, time time.Time) Model {
	model.Bieter[e.ID] = Bieter{ID: e.ID}
	return model
}

type eventBieterUpdate struct {
	Bieter
}

func (e eventBieterUpdate) Name() string {
	return "bieter-update"
}

func (e eventBieterUpdate) Validate(model Model) error {
	if _, ok := model.Bieter[e.ID]; !ok {
		return fmt.Errorf("bieter does not exist")
	}

	// TODO

	return nil
}

func (e eventBieterUpdate) Execute(model Model, time time.Time) Model {
	model.Bieter[e.ID] = e.Bieter
	return model
}

type eventBieterDelete struct {
	ID int `json:"id"`
}

func (e eventBieterDelete) Name() string {
	return "bieter-delete"
}

func (e eventBieterDelete) Validate(model Model) error {
	if _, ok := model.Bieter[e.ID]; !ok {
		return fmt.Errorf("bieter does not exist")
	}

	return nil
}

func (e eventBieterDelete) Execute(model Model, time time.Time) Model {
	delete(model.Bieter, e.ID)
	return model
}
