package effect

import (
	"fmt"
	"time"

	"github.com/kjkrol/ceremony/tag"
	"github.com/kjkrol/uid"
)

// Forever is the length of a cast that lasts until Dispel.
const Forever = time.Duration(-1)

// Effects is a register of effects: their definitions, by name, which a store's binding runs.
type Effects struct {
	defs   []Definition
	name   func(name string) tag.Tag[States]
	frozen bool

	byName map[string]uint8
	marks  []tag.Tag[States] // by effect
	guard  func(name string)
}

// New makes the register, naming each effect's marker with name, which must give Changed
// first.
func New(name func(name string) tag.Tag[States]) *Effects {
	if t := name(changedName); t != Changed {
		panic(fmt.Sprintf("effects: the markers have tags of their own before %q", changedName))
	}
	return &Effects{name: name}
}

const markerPrefix = "effect."

// Define registers an effect under name, with its own marker on while it runs (Effect.Mark). It
// panics once the register is frozen, or past 63 effects.
func (e *Effects) Define(name string, spec Spec) {
	if e.guard != nil {
		e.guard(name)
	}
	if e.frozen {
		panic(fmt.Sprintf("effects: %q defined after the game was set up", name))
	}
	if len(e.defs) == tag.MaxTagsPerFamily-1 {
		panic(fmt.Sprintf("effects: %q is one too many: at most %d effects, each with its marker", name, tag.MaxTagsPerFamily-1))
	}
	mark := e.name(markerPrefix + name)
	if mark == Changed {
		panic(fmt.Sprintf("effects: %q is named as Changed is", name))
	}
	d := Definition{Name: name}
	Grant(mark).apply(&d)
	for _, t := range spec {
		t.apply(&d)
	}
	e.defs = append(e.defs, d)
	e.marks = append(e.marks, mark)
	id := uint8(len(e.defs) - 1)
	if e.byName == nil {
		e.byName = map[string]uint8{}
	}
	e.byName[name] = id
}

// Named is the effect defined as name. An unknown name panics.
func (e *Effects) Named(name string) Effect {
	ef, ok := e.Lookup(name)
	if !ok {
		panic(fmt.Sprintf("effects: no effect is defined as %q", name))
	}
	return ef
}

// Lookup is Named for a name that may be unknown: false for one.
func (e *Effects) Lookup(name string) (Effect, bool) {
	id, ok := e.byName[name]
	if !ok {
		return Effect{}, false
	}
	return Effect{owner: e, id: id, mark: e.marks[id]}, true
}

// Guard has Define call guard with each effect's name first.
func (e *Effects) Guard(guard func(name string)) { e.guard = guard }

// Freeze closes the register: the binding calls it as it is set up, and a Define after panics.
func (e *Effects) Freeze() { e.frozen = true }

// Count is how many effects are defined.
func (e *Effects) Count() int { return len(e.defs) }

// Definition is the definition of the effect with the id, for the binding; Shows may turn on
// after the binding is set up, so it reads the definition where it stands.
func (e *Effects) Definition(id uint8) *Definition { return &e.defs[id] }

// Lasts is how long a cast of effect lasts by its Spec: Forever without a Lasts.
func (e *Effects) Lasts(effect Effect) time.Duration {
	if d := e.defs[effect.id].Lasts; d > 0 {
		return d
	}
	return Forever
}

// Effect is one effect defined with its Effects: what a rule or a plan casts.
type Effect struct {
	owner *Effects
	id    uint8
	mark  tag.Tag[States]
}

// ID is the effect's place in its register, what a store's binding keeps on an entity.
func (e Effect) ID() uint8 { return e.id }

// Shows says something is drawn by the effect's marker: the entity's Changed goes on as the
// effect begins and ends, though it alters nothing.
func (e Effect) Shows() { e.owner.defs[e.id].Shows = true }

// Mark is the effect's own marker, on while it runs.
func (e Effect) Mark() tag.Tag[States] { return e.mark }

// Description is what the Spec's Described said of the effect, "" without one.
func (e Effect) Description() string { return e.owner.defs[e.id].describe }

// Runtime is the effects as a store's binding runs them on its entities: what a rule's steps,
// a plan's and a plugin's pass cast and ask through.
type Runtime interface {
	// Has reports whether id is under effect.
	Has(id uid.UID64, effect Effect) bool
	// Cast puts effect on id for as long as its Spec says, landing with the effects' next pass;
	// cast again, it is refreshed unless it stacks.
	Cast(id uid.UID64, effect Effect)
	// CastFor is Cast for a set time of the clock's; Forever for one that lasts until Dispel.
	CastFor(id uid.UID64, effect Effect, d time.Duration)
	// Dispel ends effect on id with the effects' next pass.
	Dispel(id uid.UID64, effect Effect)
}
