package host

import (
	"fmt"

	"github.com/kjkrol/ceremony/store"
)

type eachRule[P any] interface {
	BindColumns(cols store.Columns)
	RunEach(t Tick, row store.Row, n int, keep func(i int) bool, about func(i int) P)
}

type onceRule[P any] interface {
	RunOnce(t Tick, about P)
}

// Rules are the rules of a moment P a pass runs over the entities it walks anyway: Bind them
// to the columns of its query as it is built, Run them over each block of entities.
type Rules[P any] struct {
	rules []eachRule[P]
	bound bool
}

// Empty reports whether no rule was added.
func (r *Rules[P]) Empty() bool { return len(r.rules) == 0 }

// Add takes a rule of P; ErrUnhosted for another, ErrHostBuilt after Bind.
func (r *Rules[P]) Add(rule any) error {
	rule = hosted(rule)
	each, ok := rule.(eachRule[P])
	if !ok {
		return fmt.Errorf("%w: %v", ErrUnhosted, rule)
	}
	if r.bound {
		return fmt.Errorf("%w: %v", ErrHostBuilt, rule)
	}
	r.rules = append(r.rules, each)
	return nil
}

// Bind binds what the rules read to the columns of the pass's query. Call it once, as the
// query is built.
func (r *Rules[P]) Bind(cols store.Columns) {
	r.bound = true
	for _, each := range r.rules {
		each.BindColumns(cols)
	}
}

// Run runs every rule over a block of n entities; about(i) is the moment of its i-th entity.
func (r *Rules[P]) Run(t Tick, row store.Row, n int, about func(i int) P) {
	r.RunWhere(t, row, n, nil, about)
}

// RunWhere is Run over the entities of the block keep lets through: a moment only some have.
func (r *Rules[P]) RunWhere(t Tick, row store.Row, n int, keep func(i int) bool, about func(i int) P) {
	for _, each := range r.rules {
		each.RunEach(t, row, n, keep, about)
	}
}

// StepRules are the rules of a moment P of the world as a whole, run once a step and walking no
// entities. A role's fire while the entity the moment is About plays the role (Tick.Roles).
type StepRules[P any] struct {
	rules []stepRule[P]
	bound bool
}

type stepRule[P any] struct {
	once  onceRule[P]
	roles uint64
}

type playedRule interface{ OnceRoles() uint64 }

// Add takes a rule of P asking nothing of the entity; ErrUnhosted for another, ErrHostBuilt
// after Bind.
func (r *StepRules[P]) Add(rule any) error {
	rule = hosted(rule)
	once, ok := rule.(onceRule[P])
	if !ok {
		return fmt.Errorf("%w: %v (a rule of the world as a whole takes no filter)", ErrUnhosted, rule)
	}
	if r.bound {
		return fmt.Errorf("%w: %v", ErrHostBuilt, rule)
	}
	s := stepRule[P]{once: once}
	if p, ok := rule.(playedRule); ok {
		s.roles = p.OnceRoles()
	}
	r.rules = append(r.rules, s)
	return nil
}

// Bind marks the system running the rules built, so no rule is taken after. Call it in its Init.
func (r *StepRules[P]) Bind() { r.bound = true }

// Empty reports whether no rule was added.
func (r *StepRules[P]) Empty() bool { return len(r.rules) == 0 }

// Run runs every rule on about.
func (r *StepRules[P]) Run(t Tick, about P) {
	for _, s := range r.rules {
		if s.roles != 0 && !plays(t, &about, s.roles) {
			continue
		}
		s.once.RunOnce(t, about)
	}
}

func plays(t Tick, moment any, roles uint64) bool {
	a, ok := moment.(About)
	return ok && t.Roles != nil && t.Roles(a.Who())&roles == roles
}
