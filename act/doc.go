// Package act is what a rule does: the Then of a Branch, every step said of whom it is done
// by or to, and the Branch itself, which given to a role is a rule.
//
//	act.Branch[Review]{
//		Name: "escalate",
//		When: cond.Self.If(overdue).And(cond.Self.If(cond.Not(escalated))),
//		Then: act.Steps(act.Self.Put(escalated), act.Self.Do(Notify{To: "lead"})),
//	}
//
// # Steps
//
// [Self], [Other] and [At] are the subjects; their methods are the steps: [Subject.Put],
// [Subject.Take] or [Subject.Hold] an effect, [Subject.Do] a command. [Steps] runs its parts
// in order, [OneOf] the first that does not fail; both take steps and branches.
//
// # Branches
//
// A [Branch] at the top of a rule says its Name; every Branch says its When and its Then. It
// fires at its moment and is done within the pass. A rule turning to places (cond.At, [At]: the
// places its moment's host tells) runs once a place: the smallest part of it holding every
// clause of a place.
//
// # For package plan
//
// A plan is made of what lasts over steps, in package plan, out of the same steps: every
// [Step] is a [Lasting], and [NewWait], [NewIdle], [NewLoop], [LastingSteps], [LastingOneOf]
// and [TopNode] are the parts package plan builds on.
package act
