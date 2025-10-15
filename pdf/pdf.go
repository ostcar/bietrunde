package pdf

import (
	_ "embed" // for embedding
	"fmt"
	"strings"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/code"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	marotoConfig "github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/ostcar/bietrunde/model"
)

//go:embed pdf_header_image.png
var headerImage []byte

// Bietervertrag creates the bietervertrag pdf for a bieter
func Bietervertrag(domain string, bieter model.Bieter) ([]byte, error) {
	cfg := marotoConfig.NewBuilder().
		Build()

	m := maroto.New(cfg)

	abbuchungText := "Die Abbuchung erfolgt jeweils am ersten Werktag eines Monats von April 2026 bis März 2027."
	abbuchungBetrag := bieter.Gebot.String()
	monatlicherBetrag := bieter.Gebot.String()
	var abstandBetrag float64 = 5
	if bieter.Jaehrlich {
		abbuchungText = "Die Abbuchung erfolgt am 1. Werktag im April 2026."
		abbuchungBetrag = (bieter.Gebot * 12).String()
	}

	if bieter.Gebot.Empty() {
		abstandBetrag = 10
		monatlicherBetrag = ""
		abbuchungBetrag = ""
	}

	adresse := strings.ReplaceAll(bieter.Adresse, "\n", ", ")

	m.AddRows(
		// Header
		row.New(20).Add(
			col.New(6).Add(
				text.New("Solidarische Landwirtschaft Baarfood e. V", props.Text{Size: 10, Top: 3.5}),
				text.New("Neckarstrasse 120", props.Text{Size: 10, Top: 7}),
				text.New("78056 Villingen-Schwenningen", props.Text{Size: 10, Top: 10.5}),
				text.New("www.baarfood.de", props.Text{Size: 10, Top: 14}),
			),
			code.NewQrCol(3, fmt.Sprintf("%s?biet-id=%d", domain, bieter.ID)),
			image.NewFromBytesCol(3, headerImage, extension.Png, props.Rect{Center: true}),
		),

		// Gemüsevertrag
		text.NewRow(15, "Gemüsevertrag", props.Text{
			Size:  14,
			Style: fontstyle.Bold,
			Align: align.Center,
			Top:   5,
		}),

		// Vertragstext
		text.NewRow(10, fmt.Sprintf(`
			Ich, %s (E-Mail: %s ), bin Mitglied im des Vereins Solidarische Landwirtschaft Baarfood e.V.
			und möchte im Gemüsejahr 2026/2027 (April 2026 – März 2027) einen Gemüseanteil beziehen.`,
			bieter.Name(), bieter.Mail),
		),
		text.NewRow(35,
			`Der Gemüsevertrag gilt im oben genannten Zeitraum, daher für 12 Monate.
			Ich werde mein Gemüse wöchentlich an einer vorher festgelegten Verteilstelle abholen.
			Ich respektiere die in den Verteilstellen genannten Anteilsmengen und Abholfristen.
			Ich habe keinen Anspruch auf eine bestimmte Menge und Qualität der Produkte.
			Sollte es mir vorübergehend nicht möglich sein, meinen Pflichten (Abholung) nachzukommen,
			so sorge ich selbst in diesem Zeitraum für einen Ersatz. Im Falle einer Urlaubsvertretung weise
			ich persönlich in die Abholmodalitäten ein.
			Die endgültige Abgabe meines Anteils im laufenden Jahr ist nur möglich, wenn ein anderes
			Vereinsmitglied, das bisher keinen Ernteanteil bezieht oder ein neues Mitglied, den
			oben genannten monatlichen finanziellen Beitrag für die verbleibenden Monate übernimmt.
			Erst ab Abschluss des neuen Gemüsevertrages erfolgt der Lastschrifteinzug von diesem neuen Mitglied.`,
		),
		text.NewRow(abstandBetrag,
			fmt.Sprintf(`Ich hole meinen Anteil in der Verteilstelle in %s ab.`, bieter.Verteilstelle.String()),
		),
		text.NewRow(abstandBetrag,
			fmt.Sprintf("Mein Beitrag für den Gemüseanteil monatlich: %s", monatlicherBetrag),
			props.Text{Style: fontstyle.Bold},
		),
		text.NewRow(5, abbuchungText),
		text.NewRow(5,
			`Ist eine Abbuchung nicht möglich, so geht die Rückbuchungsgebühr zu meinen Lasten.`,
		),

		// Datum Unterschrift
		row.New(20).Add(
			col.New(6).Add(
				text.New("_________________________", props.Text{Top: 10, Size: 8}),
				text.New("Ort, Datum", props.Text{Top: 15}),
			),
			col.New(6).Add(
				text.New("_________________________", props.Text{Top: 10, Size: 8}),
				text.New("Unterschrift", props.Text{Top: 15}),
			),
		),
		row.New(20),

		// Sepa
		text.NewRow(15, "SEPA Lastschriftmandat", props.Text{
			Size:  14,
			Style: fontstyle.Bold,
			Align: align.Center,
			Top:   5,
		}),
		text.NewRow(5, `Gläubiger-Identifikationsnummer: DE62ZZZ00001997635`),
		text.NewRow(5, fmt.Sprintf(`Mandatsreferenz: 25%d`, bieter.ID)),
		text.NewRow(abstandBetrag, abbuchungText),
		text.NewRow(abstandBetrag, fmt.Sprintf("Der Betrag lautet: %s", abbuchungBetrag), props.Text{Style: fontstyle.Bold}),
		text.NewRow(15,
			`Ich ermächtige den Verein Solidarische Landwirtschaft Baarfood e.V.
			Lastschriften von meinem Konto einzuziehen. Zugleich weise ich mein
			Kreditinstitut an, die von Solidarische Landwirtschaft Baarfood e.V.
			auf mein Konto gezogenen Lastschriften einzulösen.`,
		),
		text.NewRow(10,
			`Ich kann innerhalb von acht Wochen, beginnend mit dem Belastungsdatum,
			die Erstattung des belasteten Betrages verlangen. Es gelten dabei die
			mit meinem Kreditinstitut vereinbarten Bedingungen.`,
		),

		text.NewRow(5, fmt.Sprintf(`Kontoinhaber: %s`, bieter.ShowKontoinhaber())),
		text.NewRow(5, fmt.Sprintf(`Adresse: %s`, adresse)),
		text.NewRow(5, fmt.Sprintf(`IBAN: %s`, bieter.IBAN)),

		// Datum Unterschrift
		row.New(20).Add(
			col.New(6).Add(
				text.New("_________________________", props.Text{Top: 10, Size: 8}),
				text.New("Ort, Datum", props.Text{Top: 15}),
			),
			col.New(6).Add(
				text.New("_________________________", props.Text{Top: 10, Size: 8}),
				text.New("Unterschrift Kontoinhaber", props.Text{Top: 15}),
			),
		),
	)

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate: %w", err)
	}

	return document.GetBytes(), nil
}
