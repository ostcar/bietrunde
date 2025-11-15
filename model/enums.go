package model

// Verteilstelle is an enum type
type Verteilstelle int

// Different enum values
const (
	VerteilstelleNone Verteilstelle = iota
	VerteilstelleVillingen
	VerteilstelleSchwenningen
	VerteilstelleUeberauchen
)

func (v Verteilstelle) String() string {
	switch v {
	case VerteilstelleVillingen:
		return "Villingen"
	case VerteilstelleSchwenningen:
		return "Schwenningen"
	case VerteilstelleUeberauchen:
		return "Überauchen"
	default:
		return "-"
	}
}

// VerteilstelleFromAttr converts a attribute string representation to
// Verteilstelle.
func VerteilstelleFromAttr(attr string) Verteilstelle {
	switch attr {
	case "villingen":
		return VerteilstelleVillingen
	case "schwenningen":
		return VerteilstelleSchwenningen
	case "ueberauchen":
		return VerteilstelleUeberauchen
	default:
		return VerteilstelleNone
	}
}

// ToAttr converts the Verteilstelle to a simple string that can be used in select statements.
func (v Verteilstelle) ToAttr() string {
	switch v {
	case VerteilstelleVillingen:
		return "villingen"
	case VerteilstelleSchwenningen:
		return "schwenningen"
	case VerteilstelleUeberauchen:
		return "ueberauchen"
	default:
		return "-"
	}
}

// GanzOderHalb is an enum to tell, if the bieter wants to share his anteil.
type GanzOderHalb int

// These are the enum values for GanzOderHalb.
const (
	GanzerAnteil GanzOderHalb = iota
	HalberAnteilSuche
	HalberAnteil
	HalberAnteilMoeglich
)

func (v GanzOderHalb) String() string {
	return v.ToAttr()
}

// GanzOderHalbFromAttr converts a attribute string representation to
// GanzOderHalb.
func GanzOderHalbFromAttr(attr string) GanzOderHalb {
	switch attr {
	case "ganz":
		return GanzerAnteil
	case "halb-suche":
		return HalberAnteilSuche
	case "halb":
		return HalberAnteil
	case "halb-moeglich":
		return HalberAnteilMoeglich
	default:
		return GanzerAnteil
	}
}

// ToAttr converts a GanzOderHalb enum value to a string representation.
func (v GanzOderHalb) ToAttr() string {
	switch v {
	case GanzerAnteil:
		return "ganz"
	case HalberAnteilSuche:
		return "halb-suche"
	case HalberAnteil:
		return "halb"
	case HalberAnteilMoeglich:
		return "halb-moeglich"
	default:
		return "ganz"
	}
}

// ServiceState is the state of the service.
type ServiceState int

// States of the service.
const (
	StateInvalid ServiceState = iota
	StateRegistration
	StateValidation
	StateOffer
	StateFinish
)

// States returns all possible states
func States() []ServiceState {
	return []ServiceState{
		StateRegistration,
		StateValidation,
		StateOffer,
		StateFinish,
	}
}

func (s ServiceState) String() string {
	all := [...]string{
		"Ungültig",
		"Registrierung",
		"Überprüfung",
		"Gebote",
		"Fertig",
	}
	if int(s) >= len(all) {
		s = 0
	}
	return all[s]
}

// ToAttr converts to a attribute for select elements
func (s ServiceState) ToAttr() string {
	switch s {
	case StateRegistration:
		return "registration"
	case StateValidation:
		return "validation"
	case StateOffer:
		return "offer"
	case StateFinish:
		return "finish"
	default:
		return "-"
	}
}

// StateFromAttr converts a attribute string representation to
// ServiceState.
func StateFromAttr(attr string) ServiceState {
	switch attr {
	case "registration":
		return StateRegistration
	case "validation":
		return StateValidation
	case "offer":
		return StateOffer
	case "finish":
		return StateFinish
	default:
		return StateInvalid
	}
}
