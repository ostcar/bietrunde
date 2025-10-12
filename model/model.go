package model

import (
	"fmt"
	"math/rand"
	"net/mail"
	"strconv"
	"strings"

	"github.com/jbub/banking/iban"
)

// Gebot is an offer.
type Gebot int

func (g Gebot) String() string {
	if g == 0 {
		return "-"
	}
	euro := int(g) / 100
	cent := int(g) % 100
	euroStr := strReverse(strings.Join(strInGroup(strReverse(strconv.Itoa(euro)), 3), "."))
	return fmt.Sprintf("%s,%02d €", euroStr, cent)
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

// GebotFromString parses a string into an gebot.
func GebotFromString(str string) (Gebot, error) {
	euroStr, centStr, hasCent := strings.Cut(str, ".")
	euro, err := strconv.Atoi(euroStr)
	if err != nil {
		return 0, fmt.Errorf("euro has to be a number")
	}

	cent := 0
	if hasCent {

		cent, err = strconv.Atoi(centStr)
		if err != nil {
			return 0, fmt.Errorf("cent has to be a number")
		}
		if len(centStr) == 1 {
			cent *= 10
		}
	}
	return Gebot(euro*100 + cent), nil
}

// Empty is true, if there is no offer.
func (g Gebot) Empty() bool {
	return int(g) == 0
}

// Bieter is a person that makes an offer.
type Bieter struct {
	ID               int           `json:"id"`
	Vorname          string        `json:"vorname"`
	Nachname         string        `json:"nachname"`
	Mail             string        `json:"mail"`
	Adresse          string        `json:"adresse"`
	Telefon          string        `json:"telefon"`
	Mitglied         bool          `json:"mitglied"`
	Verteilstelle    Verteilstelle `json:"verteilstelle"`
	GanzOderHalb     GanzOderHalb  `json:"ganz_oder_halb"`
	Teilpartner      string        `json:"teilpartner"`
	IBAN             string        `json:"iban"`
	Kontoinhaber     string        `json:"kontoinhaber"`
	Jaehrlich        bool          `json:"jaehrlich"`
	Gebot            Gebot         `json:"gebot"`
	Anwesend         bool          `json:"anwesend"`
	CanSelfEdit      bool          `json:"can_edit"`
	ContractAccepted bool          `json:"contract_accepted"`
}

// Name returns the full name.
func (b Bieter) Name() string {
	if len(b.Vorname)+len(b.Nachname) == 0 {
		return ""
	}
	return b.Vorname + " " + b.Nachname
}

// ShowKontoinhaber the kontoinhaber
func (b Bieter) ShowKontoinhaber() string {
	if b.Kontoinhaber != "" {
		return b.Kontoinhaber
	}
	return b.Name()
}

// InvalidFields returns a map from invalid fields to an error message.
func (b Bieter) InvalidFields() map[string]string {
	invalid := make(map[string]string)
	if b.Vorname == "" {
		invalid["vorname"] = "Vorname darf nicht leer sein."
	}

	if b.Nachname == "" {
		invalid["nachname"] = "Nachname darf nicht leer sein."
	}

	if _, err := mail.ParseAddress(b.Mail); err != nil {
		if b.Mail == "" {
			invalid["mail"] = "E-Mail Adresse muss angegeben sein."
		} else {
			invalid["mail"] = "Dies ist keine gültige E-Mail-Adresse"
		}
	}

	if b.Adresse == "" {
		invalid["adresse"] = "Adresse darf nicht leer sein."
	}

	if !b.Mitglied {
		invalid["mitglied"] = "Du und der Teilnehmer müssen Mitglieder im Verein sein."
	}

	if b.Verteilstelle.ToAttr() == "-" {
		invalid["verteilstelle"] = "Wähle eine Verteilstelle aus."
	}

	if _, err := iban.Parse(strings.ReplaceAll(b.IBAN, " ", "")); err != nil {
		if b.IBAN == "" {
			invalid["iban"] = "Gebe eine IBAN an."
		} else {
			invalid["iban"] = "IBAN ist nicht gültig. Bitte überprüfe sie."
		}
	}
	return invalid
}

// BieterCSVHeader returns the header coresponding to Bieter.CSVRecord().
func BieterCSVHeader() []string {
	return []string{
		"id",
		"vorname",
		"nachname",
		"mail",
		"adresse",
		"mitglied",
		"verteilstelle",
		"teilpartner",
		"iban",
		"kontoinhaber",
		"abbuchung",
		"gebot-monat",
		"gebot-jaehrlich",
	}
}

// CSVRecord returns the bieter as csv record
func (b Bieter) CSVRecord() []string {
	mitglied := "Nein"
	if b.Mitglied {
		mitglied = "Ja"
	}
	abbuchung := "monatlich"
	if b.Jaehrlich {
		abbuchung = "jährlich"
	}
	gebotMonat := ""
	if !b.Gebot.Empty() && !b.Jaehrlich {
		gebotMonat = b.Gebot.String()
	}

	gebotJahr := ""
	if !b.Gebot.Empty() && b.Jaehrlich {
		gebotJahr = (b.Gebot * 12).String()
	}

	return []string{
		strconv.Itoa(b.ID),
		b.Vorname,
		b.Nachname,
		b.Mail,
		b.Adresse,
		mitglied,
		b.Verteilstelle.String(),
		b.Teilpartner,
		b.IBAN,
		b.ShowKontoinhaber(),
		abbuchung,
		gebotMonat,
		gebotJahr,
	}
}

func strInGroup(str string, size int) []string {
	var parts []string
	for len(str) > 0 {
		sep := min(len(str), size)
		parts = append(parts, str[:sep])
		str = str[sep:]
	}
	return parts
}

func strReverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func formatIBAN(iban string) string {
	withoutSpace := strings.ReplaceAll(iban, " ", "")
	parts := strInGroup(withoutSpace, 4)
	return strings.Join(parts, " ")
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
	id := rand.Intn(899_999_999) + 100_000_000
	return id, eventBieterCreate{ID: id}
}

// BieterUpdate updates all fields of a bieter.
func (m Model) BieterUpdate(bieter Bieter) Event {
	bieter.IBAN = formatIBAN(bieter.IBAN)
	return eventBieterUpdate{bieter}
}

// BieterDelete deletes a bieter.
func (m Model) BieterDelete(id int) Event {
	return eventBieterDelete{ID: id}
}

// BieterAcceptContract accepts a contract for a bieter.
func (m Model) BieterAcceptContract(bieterID int) Event {
	return eventAcceptContract{bieterID}
}

// SetState sets the service state.
func (m Model) SetState(state ServiceState) Event {
	return eventStateSet{State: state}
}

// SetGebot sets the gebot for an user.
func (m Model) SetGebot(bietID int, gebot Gebot) Event {
	return eventGebot{BietID: bietID, Gebot: gebot}
}

// ResetGebot sets all gebote to 0.
func (m Model) ResetGebot() Event {
	return eventResetGebot{}
}

// BieterSetAnwesend deletes a bieter.
func (m Model) BieterSetAnwesend(id int, anwesend bool) Event {
	return eventSetAnwesend{BietID: id, Anwesend: anwesend}
}

// BieterSetCanSelfEdit sets the attribute CanSelfEdit.
func (m Model) BieterSetCanSelfEdit(id int, canEdit bool) Event {
	return eventSetCanSelfEdit{BietID: id, CanSelfEdit: canEdit}
}
