package customer

// Status represents an active or not customer
type Status string

const (
	Activated    = "Activated"
	NotActivated = "NotActivated"
)

// String returns a string representation of the active
func (a Status) String() string {
	switch a {
	case Activated:
		return "Activated"
	case NotActivated:
		return "Not activated"
	}
	return string(a)
}
