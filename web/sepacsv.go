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

var abbuchDaten = []string{
	"01.04.2026",
	"04.05.2026",
	"01.06.2026",
	"01.07.2026",
	"03.08.2026",
	"01.09.2026",
	"01.10.2026",
	"02.11.2026",
	"01.12.2026",
	"04.01.2027",
	"01.02.2027",
	"01.03.2027",
}

func writeSEPACSVToZip(w io.Writer, bieter []model.Bieter) error {
	if _, err := w.Write([]byte(SEPACSVHeader)); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, b := range bieter {
		if b.Gebot.Empty() {
			continue
		}

		if b.Jaehrlich {
			fmt.Fprintf(w, SEPACSVLine, abbuchDaten[0], b.ShowKontoinhaber(), b.IBAN, b.Gebot.ForYear().NumberString(), b.ID)
			continue
		}

		for i := range abbuchDaten {
			fmt.Fprintf(w, SEPACSVLine, abbuchDaten[i], b.ShowKontoinhaber(), b.IBAN, b.Gebot.NumberString(), b.ID)
		}
	}

	return nil
}
