package goke

import (
	"encoding/gob"
	"reflect"
	"time"
	"unsafe"

	"github.com/kjkrol/ceremony/effect"
	"github.com/kjkrol/ceremony/tag"
	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

var _ goke.System = (*effectSystem)(nil)

var statesType = reflect.TypeFor[tag.Tags[effect.States]]()

const changedBits = uint64(1) << uint64(effect.Changed)

type effectSystem struct {
	reg       *effect.Effects
	originals *originals

	narrow, wide walk
	active       goke.Comp[Active]
	wideActive   goke.Comp[Wide]
	touched      map[reflect.Type]bool // scratch, reused
	thens        []uint8               // scratch, reused

	lookup       *goke.Query
	lookupActive goke.Comp[Active]
	lookupWide   *goke.Query
	lookupSlots  goke.Comp[Wide]
	activeID     goke.CompID
	cb           *goke.CmdBuf
	// Actives on their way to entities, until the next pass: a second cast before the first lands
	// joins it, the add assigned last carrying both.
	fresh map[uid.UID64]Active
	// families on their way to entities this pass, likewise joined.
	attaching map[attachment]uint64
}

type walk struct {
	query   *goke.Query
	states  *goke.OptColumn
	columns map[reflect.Type]*goke.OptColumn
	ids     map[reflect.Type]goke.CompID // the families', to attach
}

func (w *walk) bind(si *goke.SysInit, reg *effect.Effects, slots goke.Trackable) {
	qb := si.NewQueryBuilder(slots)
	w.columns, w.ids = map[reflect.Type]*goke.OptColumn{}, map[reflect.Type]goke.CompID{}
	w.states = w.column(si, qb, statesType, true)
	for k := range reg.Count() {
		d := reg.Definition(uint8(k))
		for _, g := range d.Grants {
			w.column(si, qb, g.Family, true)
		}
		for _, a := range d.Alters {
			w.column(si, qb, a.Comp, false)
		}
	}
	w.query = qb.Build()
}

func (w *walk) column(si *goke.SysInit, qb *goke.QueryBuilder, t reflect.Type, family bool) *goke.OptColumn {
	if col, ok := w.columns[t]; ok {
		return col
	}
	col := goke.NewOptColumn(t)
	qb.Optional(col)
	w.columns[t] = col
	if family {
		w.ids[t] = si.RegCompOf(t)
	}
	return col
}

// the i-th value of col on the block, which is present
func at(col *goke.OptColumn, cursor *goke.Cursor, i int) unsafe.Pointer {
	return unsafe.Add(col.Base(cursor), uintptr(i)*col.Size())
}

func word(col *goke.OptColumn, cursor *goke.Cursor, i int) *uint64 {
	return (*uint64)(at(col, cursor, i))
}

type attachment struct {
	id     uid.UID64
	family reflect.Type
}

func newEffectSystem(reg *effect.Effects) *effectSystem {
	return &effectSystem{reg: reg, originals: newOriginals(), touched: map[reflect.Type]bool{},
		fresh: map[uid.UID64]Active{}, attaching: map[attachment]uint64{}}
}

func (s *effectSystem) Init(si *goke.SysInit) {
	s.reg.Freeze()
	for k := range s.reg.Count() {
		for _, a := range s.reg.Definition(uint8(k)).Alters {
			gob.Register(reflect.New(a.Comp).Elem().Interface())
		}
	}
	s.narrow.bind(si, s.reg, &s.active)
	s.wide.bind(si, s.reg, &s.wideActive)
	s.lookup = si.NewQueryBuilder(&s.lookupActive).Build()
	s.lookupWide = si.NewQueryBuilder(&s.lookupSlots).Build()
	s.activeID = si.RegComp[Active]()
}

func (s *effectSystem) Update(cb *goke.CmdBuf, d time.Duration) {
	s.cb = cb
	clear(s.fresh)
	clear(s.attaching)
	for s.narrow.query.All(); s.narrow.query.Next(); {
		cursor := s.narrow.query.Cursor()
		actives := s.active.Slice(cursor)
		for i, id := range cursor.IDs {
			s.advance(&s.narrow, cursor, i, id, actives[i].Slots[:], d)
		}
	}
	for s.wide.query.All(); s.wide.query.Next(); {
		cursor := s.wide.query.Cursor()
		wides := s.wideActive.Slice(cursor)
		for i, id := range cursor.IDs {
			s.advance(&s.wide, cursor, i, id, wides[i].Slots[:], d)
		}
	}
}

func (s *effectSystem) advance(w *walk, cursor *goke.Cursor, i int, id uid.UID64, slots []effectSlot, d time.Duration) {
	changed := false
	if !empty(slots) {
		changed = s.step(w, cursor, i, id, slots, d)
	}
	s.mark(w, cursor, i, id, changed)
}

func (s *effectSystem) mark(w *walk, cursor *goke.Cursor, i int, id uid.UID64, changed bool) {
	if !w.states.Present(cursor) {
		if changed {
			s.attach(w, id, statesType, changedBits)
		}
		return
	}
	if changed {
		*word(w.states, cursor, i) |= changedBits
	} else {
		*word(w.states, cursor, i) &^= changedBits
	}
}

func (s *effectSystem) step(w *walk, cursor *goke.Cursor, i int, id uid.UID64, slots []effectSlot, d time.Duration) bool {
	touched := s.touched
	clear(touched)
	s.thens = s.thens[:0]
	shown := false
	for k := range slots {
		slot := &slots[k]
		switch slot.State {
		case slotPending:
			if s.begin(w, cursor, i, id, slot, touched) {
				slot.State = slotRunning
				shown = shown || s.reg.Definition(slot.Kind).Shows
			}
		case slotRunning:
			if slot.Left != effect.Forever {
				slot.Left -= d
				if slot.Left <= 0 {
					shown = shown || s.reg.Definition(slot.Kind).Shows
					s.end(w, cursor, i, slots, k, touched)
				}
			}
		}
	}
	// what Then casts begins in the pass its forerunner ran out, its time counted from the next
	for _, next := range s.thens {
		s.queue(slots, next, lastsOf(s.reg.Definition(next)))
		if k := slotOf(slots, next); k >= 0 && slots[k].State == slotPending &&
			s.begin(w, cursor, i, id, &slots[k], touched) {
			slots[k].State = slotRunning
			shown = shown || s.reg.Definition(next).Shows
		}
	}
	for comp := range touched {
		s.recompute(w, cursor, i, id, slots, comp)
	}
	if empty(slots) {
		s.originals.forget(id)
	}
	return len(touched) > 0 || shown
}

func (s *effectSystem) begin(w *walk, cursor *goke.Cursor, i int, id uid.UID64, slot *effectSlot, touched map[reflect.Type]bool) bool {
	d := s.reg.Definition(slot.Kind)
	ready := true
	for k, g := range d.Grants {
		if w.columns[g.Family].Present(cursor) || attachedBefore(d.Grants[:k], g.Family) {
			continue
		}
		s.attach(w, id, g.Family, g.Bits)
		ready = false
	}
	if !ready {
		return false
	}
	for _, g := range d.Grants {
		*word(w.columns[g.Family], cursor, i) |= g.Bits
	}
	for _, a := range d.Alters {
		if w.columns[a.Comp].Present(cursor) {
			touched[a.Comp] = true
		}
	}
	return true
}

// a family is attached as one component, the bits of every attach this pass joined
func (s *effectSystem) attach(w *walk, id uid.UID64, family reflect.Type, bits uint64) {
	key := attachment{id: id, family: family}
	bits |= s.attaching[key]
	s.attaching[key] = bits
	*(*uint64)(s.cb.Reserve(id, w.ids[family])) = bits
}

func attachedBefore(grants []effect.Granted, family reflect.Type) bool {
	for _, g := range grants {
		if g.Family == family {
			return true
		}
	}
	return false
}

func (s *effectSystem) end(w *walk, cursor *goke.Cursor, i int, slots []effectSlot, k int, touched map[reflect.Type]bool) {
	d := s.reg.Definition(slots[k].Kind)
	if d.Follows && !slots[k].Dispelled {
		s.thens = append(s.thens, d.Next)
	}
	slots[k] = effectSlot{}
	for _, g := range d.Grants {
		bits := g.Bits
		for _, other := range slots {
			if other.State != slotRunning {
				continue
			}
			for _, og := range s.reg.Definition(other.Kind).Grants {
				if og.Family == g.Family {
					bits &^= og.Bits
				}
			}
		}
		if col := w.columns[g.Family]; col.Present(cursor) {
			*word(col, cursor, i) &^= bits
		}
	}
	for _, alt := range d.Alters {
		if w.columns[alt.Comp].Present(cursor) {
			touched[alt.Comp] = true
		}
	}
}

func (s *effectSystem) recompute(w *walk, cursor *goke.Cursor, i int, id uid.UID64, slots []effectSlot, comp reflect.Type) {
	ptr := at(w.columns[comp], cursor, i)
	value := reflect.NewAt(comp, ptr).Elem()
	if original, ok := s.originals.get(id, comp); ok {
		value.Set(reflect.ValueOf(original))
	} else {
		s.originals.put(id, comp, value.Interface())
	}
	applied := false
	for _, slot := range slots {
		if slot.State != slotRunning {
			continue
		}
		for _, alt := range s.reg.Definition(slot.Kind).Alters {
			if alt.Comp == comp {
				alt.Fn(ptr)
				applied = true
			}
		}
	}
	if !applied {
		s.originals.drop(id, comp)
	}
}

func (s *effectSystem) cast(id uid.UID64, kind uint8, left time.Duration) {
	if slots := s.slotsOf(id); slots != nil {
		s.queue(slots, kind, left)
		return
	}
	a := s.fresh[id]
	s.queue(a.Slots[:], kind, left)
	s.fresh[id] = a
	s.cb.AddOne(id, s.activeID, a)
}

// with no slot free nothing happens
func (s *effectSystem) queue(slots []effectSlot, kind uint8, left time.Duration) {
	if k := slotOf(slots, kind); k >= 0 && !s.reg.Definition(kind).Stacking {
		slots[k].Left, slots[k].Dispelled = left, false
		return
	}
	if k := free(slots); k >= 0 {
		slots[k] = effectSlot{Kind: kind, Left: left, State: slotPending}
	}
}

func (s *effectSystem) dispel(id uid.UID64, kind uint8) {
	slots := s.slotsOf(id)
	for k := range slots {
		if slots[k].State != slotEmpty && slots[k].Kind == kind {
			slots[k].Left, slots[k].Dispelled = 0, true
			if slots[k].State == slotPending {
				slots[k] = effectSlot{}
			}
		}
	}
}

func (s *effectSystem) has(id uid.UID64, kind uint8) bool { return has(s.slotsOf(id), kind) }

func (s *effectSystem) slotsOf(id uid.UID64) []effectSlot {
	if s.lookup.Seek(id) {
		return s.lookupActive.At(s.lookup.Cursor()).Slots[:]
	}
	if s.lookupWide.Seek(id) {
		return s.lookupSlots.At(s.lookupWide.Cursor()).Slots[:]
	}
	return nil
}
