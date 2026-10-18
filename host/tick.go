package host

import (
	"time"

	"github.com/kjkrol/ceremony/command"
	"github.com/kjkrol/ceremony/effect"
	"github.com/kjkrol/uid"
)

// Tick is what a host tells the rules it runs about its pass.
type Tick struct {
	Now      time.Time // read once for the whole pass
	Dt       time.Duration
	Commands *command.Carrier
	Effects  effect.Runtime
	Time     time.Duration // game time at the step's end
	Seed     uint64
	World    uid.UID64 // the world's own entity
	// Places tells each the places p are of the moment, a pointer to it; nil where there are
	// none.
	Places func(moment any, p Places, each func(uid.UID64))
	// Roles tells the roles an entity plays, a bit each; nil where there are none.
	Roles func(id uid.UID64) uint64
}

// TickSource builds the Tick of a pass.
type TickSource func(dt time.Duration) Tick

// Of is the Tick of a pass over dt; with no source, one of no world.
func (f TickSource) Of(dt time.Duration) Tick {
	if f == nil {
		return Tick{Now: time.Now(), Dt: dt}
	}
	return f(dt)
}

const maxFamilies = 8

// Marks is which tags of a host's families one entity carries, as PairRules reads them.
type Marks struct{ words [maxFamilies]uint64 }
