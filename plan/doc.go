// Package plan is what an entity does over time: loops run on its mind every step, made of
// the same steps a rule has and of what lasts.
//
//	plan.New("order", plan.OneOf(
//		plan.Loop[Received]{
//			Name:    "fulfil",
//			When:    cond.Self.If(cond.Always),
//			Then:    plan.Self.Do(Invoice{}),
//			Until:   cond.Self.If(Paid{}),
//			Timeout: 14 * 24 * time.Hour,
//		},
//		plan.Loop[any]{Name: "idle", When: cond.Self.If(cond.Always), Then: plan.Self.Idle()},
//	))
//
// # Loops
//
// A [Loop] says when it begins (When), what it does (Then) and when it ends: Until, a
// condition that comes to hold from its start — a fact's value in it, cond.Self.If(reply{}),
// is the fact come afresh — Timeout, after which it fails, and Cooldown, within which it does
// not begin again. A plan done begins again the next step. [Steps] runs loops in turn, [OneOf]
// the first whose When holds, an earlier one taking over once it can.
//
// # Steps
//
// [Self] has every step a rule's act.Self has, and [Entity.Wait] and [Entity.Idle], which
// last; [Other] is act.Other. A rule's steps and groups (act.Self.Do, act.Steps) are parts of
// a plan too; a plan's are not a rule's.
package plan
