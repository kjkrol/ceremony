package host

import "fmt"

// Placing is a kind of places the hosts of a library tell for their moments (Tick.Places), named
// for the messages: Takes says of a moment whether it has them.
type Placing struct {
	name  string
	takes func(moment any) bool
}

// NewPlacing is the kind of places name, of the moments takes reports, each a pointer to the
// moment a rule fires on.
func NewPlacing(name string, takes func(moment any) bool) *Placing {
	return &Placing{name: name, takes: takes}
}

func (p *Placing) String() string { return p.name }

// Takes reports whether moment has places of the kind.
func (p *Placing) Takes(moment any) bool { return p.takes(moment) }

// Places are those of a Placing within N of the moment: whom cond.At and act.At turn a rule to,
// a branch run once a place.
type Places struct {
	Placing *Placing
	N       int
}

func (p Places) String() string { return fmt.Sprintf("%s(%d)", p.Placing, p.N) }
