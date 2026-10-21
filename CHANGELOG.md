# Changelog

## v0.3.1

- A step's `Do(cmd)`, in a rule or a plan, gives its command without allocating: before, two
  allocations a command (the copy told whom it aims at, and its box for the carrier). gram's
  navigation tick of 1,000 walkers in the open: 2,599 allocations → 90, its time unchanged.
- `command.Carrier.PutFromOf(entity, cmd)`: `PutFrom` of a command whose type is known, put into
  its `Queue` without being boxed; a `Routed` command, or one of an interface type, goes as
  through `PutFrom`.

## v0.3.0

Places are the host's library's, ceremony naming none of its own.

- `host.NewPlacing(name, takes)` is a kind of places a library's hosts tell for their moments
  (those round where an entity stands, those in front of it), `takes` saying which moments have
  them; `host.Places{Placing, N}` are some of a kind within `N`; `host.Tick.Places` tells them.
- `cond.At(places)` and `act.At(places)` turn a rule to each place, once a place, as
  `Around(n)` and `Ahead(n)` did; one kind and one reach of places a rule; a moment its kind
  does not take refuses the rule as it is built.
- Gone: `cond.Around`, `cond.Ahead`, `act.Around`, `act.Ahead`, `plan.Around`, `plan.Ahead`,
  `host.Placed`, `host.Headed`, `host.Tick.Around`, `host.Tick.Ahead`: a library having such
  places names its own kinds.

## v0.2.1

- `cond.Ahead(n)` and `act.Ahead(n)`: the places on the cells one to n steps ahead of the
  entity, the way it faces, beside `Around(n)`; a rule of a moment that is `host.Headed` (a
  `Placed` one facing somewhere), the host telling them through `Tick.Ahead`. One kind of
  places a rule: `Around` and `Ahead` in one are refused as it is built.
- `goke`: `Active.Left(e)` and `Wide.Left(e)` say what remains of an effect on the entity
  (`Forever` for one without an end, false while it is not under it), for whoever draws an
  effect's course.
- `goke`: `NewPlans(plans, opts...)` and `NewEffects(register, opts...)` take goke options for
  their queries from the caller (a filter, a pace: goke 3.5's `PacedBy`) and walk them with
  `Query.Paced`, an effect counting down by the step its block is given. Without options, as
  before.
- Fixed: what `Then` casts begins in the pass its forerunner runs out, not the next — there
  was a step with neither effect's marker, a false window for every rule asking "not under".
- `doc/ceremony.md`: "How an effect runs", the rules of a cast on a running effect, `Take`,
  running out, `Then`, `Hold`, `Forever`, `For`, `Toggle`, netting.

## v0.2.0

The language is free of goke: the one package importing it is the binding `ceremony/goke`, and
the engine reads and writes the store through the contract `store` at the cost of the typed
code it replaced (gram's benchmarks of plans, rules and effects, alternately against the tree
before).

- `store`, new: `Row` and `Col` (a block of entities, an array a column, the engine indexing
  it and nothing else), `Columns` and `Facts` (what a binding gives the engine to bind the
  columns it reads and to stage what it writes), `Mind`, `StepSet`, `MaxSteps` (from
  `internal/steps`; `plan/store.go` and its aliases are gone).
- `goke`, new: the binding — `Columns` (`NewColumns(qb)`, `Own`, `Row`), `Spawner`, `LoadComp`,
  `Effects` (`NewEffects(register)`: the register itself and the `effect.Runtime`, its
  `Module`, `Active` and `Wide`), `Plans` (`NewPlans(plans)`: the system and the `Facts`).
- `effect` is definitions alone: `Effects` is the register (`Define`, `Named`, `Lookup`,
  `Guard`, and for a binding `Definition`, `Count`, `Freeze`, `Lasts`), `Effect` has `ID`;
  `Runtime` is what casts and asks (`Has`, `Cast`, `CastFor`, `Dispel`, without a command
  buffer); `Granted` and `Alteration` are what a `Grant` and an `Alter` define, the latter a
  function of an address. `Active`, `Wide`, `Effects.Cast/CastFor/Dispel/Has`, `Module` and
  `Effect.On` are gone from `effect`.
- `host`: `Tick` carries no command buffer, its `Effects` is an `effect.Runtime`, `TickSource`
  takes the step alone; `Rules.Bind` and `PairRules.Bind` take `store.Columns`, `Run` and
  `RunWhere` a `store.Row` and a count, `PairRules.Read` replaces `InChunk` and `At`;
  `Columns`, `Of` and `Own` are gone (`goke.Columns.Own` shares a pass's own column).
- `plan`: `New` makes a `plan.Planned` (from `act`), `Plans` is a real type (`NewPlans`,
  `Register`, `Roles`, `Bind`, `Facts`, `Tick`); `act.Planned`, `act.NewPlan` are gone,
  `act.TopNode` is for package plan.
- `comp`: a `Comp` is `Type`, `Row` and `Fill(dst, row)`; `Spawner`, `LoadToken`, `TypeOf`
  and `RowOf` are gone.
- Requires goke v3.4.0 (`RegCompOf`, `LoadCompOf`, `Column`, `OptColumn`, `CmdBuf.Reserve`,
  `ECS.CmdBuf`, `Comp.Index`).

Saves name the component types by their paths: `effect.Active` and `Wide` are `goke.Active`
and `Wide`, `steps.Mind` is `store.Mind`, so saves written before do not load.

## v0.1.0

Carved out of [gram](https://github.com/kjkrol/gram) at commit `fb56d8d`: the language and the
engine of behaviour, with nothing of the game in it. The packages, where they came from and
what changed on the way:

- `tag` and `comp` out of gram's `entity/tag` and `entity/kind/comp`, as they were.
- `fact`, a new leaf: the mark `fact.Fact` (a type is a fact by its method `Fact()`, confirmed
  where it is defined), `From` (a message's sender; it was `rule.From`), `Subject` (a fact or a
  moment about another entity, one definition for `host` and the engine), `Message`, `Sent`,
  `MaxChain`, `AskLife`. A loop's type must be a fact or `any`.
- `effect` out of `rule/effect`, without drawing: `Look`, `Sprites` and `Looks` are gone — the
  game keeps the twins of sprites in its atlas. `effect.Knob` marks a component an effect may
  turn, and `Alter[T Knob]` takes knobs alone.
- `command` out of gram's `control/command.go`: `Command`, `Queue`, `CommandQueue`, `Carrier`,
  `Issued`, `Routed`, `Unwrap`; `command.Issuer` (`Nobody`) in place of `control.PlayerID`, the
  field `Issued.Issuer` in place of `Issued.Player`.
- `host` out of gram's `plugin` (`rules.go`, `pair_rules.go`, `tick.go`, `moments.go`):
  `Rules`, `PairRules`, `StepRules`, `Tick`, `TickSource`, `About`, `Met`, `Placed`, `Aimed`,
  `Hosted`, `Host`, `ErrUnhosted`, `ErrHostBuilt`.
- `cond`, `act` and `plan` out of `rule/cond`, `rule/act` and `rule/plan`; the root `ceremony`
  out of `rule` (`Part`, `NewPart`, `Plays`, `Played`, `Command`, `Put`, `Take`, `Toggle`,
  `Trigger`) and gram's `entity/label.go` (`Label`, `LabelOf`, `Whom`, `Named`, `Group`, `ID`,
  `World`); `ceremony.Deliver` out of gram's `internal/hosts`, taking any `act.Rule` and a
  role's rules one by one.
- `internal/steps` as it was, the engine of trees of steps, minds and messages.

Saves of a game record the component types by their paths (`ceremony/tag.Tags[…]`,
`internal/steps.Mind`, `effect.Active`), so saves written before this carving do not load.
- `plan/store.go` exposes the engine's `Mind`, `Plans`, `NewPlans`, `PlanID` and `MaxSteps` as
  aliases for a user over goke, until the store binding takes them.
