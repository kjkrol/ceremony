// Package host is the pass's side of the rules: a pass that perceives a moment of its entities
// keeps a host of that moment, binds it into its query and runs it over every chunk, and the
// host runs the rules given to it. A rule of a moment no host in use catches is an error, never
// a silent no-op.
//
// # Moments
//
// A moment is what a pass perceives of an entity, handed to rules and forgotten: a value that
// is [About] (Who, the entity it is of), [Met] for one entity with others, [Placed] for one
// standing on places that are entities of their own. [Aimed] is a command told the subject of
// the moment it is given on.
//
// # Hosts
//
// [Rules] runs the rules of a moment over the entities of a pass; [PairRules] matches pairs by
// the tag families the rules name ([Side], [Marks]); [StepRules] runs rules of a moment of the
// whole once a step. Each is a [Host]: Add takes a rule or refuses it with [ErrUnhosted], and
// after Bind with [ErrHostBuilt]. [Tick] is what a host tells its rules of the pass: the
// commands' carrier, the effects, the time, the seed, the places round and the roles played.
package host
