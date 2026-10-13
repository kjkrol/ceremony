// Package goke binds ceremony to the goke ECS, the one package of the library that imports it:
// the store that keeps the entities, reads them for the engine and stages what it writes,
// every read and write through goke's own queries and command buffers.
//
// # Columns and Spawner
//
// [Columns] binds what a host or the plans read to a query as optional columns and hands each
// block of it to the engine as a store.Row ([Columns.Row]), the arrays goke's cursor gives;
// a pass binds its hosts with it (host.Rules.Bind, host.PairRules.Bind). [NewSpawner] writes
// a kind's components onto the entities a factory makes; [LoadComp] names one to a save.
//
// # Effects
//
// [Effects] runs an effect.Effects register: it is the register, to define and name the
// effects, and their effect.Runtime, casting and asking; its [Effects.Module] is installed in
// the ECS and runs the effects' system once a step. [Active] is what an entity is under, saved
// with it; [Wide] the same with a slot for every effect, for an entity standing for a whole.
//
// # Plans
//
// [Plans] runs a plan.Plans: a system walking every entity with a store.Mind once a step and
// the store.Facts the plans read and write through.
package goke
