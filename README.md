# ceremony

A language and an engine for the behaviour of things that have state and are told what happens
to them. An **entity** carries state; a **fact** is told to it when something happens; a
**state** is put on it for a while; a **role** says what it does about either — **rules** that
fire within a pass and **plans** that run over time — and what a rule or a plan asks for is a
**command**, carried by its type to whoever carries it out. Nothing is a callback: a rule is a
value saying who, when, on whom and what, checked as it is built; the engine runs it.

It was carved out of a game engine, where the entities are units, cells and the sky, and reads as
well over orders, tickets and cases: an entity is a case, a fact an event, a state a status with a
duration, a role a policy, a rule a reaction, a plan a workflow with its waits, a command a
service call. The store of state is behind a contract (`store`): the engine reads an entity's
facts and tags off the blocks a store hands it and writes into the places it stages; the one
binding is `ceremony/goke`, over the [goke](https://github.com/kjkrol/goke) ECS, the one package
that imports it.

## The dictionary

| word | what it is | in code |
|---|---|---|
| entity | a thing with identity carrying state | a `uid.UID64` and its components |
| fact | what happened to an entity, carried until it goes | `fact.Fact` marks the type |
| message | a fact with a sender | `fact.From` embedded in a command |
| moment | what a pass perceives of an entity, handed to rules | `host.About`, `Met` |
| places | entities a host tells of a moment, of a kind its library names | `host.NewPlacing`, `Places` |
| state | an effect: on for a while, with a marker, turning knobs | `effect.Spec`, `Effect.Mark()` |
| knob | a component an effect turns and its owner only reads | `effect.Knob`, `effect.Alter[T]` |
| role | a behaviour an entity plays: the rules it obeys | `ceremony.Part`, a bit of `cond.Roles` |
| word | a predicate of a moment or a fact, with its side | `cond.NewWord`, `.OfOther()` |
| condition | who is in what state | `cond.Self.If(x)`, `Other`, `At(places)`, `On(o)` |
| step | what is done, said of whom | `act.Self.Do(cmd)`, `Put(e)`, `Take(e)`, `Hold(e)` |
| command | a request carried by type to its handler | `command.Command`, `Queue`, `Carrier` |
| rule | a reaction within a pass | `act.Branch[Moment]{Name, When, Then}` |
| plan | a process over time on the entity's mind | `plan.New`, `plan.Loop[F]`, `plan.Self.Wait` |
| host | the pass catching a moment and running its rules | `host.Rules`, `PairRules`, `StepRules` |

## A rule

A rule is a sentence: who (the role), when (the condition), on whom and what (the steps). This
one is from the game it comes from: `Sighting` is a moment its vision pass catches (who sees
whom), `traveller` a role, `talking` and `talked` effects, `Begin` a command its dialogue pass
carries out.

```go
host := ceremony.NewPart("host", hostTag).Obeys(
	act.Branch[vision.Sighting]{
		Name: "talk",
		When: cond.Other.If(traveller).
			And(cond.Self.If(cond.Not(talking))).
			And(cond.Self.If(cond.Not(talked))),
		Then: act.Self.Do(dialog.Begin{}),
	},
	act.Branch[vision.Sighting]{
		Name: "part",
		When: cond.Self.If(vision.SeesNobody()),
		Then: act.Self.Take(talking),
	},
)
```

A rule holds no Go code but its words: `vision.SeesNobody()` is a `cond.Word` of the moment, made
with `cond.NewWord`. A rule fires at its moment, does its steps within the pass and remembers
nothing; its memory is an effect.

## A plan

A plan runs over time on the entity's mind, in loops: `Then` is run from the start as `When`
comes to hold, `Until` ends it, `Timeout` fails it. A loop of a fact (`Loop[Arrived]`) begins once
as the fact comes and runs to the end; one of `any` runs while its `When` holds. From the same
game: `MoveTo`, `StepAside` and `Onward` are commands its navigation carries out, `Arrived` and
`Blocked` the facts it tells, `Stuck` a word of `Blocked`, `alarm` an effect on the world's own
entity.

```go
guard := plan.New("guard", plan.OneOf(
	plan.Loop[any]{
		Name:    "to the gate",
		When:    cond.On(world).If(alarm),
		Then:    plan.Self.Do(navigation.MoveTo{To: ceremony.Named("gate")}),
		Until:   cond.Self.If(navigation.Arrived{}),
		Timeout: 20 * time.Second,
	},
	plan.Loop[navigation.Blocked]{
		Name: "blocked",
		When: cond.Self.If(navigation.Stuck()),
		Then: plan.Self.Do(navigation.StepAside{}),
	},
	plan.Loop[navigation.Arrived]{
		Name: "rest",
		When: cond.Self.If(cond.Always),
		Then: plan.Self.Wait(2 * time.Second),
	},
	plan.Loop[any]{
		Name:  "patrol",
		When:  cond.Self.If(cond.Always),
		Then:  plan.Self.Do(navigation.Onward{}),
		Until: cond.Self.If(navigation.Arrived{}),
	},
	plan.Loop[any]{Name: "idle", When: cond.Self.If(cond.Always), Then: plan.Self.Idle()},
))
```

`OneOf` is checked every step from the top, an earlier loop taking over from a later one once it
can; a plan done begins again. A plan talks to another entity with a **message**, a command
embedding `fact.From` given with `plan.Other.Do(msg)`: it lands on the other as a fact, and a loop
of its type is the answer. The language is in [doc/ceremony.md](doc/ceremony.md).

## The layers

- **The language** — `cond` (conditions), `act` (steps and the branch), `plan` (loops and what
  lasts), `effect` (states and knobs), and the root `ceremony` (roles, the commands about states,
  who an entity is called).
- **The engine** — `internal/steps` runs the trees of steps on minds and at moments; `host` is
  the pass's side: the hosts of rules and the `Tick` they hand the rules.
- **The store** — `store` is the contract: a `Row` is a block of entities, an array a column,
  which the engine indexes and nothing else; `Facts` stages what it writes; `Mind` is a plan's
  state on the entity. `comp` names a component for a kind's template, `tag` the families of
  bits. `goke` is the binding over the goke ECS: its `Columns` hand the engine each chunk's
  arrays out of goke's own accessors, its `Effects` run the effects, its `Plans` the plans, its
  `Spawner` writes a kind's components, all through goke's queries and command buffers.
- **Commands** — `command`: a queue a type, a carrier that sorts them, an issuer.

## Plugging in

- **Perceive**: write a moment type (`Who()` the entity; `Whom` for a pair; the places it has,
  told by `Tick.Places`, of a kind `host.NewPlacing` names), keep a `host.Rules[M]` (`PairRules`
  for pairs, `StepRules` for a moment of the whole), `Bind` it to the columns of the pass's query
  (`goke.NewColumns(qb)`) and `Run` it over every chunk (`cols.Row(cursor)`) with a `host.Tick`.
- **Tell facts**: put a component marked `fact.Fact` on the entity when the thing happens and
  take it off when it is over; plans begin on it and wait for it.
- **Define states and roles**: `effect.New(name)` then `Define(name, effect.Spec{...})`, run
  by the binding (`goke.NewEffects(register)`, its `Module()` installed); `ceremony.NewPart(name, tag).Obeys(rules...)`; a kind carries
  `ceremony.Plays(roles...)`; `ceremony.Deliver(hosts, roles...)` hands every rule to the host
  that catches its moment, failing with `host.ErrUnhosted` for one none does.
- **Carry commands**: a `command.Queue[C]` a type, given to a `command.Carrier`, which the
  `host.Tick` hands the rules; drain it in the handler's pass. A command about a state
  (`ceremony.Put(e).On(whom)`) is carried out by whoever owns the target.

## What is marked and asserted

A type is a fact by its method `Fact()`, a knob by `Knob()`, a command by `Command()` — marks
never called, confirmed beside the type (`var _ fact.Fact = Arrived{}`). A moment is one by its
`Who()`, a pair by `Whom`; one has places of a kind its `host.Placing` takes. What the grammar
cannot run is refused as the branch is built, by the branch's name: a word of the wrong side, a
fact in a `When`, `At` in a plan or on a moment without those places, a loop on a type that is
no fact.

## Install

```
go get github.com/kjkrol/ceremony
```

Requires Go 1.27.

## Licence

MIT.
