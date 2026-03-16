package mastery

import "fmt"

// State is a value object representing the mastery level of an item.
type State string

const (
	Unknown State = "UNKNOWN"
	Fragile State = "FRAGILE"
	OK      State = "OK"
	Solid   State = "SOLID"
)

// Valid returns true if the state is one of the known values.
func (s State) Valid() bool {
	switch s {
	case Unknown, Fragile, OK, Solid:
		return true
	}
	return false
}

// ParseState converts a string to a State, returning an error if invalid.
func ParseState(s string) (State, error) {
	st := State(s)
	if !st.Valid() {
		return "", fmt.Errorf("mastery.ParseState: invalid state %q", s)
	}
	return st, nil
}
