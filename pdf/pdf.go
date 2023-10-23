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

	abbuchung := "Monatlich"
	abbuchungText := "Die Abbuchung erfolgt am ersten Werktag eines Monats von April 2024 bis März 2025"
	abbuchungBetrag := bieter.Gebot.String()
	if bieter.Jaehrlich {
		abbuchung = "Jährlich"
		abbuchungText = "Die Abbuchung erfolgt am 1. Werktag im April 2024"
		abbuchungBetrag = (bieter.Gebot * 12).String()
	}

	if bieter.Gebot.Empty() {
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
			und möchte im Gemüsejahr 2024/25 (April 2024 – März 2025) einen Gemüseanteil beziehen.`,
			bieter.Name(), bieter.Mail),
		),
		text.NewRow(10,
			`Nach erfolgreicher Bieterrunde schließe ich mit dem Verein Solidarische Landwirtschaft 
			Baarfood e.V. diesen Gemüsevertrag ab.`,
		),
		text.NewRow(40,
			`Die Gemüsevertrag gilt von April 2024 bis März 2025 (=12 Monate). 
			Ich kann mein Gemüse wöchentlich an einer vorher festgelegten Verteilstelle abholen. 
			Ich respektiere die in den Verteilstellen genannten Anteilsmengen und Abholfristen. 
			Ich habe keinen Anspruch auf eine bestimmte Menge und Qualität der Produkte. 
			Sollte es mir vorübergehend nicht möglich sein, meinen Pflichten (Abholung) nach zu kommen, 
			so sorge ich selbst in diesem Zeitraum für einen Ersatz. Im Falle einer Urlaubsvertretung weise 
			ich persönlich in die Abholmodalitäten ein. Ein finanzieller Ausgleich wird privat organisiert. 
			Die endgültige Abgabe meines Anteils im laufenden Jahr ist nur möglich, wenn ein anderes 
			Vereinsmitglied, das bisher keinen Ernteanteil bezieht, oder ein neues Mitglied, den 
			oben genannten monatlichen finanziellen Beitrag für die verbleibenden Monate übernimmt. 
			Erst ab diesem Zeitpunkt erfolgt der Lastschrifteinzug von diesem neuen Mitglied.`,
		),
		text.NewRow(5,
			fmt.Sprintf(`Ich hole meinen Anteil in der Verteilstelle in %s ab.`, bieter.Verteilstelle.String()),
		),
		text.NewRow(5,
			fmt.Sprintf(`Die Abbuchung meines Beitrages für den Ernteanteil erfolgt von April 2024 bis März 2025 %s.`, abbuchung),
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

		// Sepa
		text.NewRow(15, "SEPA Lastschriftmandat", props.Text{
			Size:  14,
			Style: fontstyle.Bold,
			Align: align.Center,
			Top:   5,
		}),
		text.NewRow(5, `Gläubiger-Identifikationsnummer: DE62ZZZ00001997635`),
		text.NewRow(5, fmt.Sprintf(`Mandatsreferenz: 24%d`, bieter.ID)),
		text.NewRow(10, abbuchungText),
		text.NewRow(5, fmt.Sprintf("Der Betrag lautet: %s", abbuchungBetrag), props.Text{Style: fontstyle.Bold}),
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
		text.NewRow(5,
			`Ist eine Abbuchung nicht möglich, so geht die Rückbuchungsgebühr zu meinen Lasten.`,
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