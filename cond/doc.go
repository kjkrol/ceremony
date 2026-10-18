// Package cond is the When of a branch: who is in what state, every clause saying whom it is
// about.
//
//	When: cond.Other.If(traveller).
//		And(cond.Self.If(cond.Not(talking))).
//		And(cond.Self.If(cond.Not(talked))),
//
// # Clauses
//
// [Self], [Other], [Around], [Ahead] and [On] each take one state: an effect the entity is under, a
// role it plays, a [Word] of the moment, [Chance], [Always], or [Not] of any of them. Clauses
// join with And and Or; [Not] of a whole Cond turns it round. At the top of a rule a role of
// Self or of Other narrows whom the host walks; anywhere else it is asked of the entity as the
// branch runs.
//
// # Words
//
// A moment's or a fact's words are its owner's, made with [NewWord]: a predicate of the moment
// or the fact under a name. A word says whether it is of the entity or of the other
// ([Word.OfOther]), and a clause of the wrong side is refused as the branch is built.
package cond
