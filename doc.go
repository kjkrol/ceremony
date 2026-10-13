// Package ceremony is a language and an engine for the behaviour of things that have state and
// are told what happens to them. An entity carries state; a fact is told to it when something
// happens (package fact); a state, an effect, is put on it for a while and turns knobs (package
// effect); a role says what it does about either — rules that fire within a pass at the moments
// it catches (package act, their conditions package cond) and plans that run over time on its
// mind (package plan) — and what a rule or a plan asks for is a command, carried by its type to
// whoever carries it out (package command). The hosts of rules, the pass's side, are package
// host; the components an entity is made of, package comp and package tag. This package ties
// them together: the roles whose players obey rules, the commands about effects, who an entity
// is called.
//
//	desk := ceremony.NewPart("desk", deskTag).Obeys(act.Branch[Review]{
//		Name: "escalate",
//		When: cond.Self.If(overdue).And(cond.Self.If(cond.Not(escalated))),
//		Then: act.Steps(act.Self.Put(escalated), act.Self.Do(Notify{To: "lead"})),
//	})
//
// # Roles
//
// A [Part] is a role: the rules those playing it obey ([Part.Obeys]), narrowed to its players.
// An entity plays roles through one component, [Plays]; [NewPart] makes a role out of a tag of
// cond.Roles, and [Deliver] hands every rule of every role to the host catching its moment.
//
// # Commands
//
// A [Command] is what somebody wants done about an effect, written as a sentence: [Put],
// [Take] or [Toggle] an effect, On whom, By whom set off, For how long. A user gives one, a
// rule gives one with act.Self.Do, and an entity gives every Command whose By names it with
// [Trigger]. Whom a command is for is a [Whom]: [Named], [Group], [ID] or [World]; an entity is
// called by its [Label].
package ceremony
