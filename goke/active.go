package goke

import (
	"time"

	"github.com/kjkrol/ceremony/effect"
	"github.com/kjkrol/ceremony/tag"
)

const maxEffects = 8

const wideEffects = tag.MaxTagsPerFamily - 1

// Active is what an entity is under: up to eight effects, saved with the entity.
type Active struct {
	Slots [maxEffects]effectSlot
}

type effectSlot struct {
	Kind      uint8
	Left      time.Duration
	State     slotState
	Dispelled bool // its Then is not cast
}

type slotState uint8

const (
	slotEmpty   slotState = iota
	slotPending           // cast, not yet applied
	slotRunning
)

// Wide is an Active with a slot for every effect a game may define: what a plugin's own entity
// is under, saved with it.
type Wide struct {
	Slots [wideEffects]effectSlot
}

func has(slots []effectSlot, kind uint8) bool { return slotOf(slots, kind) >= 0 }

func slotOf(slots []effectSlot, kind uint8) int {
	for i, s := range slots {
		if s.State != slotEmpty && s.Kind == kind {
			return i
		}
	}
	return -1
}

func free(slots []effectSlot) int {
	for i, s := range slots {
		if s.State == slotEmpty {
			return i
		}
	}
	return -1
}

func empty(slots []effectSlot) bool {
	for _, s := range slots {
		if s.State != slotEmpty {
			return false
		}
	}
	return true
}

func lastsOf(d *effect.Definition) time.Duration {
	if d.Lasts > 0 {
		return d.Lasts
	}
	return effect.Forever
}
