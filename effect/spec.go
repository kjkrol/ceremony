package effect

import (
	"reflect"
	"time"
	"unsafe"

	"github.com/kjkrol/ceremony/tag"
)

// Spec is the definition of one effect: what it does to an entity and for how long.
type Spec []Trait

// Trait is one part of a Spec.
type Trait interface{ apply(d *Definition) }

// Definition is an effect as defined, what a store's binding runs: how long it lasts, what it
// grants and alters, what follows it.
type Definition struct {
	Name     string
	Lasts    time.Duration // 0: Forever
	Stacking bool
	Grants   []Granted
	Alters   []Alteration
	Next     uint8 // the effect Then casts, while Follows
	Follows  bool
	Shows    bool
	describe string
}

// Granted is tags of one family an effect gives while it runs: the family's component type,
// tag.Tags[F], and the bits.
type Granted struct {
	Family reflect.Type
	Bits   uint64
}

// Alteration is a change an effect makes to one component while it runs: Fn turns the value at
// its address.
type Alteration struct {
	Comp reflect.Type
	Fn   func(v unsafe.Pointer)
}

type traitFn func(d *Definition)

func (f traitFn) apply(d *Definition) { f(d) }

// Lasts is the default length of every cast of the effect; without it, an effect lasts until
// Dispel. CastFor overrides it per cast.
func Lasts(d time.Duration) Trait { return traitFn(func(e *Definition) { e.Lasts = d }) }

// Stacking makes a repeated Cast add a slot instead of refreshing the one there.
func Stacking() Trait { return traitFn(func(e *Definition) { e.Stacking = true }) }

// Described gives the effect a description for a UI, handed back by Effect.Description.
func Described(text string) Trait { return traitFn(func(e *Definition) { e.describe = text }) }

// Then casts next once the effect runs out, not when it is dispelled. next is one of the same
// Effects, defined before.
func Then(next Effect) Trait {
	return traitFn(func(e *Definition) { e.Next, e.Follows = next.id, true })
}

// Grant gives the entity these tags of family F while the effect runs, and takes them back after
// unless another running effect grants them; an entity without the family gets it attached.
func Grant[F any](tags ...tag.Tag[F]) Trait {
	bits := tag.Tags[F](0).With(tags...)
	return traitFn(func(e *Definition) {
		e.Grants = append(e.Grants, Granted{Family: reflect.TypeFor[tag.Tags[F]](), Bits: uint64(bits)})
	})
}

// Knob marks a component an effect may turn: one its owner only reads, so that an Alter ending
// puts back what it found. A type is one by its method Knob, a mark never called, confirmed
// where it is defined: var _ effect.Knob = Steering{}.
type Knob interface{ Knob() }

// Alter changes the entity's T while the effect runs; the original comes back after. Several
// running Alters of one T stack from the original in slot order, whichever ends first.
func Alter[T Knob](fn func(v *T)) Trait {
	return traitFn(func(e *Definition) {
		e.Alters = append(e.Alters, Alteration{Comp: reflect.TypeFor[T](),
			Fn: func(v unsafe.Pointer) { fn((*T)(v)) }})
	})
}
