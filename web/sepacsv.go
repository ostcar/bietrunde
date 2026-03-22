package web

import (
	"fmt"
	"io"

	"github.com/ostcar/bietrunde/model"
)

const SEPACSVHeader = `Basislastschrift;;;;;;;;;;;;;;;;;;;;
Fälligkeitstermin;Zahlungspflichtiger;Straße;Gebäude Nr.;Postleitzahl;Stadt;Länder-Kennzeichen;Land;IBAN Zahlungspflichtiger;BIC;Bei Kreditinstitut;Betrag;Verwendungszweck 1;Verwendungszweck 2;IBAN des Zahlungsempfängers;Zahlungsempfänger;Abweichender Zahlungsempfänger;Mandatsreferenz;Unterschrieben am;Ausführungsart;Gläubiger ID des Zahlungsempfängers
`

const SEPACSVLine = `%s;%s;;;;;;;%s;;;%s;Gemüseanteil 26.27 Baarfood;;DE24643901300278501001;Solidarische Landwirtschaft;;25%d;26.11.2025;einmalig;DE62ZZZ00001997635
`

const abbuchung = "02.04.2026"

func writeSEPACSVToZip(w io.Writer, bieter []model.Bieter, jaehrlich bool) error {
	if _, err := w.Write([]byte(SEPACSVHeader)); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, b := range bieter {
		if b.Gebot.Empty() {
			continue
		}

		if jaehrlich && b.Jaehrlich {
			fmt.Fprintf(w, SEPACSVLine, abbuchung, b.ShowKontoinhaber(), b.IBANTrimed(), b.Gebot.ForYear().NumberString(), b.ID)
			continue
		}

		if !jaehrlich && !b.Jaehrlich {
			fmt.Fprintf(w, SEPACSVLine, abbuchung, b.ShowKontoinhaber(), b.IBANTrimed(), b.Gebot.NumberString(), b.ID)
		}
	}

	return nil
}
