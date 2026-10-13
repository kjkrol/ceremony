package plan

import "github.com/kjkrol/ceremony/internal/steps"

// The store's part of the plans: what runs them over goke, to move to the store binding as
// it is made. These are the engine's own types, one for goke and the saves, so they are
// aliases.

// Mind is the state of an entity's plan, a component the entity carries: which plan, the
// loops running since the last tick and what each keeps.
type Mind = steps.Mind

// Plans runs the plans of every entity with a Mind, once every step of the simulation.
type Plans = steps.Plans

// NewPlans is the plans on the clock now, drawing Chance from seed, casting fx and giving
// their commands to commands.
var NewPlans = steps.NewPlans

// PlanID is the id of the plan named name: the name hashed, what a Mind carries.
var PlanID = steps.PlanID

// MaxSteps is how many steps one plan holds at most.
const MaxSteps = steps.MaxSteps
