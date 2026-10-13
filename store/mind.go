package store

import "time"

// MaxSteps is how many steps one plan holds at most.
const MaxSteps = 128

// Mind is the state of an entity's plan, a component the entity carries: which plan, the steps
// running since the last tick and what each keeps. The fields are exported and of a fixed size
// for the store.
type Mind struct {
	Plan    uint64 // the name hashed
	Running StepSet
	Slot    [MaxSteps]uint8
	Since   [MaxSteps]time.Duration // on the clock of the plans
}

// StepSet is a set of a plan's steps, a bit each.
type StepSet [MaxSteps / 64]uint64
