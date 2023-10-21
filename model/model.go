package model

import "math/rand"

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
	Gebot         int           `json:"gebot"`
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

// ServiceState is the state of the service.
type ServiceState int

// States of the service.
const (
	StateInvalid ServiceState = iota
	StateRegistration
	StateValidation
	StateOffer
)

func (s ServiceState) String() string {
	return [...]string{"0 - Ungültig", "1 - Registrierung", "2 - Überprüfung", "3 - Gebote"}[s]
}
