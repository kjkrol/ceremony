// Package command is how anything is asked for: a command is a value carried by its type to the
// one handler whose queue takes it, from a user and from an entity alike, and waits there for
// the handler's pass. Nothing is dropped and nothing is called back.
//
// # Command and Queue
//
// A type is a [Command] by its method Command, a mark confirmed where it is defined. A
// [Queue] holds the commands of one type as [Issued] — by an [Issuer] outside the entities, or
// by an entity for itself — until its handler drains them. [CommandQueue] is a Queue with its
// type erased.
//
// # Carrier
//
// A [Carrier] takes each command to the queue of its type, one carrier for every queue in use;
// a [Routed] command is written one way and carried as another, and [Unwrap] is what the
// handler takes.
package command
