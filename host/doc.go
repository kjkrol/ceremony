// Package host is the pass's side of the rules: a pass that perceives a moment of its entities
// keeps a host of that moment, binds it to the columns of its query (store.Columns) and runs it
// over every block of entities (store.Row), and the host runs the rules given to it. A rule of a moment no host in use catches is an error, never
// a silent no-op.
//
// # Moments
//
// A moment is what a pass perceives of an entity, handed to rules and forgotten: a value that
// is [About] (Who, the entity it is of), [Met] for one entity with others. [Aimed] is a command
// told the subject of the moment it is given on.
//
// # Places
//
// A library names the kinds of places its hosts tell for their moments ([NewPlacing]: those
// round where an entity stands, those in front of it), each saying which moments have them;
// [Places] are some of a kind, whom a rule's cond.At and act.At turn to, told by [Tick.Places].
//
// # Hosts
//
// [Rules] runs the rules of a moment over the entities of a pass; [PairRules] matches pairs by
// the tag families the rules name ([Side], [Marks]); [StepRules] runs rules of a moment of the
// whole once a step. Each is a [Host]: Add takes a rule or refuses it with [ErrUnhosted], and
// after Bind with [ErrHostBuilt]. [Tick] is what a host tells its rules of the pass: the
// commands' carrier, the effects' runtime, the time, the seed, the places and the roles played.
package host
