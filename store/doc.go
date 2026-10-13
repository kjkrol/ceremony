// Package store is the contract between ceremony's engine and the store that keeps the
// entities, what a binding such as ceremony/goke implements: the engine reads an entity's
// facts and tags off a [Row] the store hands it a block at a time, writes a fact into the
// place the store's [Facts] stage for it, and keeps a plan's state in the [Mind] the store
// carries on the entity.
//
// # Row and Col
//
// A [Row] is a block of entities: for every column bound through [Columns], the array of its
// values over the block, as the store lays them out, nil where the block carries none. The
// engine indexes it with [Row.At] and nothing else; the store decides what a block is and hands
// the arrays out of its own accessors, so the engine knows nothing of its layout.
//
// # Facts
//
// [Facts] is the writing side: [Facts.Stage] hands the engine the place to write a fact into,
// in the store's own staging, so a value is written once, where the store wants it, and lands
// as the store syncs; [Facts.Drop] takes a fact off the same way.
//
// # Mind
//
// [Mind] is the component every entity on a plan carries, saved with it; [MaxSteps] bounds a
// plan. A plan's shape changed, its saved minds do not load.
package store
