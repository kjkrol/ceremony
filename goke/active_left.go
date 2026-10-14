package goke

import (
	"time"

	"github.com/kjkrol/ceremony/effect"
)

// Left is what remains of e on the entity: Forever for one without an end, false while it is
// not under e.
func (a *Active) Left(e effect.Effect) (time.Duration, bool) { return left(a.Slots[:], e) }

// Left is what remains of e on the entity: Forever for one without an end, false while it is
// not under e.
func (w *Wide) Left(e effect.Effect) (time.Duration, bool) { return left(w.Slots[:], e) }

func left(slots []effectSlot, e effect.Effect) (time.Duration, bool) {
	k := slotOf(slots, e.ID())
	if k < 0 {
		return 0, false
	}
	return slots[k].Left, true
}
