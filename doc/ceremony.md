# ceremony: the language

One vocabulary for every behaviour of an entity: a **rule** is what is done at a moment a pass
catches, a **plan** is what an entity does over time, an **effect** is a state that holds for a
while, a **command** is what an entity asks for, and a **fact** is what it is told. A branch
(`act`) says when what is done, a loop (`plan`) when what is done over time and when it is over,
every clause of either (a condition, `cond`, or a step) saying whom it is about; the root package
`ceremony` ties them together: the roles whose players obey rules, the commands about effects,
who an entity is called.

## Two bricks: a Branch and a Loop

A rule is a sentence — **who, when, what** — and an `act.Branch` is all but the who: `When` the
condition, `Then` the steps, each clause of either said of a subject (`cond.Self.If(x)`,
`act.Other.Do(cmd)`). Who is the role the branch is given to, so the rules of a role read as a
paragraph under it. The type `P` a Branch is of is the moment it fires on; `any` for a branch
inside another, asking only of state. A plan is made of `plan.Loop`s, a Branch with an end: its
`P` is the fact it waits for, or `any`; `Until` says when it is done, `Timeout` when it has
failed, `Cooldown` how soon it may begin again.

```go
// a rule: a ticket's desk reviews every ticket once an hour (the moment Review, Age the time
// since it was filed); one overdue and not yet escalated is, once
overdue := cond.NewWord("overdue", func(r Review) bool { return r.Age > 2*time.Hour })
desk := ceremony.NewPart("desk", deskTag).Obeys(
	act.Branch[Review]{
		Name: "escalate",
		When: cond.Self.If(overdue).And(cond.Self.If(cond.Not(escalated))),
		Then: act.Steps(act.Self.Put(escalated), act.Self.Do(Notify{To: "lead"})),
	},
)

// a plan: an order placed is invoiced and waits to be paid; paid, it ships; unpaid in two
// weeks, it is cancelled
order := plan.New("order", plan.OneOf(
	plan.Loop[Received]{
		Name: "fulfil",
		When: cond.Self.If(cond.Always),
		Then: plan.OneOf(
			plan.Steps(
				plan.Loop[any]{
					When:    cond.Self.If(cond.Always),
					Then:    plan.Self.Do(Invoice{}),
					Until:   cond.Self.If(Paid{}),
					Timeout: 14 * 24 * time.Hour,
				},
				plan.Self.Do(Ship{}),
			),
			plan.Self.Do(Cancel{}),
		),
	},
	plan.Loop[any]{Name: "idle", When: cond.Self.If(cond.Always), Then: plan.Self.Idle()},
))
```

`Review` is a moment (it has `Who()`), `Received` and `Paid` are facts (they have `Fact()`),
`escalated` an effect, `Notify`, `Invoice`, `Ship` and `Cancel` commands (they have `Command()`)
carried out by whoever handles them.

The grammar:

```
rule    = act.Branch[Moment]                   // given to a role: Part.Obeys(rules...)
plan    = plan.New(name, root)                 // carried by the entity in its mind
root    = plan.OneOf(part...) | plan.Steps(part...) | loop
branch  = act.Branch[P]{Name, When, Then}      // Name at the top; When and Then always
loop    = plan.Loop[P]{Name, When, Then, Until, Timeout, Cooldown}
P       = Moment (a branch) | Fact (a loop) | any
When    = cond.Cond                            // a loop's Until the same: done once it holds
Cond    = clause | Cond.And(Cond) | Cond.Or(Cond) | cond.Not(Cond)
clause  = subject.If(what)
subject = cond.Self | cond.Other | cond.Around(n) | cond.Ahead(n) | cond.On(owner)
what    = effect | role | word | fact{} | cond.Chance(p) | cond.Always | cond.Not(what)
          // fact{}: in Until alone, of Self alone
word    = cond.NewWord(name, func(P) bool) | ... .OfOther()
Then    = step | act.Steps(Then...) | act.OneOf(Then...) | branch    // a branch's
part    = Then | plan.Self.Wait(d) | plan.Self.Idle()                // a loop's
        | plan.Steps(part...) | plan.OneOf(part...) | loop
step    = actor.Put(e) | actor.Take(e) | actor.Hold(e) | actor.Do(cmd)
actor   = act.Self | act.Other | act.Around(n) | act.Ahead(n)   // in a plan: plan.Self, plan.Other
```

- A Branch given to a role (`ceremony.Part.Obeys`) is a rule of `P` for the host that catches
  `P`, obeyed by the role (below, "Roles"); its `Name` is for the errors about it. A Branch
  inside another's `Then` asks more of the same `P`; one of `any` asks only of state. Every
  Branch says its `When` (`cond.Self.If(cond.Always)` for every time) and its `Then`; one missing
  panics by the branch's name. So does what the rule cannot run, as it is built: a `Not` of a
  state outside a clause, a word of the wrong side or of another moment, a fact in a `When`,
  `Around` where `P` is not `host.Placed`, `Ahead` where it is not `host.Headed`, `Around` and
  `Ahead` in one rule, a Branch of another moment inside, a message, a Branch
  at the top of a rule without a `Name`. A plan's parts — `Wait`, `Idle`, a Loop — do not
  compile in a rule's `Then`.
- `plan.New(name, root)` is a plan, a `plan.Planned`: the component its entities carry, their
  mind, and the plan defined under the name a save knows it by, run in every step. `OneOf` is
  run every step, from the top, the first loop whose When holds; an earlier one, once it can,
  takes over from a later one, which is stopped. `Steps` runs its parts one after another and
  remembers where it is. A plan done begins again the next step. Every loop at the top of a plan
  — a part of its root — says its `Name`; one inside a `Then` goes by the outer's. A rule's steps
  and groups stand in a plan too.
- **A fact is a thing that happened.** A Loop of a fact `P` — a component told to the entities
  with a plan: `Paid`, `Arrived`, a message — begins once as the fact comes to hold its When (it
  was absent, or its word did not hold; one carried already counts once), runs its Then to the
  end though the fact goes, and does not begin again until the fact comes afresh. A Loop of
  `any` is on a state: it runs while its When holds and is stopped when it stops.
- **The end is in the head.** `Then` is steps alone, run once from the start; a loop with an
  `Until` then waits for it. `Until` is checked every step from the loop's start, before the
  steps: once it holds the loop is done, its steps stopped — before any, when it holds at the
  start. A fact's value in it (`cond.Self.If(Paid{})`) is the fact come afresh since the loop
  began, one carried from before not counting; a word beside it holds while the fact is carried
  and the word holds. `Timeout` fails the loop once it has run that long — in a `OneOf` the next
  part takes over — and `Cooldown` keeps it from beginning again within that of its last run.
  `Wait(d)` and `Idle()` are steps of `plan.Self` alone; an `Until` of `Other` is refused. A rule
  runs in one pass and has none of them.
- A function writing part of a rule returns an `act.Step`, part of a plan a `plan.Part`, a whole
  rule an `act.Rule`, a whole plan a `plan.Planned`. The one who perceives ships no ready-made
  reactions: it gives moments and their words, the user writes the rule.

## Conditions

A `When` is a `cond.Cond`: clauses joined with `And` and `Or`, `cond.Not` of a whole condition
turning it round. A clause is a subject in a state, `subject.If(what)`:

- **whom**: `cond.Self`, the entity the branch is about; `cond.Other`, the other of the pair the
  moment met, or the subject of the moment or the fact — the nearest seen, the one touched, who sent
  a message — which does not hold while there is none; `cond.Around(n)`, each place within `n` rings
  of where the entity stands, in a rule of a `host.Placed` moment, and `cond.Ahead(n)`, each place
  on the cells one to `n` steps ahead of it, the way it faces, in a rule of a `host.Headed` moment
  (below, "Where one stands"); `cond.On(owner)`, an owner's own entity — anything with an
  `Entity()`: `cond.On(world).If(alarm)` a state of the whole;
- **what**: an effect, the one under it; a role, the one playing it; a word of the moment or
  the fact — `cond.NewWord(name, func(P) bool)`, `.OfOther()` for one of the other;
  `cond.Chance(p)`, drawn for whom it is asked, afresh every step from the seed, the time and
  that entity, so a load and a replay draw alike; `cond.Always`; `cond.Not` of any of them; in
  a loop's `Until` alone, a fact's value (`Paid{}`), the fact come afresh — in a `When` a fact
  is the loop's type, `plan.Loop[F]`, and is refused as a state.

A word is of the entity or of the other, and a clause of the wrong side is refused as the branch
is built. A role of `Self` or of `Other` at the top of a rule — in its `And` chain — the host
reads off its own query, on the tag bits, so a rule that does not concern an entity costs nothing
for it; anywhere else a role is asked of the entity as the branch runs. A role narrows a rule
further, on top of its When (below, "Roles").

## The five words

- **Rule** — a moment a pass catches: one of every entity, one of a pair (`host.Met`), one of
  the whole (a step's). A rule's steps are instant — `OneOf`, `Steps`, a Branch asking more,
  `Put`, `Take`, `Hold`, `Do` — so what lasts (`Wait`, `Idle`, a Loop) does not compile in one,
  and a message is refused as the rule is made. A rule remembers nothing of its own, and runs no
  Go code of its own: there is no step for it. What the steps cannot say is a moment, a word, a
  step or a knob the owner still lacks.
- **Plan** — what an entity does over time (`plan.New`). It remembers its place (its mind, the
  component the plan is: the loops running, each one's place and start on the clock), waits
  (`Until`, `Wait`), talks to other entities (messages), and is saved with the entity. `OneOf` is
  a reactive choice, `Steps` a sequence with memory; a Loop of a fact runs once as the fact
  comes, to the end. A plan holds at most 128 steps.
- **Effect** — `effect.Effect`, defined once from a `Spec` (`Lasts`, `Stacking`, `Then`,
  `Grant`, `Alter`; `Described` a sentence for a UI, `Description()`), with its own marker, on
  while it runs (`Mark()`). `Put` puts it on; `Hold` holds it as long as its branch runs, or as
  long as a rule keeps firing it; `Take` takes it off. **An effect's presence is state**: a
  branch `When: cond.Self.If(cond.Not(escalated))` that puts `escalated` on is "at most once a
  while", a memory for rules that keep none, and its marker is what rules of other passes ask
  for (`cond.Self.If(e)`, `cond.Other.If(e)`). At most 63 effects are defined, 8 on one entity
  at once.
- **Command** — `act.Self.Do(cmd)` gives the command of whom the step is said of, the same one a
  user gives, queued for the handler of its type, and goes on at once: **fire and forget**. The
  handler carries it out for the entity alone (`command.Issued.ByEntity`). One `command.Carrier`
  takes every command to the queue of its type, from users (`Issuer`) and entities alike, and the
  `host.Tick` hands it to the rules. Nothing is dropped: a command waits for its handler's pass.
  A command that is `host.Aimed` is told the subject of the fact it stands under, or of the
  rule's moment: whom the entity touched, who asked. A command embedding `fact.From`, given with
  `plan.Other.Do(…)`, is a message (below).
- **Fact** — a component told to the entities with a mind alone, so entities without a plan pay
  nothing. Facts are how a plan learns what came of its commands; a fact comes, and a loop of it
  runs once, or one waiting for it is done.

## How an effect runs

The effects' pass of a step settles every entity's effects at once; a rule or a command of the
step before lands there. The rules, for whoever writes a rule over a state:

| | |
|---|---|
| `Put` on an entity not under the effect | begins at the next effects pass: marker on, `Grant`s and `Alter`s applied, the time from then; `Has` says so from the cast |
| `Put` on an entity under it | refreshes its time to the Spec's (or the command's `For`), begins nothing again; with `Stacking`, a slot of its own instead, up to 8 on an entity (63 on a plugin's), none free nothing |
| `Take` | off at the next effects pass: marker off, the knobs put back; `Then` is not cast |
| the time runs out | off in that pass, and what `Then` names begins **in the same pass**, its time counted from the next step: no step without either |
| `Hold` in a rule | cast anew every step the branch fires, two steps long, so it lapses two steps after the branch stops; in a plan, held while the loop runs and taken off as it ends, failing the loop when someone else takes it |
| no `Lasts` | `Forever`: until `Take` |
| `.For(d)` on a command | d in place of the Spec's length, on `Put` and `Toggle` alike |
| `Toggle` | off them all where any target is under it, else on them all |
| a `Put` and a `Take` in one step | a binding may net a step's commands of one effect on one entity (gram does: the last given wins); a cast in the pass of a dispel revives |

## Idioms

- **Give, then wait for what comes of it.** `plan.Loop[any]{…, Then: plan.Self.Do(Invoice{}),
  Until: cond.Self.If(Paid{})}`: a fact in an `Until` is one come *afresh* — a fact left from
  before does not count — and a word beside it, `.And(cond.Self.If(word))`, waits for it to come
  to hold the word.
- **A round is two loops.** `plan.Steps` of a loop giving the next leg `Until` its fact of
  arriving comes, and one standing a while (`plan.Self.Wait(pause)`), goes round, the plan done
  beginning again (the patrol of the README's guard).
- **No answer is a timeout, not a fact.** `plan.Loop[any]{When: cond.Self.If(cond.Always),
  Then: plan.Self.Idle(), Until: cond.Self.If(reply{}), Timeout: d}` as the first part of a
  `plan.OneOf`, the part after it what to do then.
- **Not too often.** `Cooldown: time.Second` on a loop has it begin no sooner than a second
  after the last.
- **A rule's memory is an effect.** `When: cond.Self.If(cond.Not(e))` with `act.Self.Put(e)`
  among the steps runs a thing and not again for as long as `e` lasts.
- **Where one stands.** On a moment that is `host.Placed` — the host tells the places round
  through `Tick.Around` — `cond.Around(n).If(x)` and `act.Around(n).Put(e)` turn to each place
  within `n` rings of where the entity stands, each once; `0` the places it stands on alone. On
  one that is `host.Headed` — the host tells the places in front through `Tick.Ahead` —
  `cond.Ahead(n).If(x)` and `act.Ahead(n).Do(cmd)` turn to each place on the cells one to `n`
  steps ahead, the way the entity faces, the nearest first: a lever in front of a unit, a wall
  three cells on. The smallest part of the rule holding every clause of a place runs once a
  place; one kind of places and one ring a rule.
- **Now and then.** `cond.Self.If(cond.Chance(p))` holds with likelihood `p`; a coin is
  `act.OneOf(act.Branch[any]{When: cond.Self.If(cond.Chance(0.5)), Then: a}, b)`.
- **One after another.** `effect.Then(next)` casts `next` when an effect's time is up — burning
  leaves smouldering — but not when it is taken off: put out, nothing smoulders.

## Commands

Whatever somebody asks for — a user's key, a script, a branch's `Do` — is a command, and a
command about an effect is one sentence (`ceremony.Command`):

```go
holdLate := ceremony.Put(onHold).On(ceremony.Group("late orders")).By(ceremony.Named("audit"))
closeDesk := ceremony.Toggle(closed).On(ceremony.Named("desk 3"))
alarm := ceremony.Put(alarmed).On(ceremony.World).For(30 * time.Second)
hasten := ceremony.Put(haste).On(chosen.Playing(hasty)) // an owner's Router
```

- **The verb**: `ceremony.Put` puts the effect on, `ceremony.Take` takes it off,
  `ceremony.Toggle` takes it off them all where any is under it and puts it on them all
  otherwise — a switch. In a branch `act.Self.Put(e)` and `act.Other.Take(e)` are steps, acting
  on whom they are said of; a command with an `On` is given as any command,
  `act.Self.Do(ceremony.Put(alarmed).On(ceremony.World))`.
- **Whom** (`On`): `ceremony.Named(names...)`, the entities bearing those names, one each;
  `ceremony.Group(names...)`, all those in the groups; `ceremony.ID(ids...)` those of the
  identifiers; `ceremony.World`, the one entity standing for the whole; or an owner's own, a
  `ceremony.Router` — its `Route` is the command the owner carries out (a selection's
  `Selected(roles...)`, say). An entity is called by its `ceremony.Label` (`LabelOf(name,
  group)`): a name is one entity's, a group many's.
- **How long** (`For`): in place of what the effect's `Spec` says.
- **Who sets it off** (`By`): the entities that give themselves `ceremony.Trigger{}` — a lever,
  a plate, an audit. A role then says only *that* a plate stood on triggers; the command says
  what it opens.

`Command.Routed()` is the command as its target's owner takes it, and `command.Carrier.Put`
unwraps a `command.Routed`, so a key, a script and a rule give it one way; one naming nobody
panics there. Whoever keeps a register of commands by name sets `Command.Name` (`Defined()`),
keeps those with a `By` for `Trigger`, and checks the names they say as it starts. A command is
given the same way whoever gives it:

```go
carrier.Put(user, holdLate)       // a user, a script
act.Self.Do(holdLate)             // a rule; plan.Self.Do(holdLate) a plan
act.Self.Do(ceremony.Trigger{})   // its source, set off
```

The owner of the target carries it out a step later, with the step's effects; commands of one
effect for one entity in one step are netted, so a switch flipped twice stays as it was.

Rules and plans read the whole's states with `cond.On(world).If(e)`, as they read an entity's
with `cond.Self.If(e)`:

```go
fx.Define("alarm", effect.Spec{effect.Lasts(30 * time.Second)})
fx.Define("alert", effect.Spec{effect.Alter(func(p *Pace) { p.MaxSpeed *= 1.5 })})
guard := ceremony.NewPart("guard", guardTag).Obeys(act.Branch[Standing]{
	Name: "hurry",
	When: cond.On(world).If(fx.Named("alarm")),
	Then: act.Self.Hold(fx.Named("alert")),
})
soundAlarm := ceremony.Put(fx.Named("alarm")).On(ceremony.World)
```

The guards hurry while the alarm sounds and slow down once it is over: `Hold` holds `alert` only
while the rule fires it.

## Messages

A plan talks to another entity with the bricks it has. A command embedding `fact.From`, given
with `plan.Other.Do(…)`, lands on that entity as a fact in place of going to a queue, the engine
filling in who sent it; one without a plan takes none, and the step fails. A Loop on the
command's type is how the other answers — its `Other` is the sender, and `plan.Other.Do(…)`
again the reply, a command of its own — or passes it on, `Other` under a loop of another fact
being that fact's subject. A message nobody takes up is dropped after `fact.AskLife`; one passed
on `fact.MaxChain` times is not sent, so a chain ends. Delivery a step later means no recursion
within a step.

```go
type pass struct{ fact.From }
type reply struct {
	fact.From
	Yes bool
}

func (pass) Command()  {}
func (reply) Command() {}

func (r reply) yes() bool { return r.Yes }

// the asker: let me pass (need is a fact whose Subject is whom it needs past); walks on let
// through, flees turned away or unanswered
plan.Loop[need]{
	Name: "ask to pass",
	When: cond.Self.If(cond.Always),
	Then: plan.Steps(
		plan.Other.Do(pass{}),
		plan.OneOf(
			plan.Loop[any]{
				When:    cond.Self.If(cond.Always),
				Then:    plan.Self.Idle(),
				Until:   cond.Self.If(reply{}),
				Timeout: 500 * time.Millisecond,
			},
			plan.Self.Do(flee{}),
		),
	),
},
plan.Loop[reply]{
	Name: "let through",
	When: cond.Self.If(cond.NewWord("yes", reply.yes)),
	Then: plan.Self.Do(walk{}),
},
plan.Loop[reply]{
	Name: "turned away",
	When: cond.Self.If(cond.Not(cond.NewWord("yes", reply.yes))),
	Then: plan.Self.Do(flee{}),
},

// the one asked: lets pass where it has room, jumping aside; asks its ally on where it has one
// (room is a fact whose Subject is the ally beside) and passes the answer back
plan.Loop[pass]{
	Name: "asked to let pass",
	When: cond.Self.If(cond.Always),
	Then: plan.OneOf(
		plan.Loop[room]{
			When: cond.Self.If(cond.NewWord("free", room.free)),
			Then: plan.Steps(
				plan.Other.Do(reply{Yes: true}),
				plan.Self.Do(jump{}),
			),
		},
		plan.Loop[room]{
			When: cond.Self.If(cond.NewWord("by an ally", room.byAlly)),
			Then: plan.Other.Do(pass{}),
		},
		plan.Other.Do(reply{}),
	),
},
plan.Loop[reply]{
	Name: "pass the answer on",
	When: cond.Self.If(cond.Always),
	Then: plan.Loop[pass]{
		When: cond.Self.If(cond.Always),
		Then: plan.Other.Do(reply{Yes: true}),
	},
},
```

The user's types are the words: `pass`, `reply`, `need`, `room`. Nothing in `ceremony` names a
conversation. The whole exchange is `plan/message_test.go`.

## Roles

A **role** is a behaviour several kinds of entity share, said once: the rules those playing it
obey. `ceremony.NewPart(name, tag).Obeys(rules...)` says it, each rule narrowed to the role's
players; the tag is a bit of `cond.Roles`, one name one bit, 64 roles at most. An entity plays
its roles through the component `ceremony.Plays(roles...)` (a `tag.Tags[cond.Roles]`), and a
command may be for those playing one alone.

```go
fx.Define("haste", effect.Spec{effect.Lasts(3 * time.Second),
	effect.Alter(func(p *Pace) { p.MaxSpeed *= 2 })})

mortal := ceremony.NewPart("mortal", mortalTag).Obeys(act.Branch[Standing]{
	Name: "fall in",
	When: cond.Self.If(fallen),
	Then: act.Self.Do(Despawn{}),
})
hasty := ceremony.NewPart("hasty", hastyTag)

scout := kind.Spec{..., ceremony.Plays(mortal, hasty)}
porter := kind.Spec{..., ceremony.Plays(mortal)}
```

Nobody hands a rule to a pass: there is no hooking. Once every role is defined,
`ceremony.Deliver(hosts, roles...)` gives every rule of every role to the first host that takes
it, trying them in order; a rule none takes is an error wrapping `host.ErrUnhosted`, naming the
rule by its `String` — its name, its moment and its role: `"fall in" of Standing, for the role
mortal`. A rule of a moment of the whole (`host.StepRules`, run once a step walking no entities)
asks nothing of the entity and fires while the entity the moment is about plays the rule's role
(`Tick.Roles`).

`Obeys` narrows a rule on top of its own When: a rule of `cond.Self.If(hungry)` obeyed by mortal
fires for a mortal entity while it is hungry, one of `cond.Self.If(hasty)` for a mortal that is
hasty. On a moment that is `host.Met` it fires for the pairs whose own entity plays the role; a
pair rule with no `cond.Other.If(role)`, or whose other's role is its own, still runs at most
once a pair (`PairRules.DispatchEitherWay`), and in a moment grouped by observer
(`DispatchGrouped`) its `cond.Other.If(role)` narrows what the observer sees, not whether the
rule fires. An entity playing two roles that obey the same rule obeys it once for each.

A role in a clause holds while the one the clause is about plays it; under
`cond.Around(n).If(role)` that is the place turned to: `When: cond.Self.If(pull).And(
cond.Around(1).If(lever)), Then: act.Around(1).Do(ceremony.Trigger{})` pulls the lever beside
and nothing else there.

**Roles mean behaviour, names and groups mean which ones.** One rule of the role plate serves
every plate: which doors a plate opens is the command naming it. A hundred levers are a hundred
commands, a hundred names and the same rules — never a role or an effect each. A user defines no
tags: tags are how the owners keep their own states and groups, changed by their commands;
`cond.Self.If(e)` asks for an effect, and `cond.Other.If(role)` lets through the pairs whose
other plays a role.

## Knobs

An owner gives **knobs**: components it only reads, marked `effect.Knob`, which an effect's
`Alter[T]` turns and puts back when it ends. What the owner's own pass writes lives beside the
knob in a component of its own, given where it is missing: an `Alter` ending puts back the
original, and would put back stale state with it. A new knob goes into the component the owner
reads, never into its state.

```go
type Pace struct{ MaxSpeed float64 }

func (Pace) Knob() {}

var _ effect.Knob = Pace{}
```

## No Go code in a rule

A rule holds no Go code of its own but its words. What an owner does of its own — a price, a
log, a count — is its own work in a pass it makes anyway, never a rule of another's moment; a
rule says what it can in steps, and a missing moment, word, step or knob is added to the owner,
not written as code in a rule. There is no step running a function: it would hide what was
missing.

A switch of a behaviour for the whole is an effect on the one entity standing for it
(`ceremony.Put`, `Take` or `Toggle` on `ceremony.World`), the rule reading it with
`cond.On(world).If(e)`.

## Take and order

Many rules may touch one effect, from several passes. What happens is fixed:

- A rule that `Hold`s an effect someone took off has it back the step after: its cause goes on,
  so the effect does. A taker that is to win puts on a shield the holder keeps clear of —
  `act.Steps(act.Self.Put(doused), act.Self.Take(burning))` against
  `act.Branch[any]{When: cond.Self.If(cond.Not(doused)), Then: act.Self.Hold(burning)}`.
- A plan's `Hold` gives way when someone else takes its effect off: the loop fails and the plan
  goes on to what it does next.
- A `Put` after a `Take` in the same step takes the slot back.
- In a step, rules of one moment run in the order their roles came to be played; moments, in the
  order of their passes — that is the passes' order, not something a rule picks. Every `Put` and
  `Take` of a step lands together in the effects' pass at its end.
