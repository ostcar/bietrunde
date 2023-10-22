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
	case eventStateSet{}.Name():
		return &eventStateSet{}
	case eventGebot{}.Name():
		return &eventGebot{}
	case eventResetGebot{}.Name():
		return &eventResetGebot{}
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

type eventStateSet struct {
	State ServiceState
}

func (e eventStateSet) Name() string {
	return "set-state"
}

func (e eventStateSet) Validate(model Model) error {
	if int(e.State) <= 0 || int(e.State) > 3 {
		return fmt.Errorf("invalid state")
	}

	return nil
}

func (e eventStateSet) Execute(model Model, time time.Time) Model {
	model.State = e.State
	return model
}

type eventGebot struct {
	BietID int   `json:"bieter"`
	Gebot  Gebot `json:"gebot"`
}

func (e eventGebot) Name() string {
	return "gebot"
}

func (e eventGebot) Validate(model Model) error {
	if _, ok := model.Bieter[e.BietID]; !ok {
		return fmt.Errorf("bieter does not exist")
	}

	return nil
}

func (e eventGebot) Execute(model Model, time time.Time) Model {
	bieter := model.Bieter[e.BietID]
	bieter.Gebot = e.Gebot
	model.Bieter[e.BietID] = bieter
	return model
}

type eventResetGebot struct{}

func (e eventResetGebot) Name() string {
	return "gebot-reset"
}

func (e eventResetGebot) Validate(model Model) error {
	return nil
}

func (e eventResetGebot) Execute(model Model, time time.Time) Model {
	for k, bieter := range model.Bieter {
		bieter.Gebot = 0
		model.Bieter[k] = bieter
	}
	return model
}
