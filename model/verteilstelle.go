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
