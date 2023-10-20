package model

// Bieter is a person that makes an offer.
type Bieter struct {
	ID            int
	Name          string
	Teilpartner   string
	Mail          string
	Verteilstelle int
	Kontoinhaber  string
	Mitglied      bool
	Adresse       string
	IBAN          string
	Jaehrlich     bool
	Gebot         int
}

// Model of the service.
type Model struct {
	bieter map[int]Bieter
	state  ServiceState
}

// New returns an initialized model.
func New() Model {
	return Model{
		bieter: make(map[int]Bieter),
	}
}

// ServiceState is the state of the service.
type ServiceState int

const (
	stateInvalid ServiceState = iota
	stateRegistration
	stateValidation
	stateOffer
)

func (s ServiceState) String() string {
	return [...]string{"0 - Ungültig", "1 - Registrierung", "2 - Überprüfung", "3 - Gebote"}[s]
}
