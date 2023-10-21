package model

import (
	"fmt"
	"math/rand"
	"strconv"
)

// Gebot is an offer.
type Gebot int

func (g Gebot) String() string {
	euro := int(g) / 100
	cent := int(g) % 100
	return fmt.Sprintf("%d,%02d €", euro, cent)
}

// NumberString returns a number for the form field.
func (g Gebot) NumberString() string {
	euro := int(g) / 100
	cent := int(g) % 100

	if cent == 0 {
		return strconv.Itoa(euro)
	}
	return fmt.Sprintf("%d,%02d", euro, cent)
}

// Empty is true, if there is no offer.
func (g Gebot) Empty() bool {
	return int(g) == 0
}

// Bieter is a person that makes an offer.
type Bieter struct {
	ID            int           `json:"id"`
	Vorname       string        `json:"vorname"`
	Nachname      string        `json:"nachname"`
	Mail          string        `json:"mail"`
	Adresse       string        `json:"adresse"`
	Mitglied      bool          `json:"mitglied"`
	Verteilstelle Verteilstelle `json:"verteilstelle"`
	Teilpartner   string        `json:"teilpartner"`
	IBAN          string        `json:"iban"`
	Kontoinhaber  string        `json:"kontoinhaber"`
	Jaehrlich     bool          `json:"jaehrlich"`
	Gebot         Gebot         `json:"gebot"`
}

// Name returns the full name.
func (b Bieter) Name() string {
	return b.Vorname + " " + b.Nachname
}

// ShowKontoinhaber the kontoinhaber
func (b Bieter) ShowKontoinhaber() string {
	if b.Kontoinhaber != "" {
		return b.Kontoinhaber
	}
	return b.Name()
}

// Model of the service.
type Model struct {
	Bieter map[int]Bieter
	State  ServiceState
}

// New returns an initialized model.
func New() Model {
	return Model{
		Bieter: make(map[int]Bieter),
		State:  StateRegistration,
	}
}

// BieterCreate creates a new bieter with empty data.
func (m Model) BieterCreate() (int, Event) {
	id := rand.Intn(100_000_000)
	return id, eventBieterCreate{ID: id}
}

// BieterUpdate updates all fields of a bieter.
func (m Model) BieterUpdate(bieter Bieter) Event {
	return eventBieterUpdate{bieter}
}

// BieterDelete deletes a bieter.
func (m Model) BieterDelete(id int) Event {
	return eventBieterDelete{ID: id}
}

// SetState deletes a bieter.
func (m Model) SetState(state ServiceState) Event {
	return eventStateSet{State: state}
}
