# CLAUDE.md

Guidance for Claude Code working in this repository.

## What this is

**ceremony** (module `github.com/kjkrol/ceremony`) is a language and an engine for the behaviour
of things that have state and are told what happens to them: entities carry facts and states,
play roles, obey rules at the moments a pass catches, run plans over time on their minds, and ask
for things with commands. README.md has the dictionary; `doc/ceremony.md` the language. The
store of state is behind the contract `store`; the one binding, `ceremony/goke`, is the only
package that imports goke (the user's word, 2026-10-13: a language, and a bridge that reads and
writes the data).

## Commands

```bash
go build ./... && go vet ./... && gofmt -l . && go test ./...   # the standard verification
go test ./plan/... -run TestName -v                               # a single test
```

## House rules

- **A comment says only what the code cannot.** An exported type, function, method or field
  gets at most two sentences: what it is and what it is for, never how it does it, never its
  history. Unexported code, the bodies of functions and the tests get no comment unless it
  carries a non-obvious why: a unit, a direction, a reason, a trap, a source. The sentence above
  a test stays only where the test's name does not say what it checks. No banners
  (`// === … ===`), no comment repeating the name or the next line, no constant made to carry a
  comment for a value used once.
- A package's `doc.go` keeps its `# Type` sections, each short: what the package brings and what
  a user does with it, never how it does it inside, no history.
- **Lines stay within 100 columns**, a tab counted as 4. Break a call between its arguments or a
  chain between its links, never right after an accessor, never by pushing a trailing comment
  past a closing parenthesis (it goes above the line). A struct literal of many fields is written
  a field a line. **One value a line**: never two assigned at once (`a, b = x, y`) to save a
  line. `golines -m 100 -t 4 --no-reformat-tags` does most of it; check what it made.
- **Examples are not tested**: what an example shows is tested in the package that does it.
- Commit messages use the conventional prefixes (`feat`, `fix`, `docs`, `refactor`, `perf`,
  `test`, `chore`) and carry no `Co-Authored-By` trailer. **Commit only when asked.**
- English, concise, no history in docs ("was", "used to be") outside CHANGELOG.md.

## Package layout

```
tag              Tags[F], Tag[F], Any, Anything: families of bits
comp             Comp, Template, Const, Load, Tagged, Marks, Without: a kind's components
fact             Fact (the mark), From, Subject, Message, Sent, MaxChain, AskLife
effect           Effects (the register), Effect, Spec (Lasts, Stacking, Then, Grant, Alter,
                 Described), Knob, States, Changed, Runtime (what a binding runs), Definition
command          Command, Issuer, Nobody, Issued, Queue, CommandQueue, Carrier, Routed, Unwrap
host             Rules, PairRules, StepRules, Tick, TickSource, Side, SideOf, Marks, About,
                 Met, Placed, Aimed, Hosted, Host, ErrUnhosted, ErrHostBuilt
cond             Cond, Subject (Self, Other, Around, On), If, Not, Chance, Always, Word,
                 NewWord, Roles, Role
act              Branch, Subject (Self, Other, Around), Step, Steps, OneOf, Rule, Build,
                 Lasting, TopNode and the bridges for plan
plan             Loop, Part, Steps, OneOf, Entity (Self with Wait and Idle), Other, New,
                 Planned, Plans, NewPlans (the runner a binding ticks)
store            Row, Col, Columns, Facts, Mind, StepSet, MaxSteps: the contract of a store
goke             the binding over goke: Columns (Column, Own, Row), Spawner, LoadComp, Effects, Plans,
                 Active, Wide
ceremony (root)  Part, NewPart, Plays, Played, Deliver, Command, Put, Take, Toggle, Trigger,
                 Target, Router, Label, LabelOf, Whom, Named, Group, ID, World
internal/steps   the engine: trees of steps, Instant (a rule's run), Plans (a plan's)
```

Imports go one way:

- `plan → act → cond → effect → tag`; `cond → fact`;
- `host → store, fact, command, effect, tag`;
- `act → host, store, command, internal/steps`; `plan → act, comp, store, internal/steps`;
- the root `ceremony → act, plan, cond, effect, command, comp, host`;
- `internal/steps → store, command, effect, fact`; `comp → tag`; `effect → tag`;
- `goke → store, plan, host, effect, comp, tag, goke/v3`.

External: `uid` everywhere, `goke/v3` in `ceremony/goke` alone: `go list -deps` of every other
package has no `goke/v3`. **Nothing imports gram or anything of a game**: no `render`, no
`world`, no plugin. `go list -deps ./... | grep gram` stays empty.

Where a file imports a public package and an internal namesake, the internal one gets an `i`
prefix. A public name only tests read is no API: it goes, or moves to `export_test.go`
(`act/export_test.go`: `Within`, `WithOther`).

## The seam between act and plan

`act` is a rule's: instant steps and the Branch. `plan` is what lasts, built out of `act`'s
bricks through one seam:

- `act.Step` is a part of a rule's `Then`; every `Step` has `Lasting()`, so it is a `plan.Part`
  too (a rule's steps and groups stand in a plan).
- `act.Lasting` is a step as a plan takes it. `plan` makes its own with the bridges written for
  it: `act.NewWait`, `act.NewIdle`, `act.NewLoop[P]`, `act.LastingSteps`, `act.LastingOneOf`
  and `act.NewPlan`; `plan.Loop.Lasting()` calls `act.NewLoop`, `plan.New` calls `act.NewPlan`.
- A plan's parts (`Wait`, `Idle`, a `Loop`) are `plan.Part`s and not `act.Step`s, so they do not
  compile in a rule's `Then`; what a rule cannot run anyway is refused as it is built.

`cond` lowers a condition into the engine's guards (`steps.Lowered`, `Lowering`) with a
`LowerAs` saying where it stands: `AtTop` (the roles of Self and Other go to the host's
filter), `Inside`, `AsUntil` (a fact's value allowed).

## What is checked at build, and why

A branch is a value; what the engine cannot run is refused as the branch is built, by its
name, so a mistake is found as the rules are defined and never as a silent no-op in a pass:

- a fact is marked (`fact.Fact`): a loop's type `P` must be a fact or `any`; a struct without
  `Fact()` in `If` is "no such state"; a fact's value is allowed in an `Until` alone, of `Self`
  alone, never under `Not`;
- a word says its side (`cond.Word.OfOther`): a word of the other in `Self`, of the entity in
  `Other`, or in `Around`, is refused; in a rule a word must be of the rule's moment;
- a loop's `Until` is lowered `AsUntil`; a loop and a branch need a `When` and a `Then`, one at
  the top of a rule or a plan a `Name`;
- places are a rule's: `Around` in a plan is refused, so is more than one ring in a rule, and a
  moment that is not `host.Placed` refuses `Around` as the rule is run;
- the other's role needs a moment with others (`host.Met`), and is said once; `cond.On` takes
  something with an `Entity()`; `cond.Not` of a state outside a clause says of nobody;
- a command naming nobody panics in `Command.Routed`; a plan needs a name and is defined once.

A rule no host takes is `host.ErrUnhosted` at `Deliver`, one handed over after its system was
built `host.ErrHostBuilt`.

## Testing conventions

Tests sit in the package whose code they test, external (`_test`) unless they need unexported
state; `act/helpers_test.go` and `plan/helpers_test.go` make roles out of tags without a store,
`plan/toys_test.go` a rig of commands and facts on a bare goke ECS through the binding
(`ceremony/goke`, imported by external tests of `act`, `plan` and `host` as `cgoke`). **A new semantics test is
checked with its fix disabled**: before a test of a behaviour is kept, the code that gives the
behaviour is taken out (a guard — `GuardCame`, a `Timeout` — made to pass) and the test must
fail; a test that passes either way tests nothing.

## The store and its binding: zero cost

The engine never knows the store's layout and never pays for not knowing it (the user's word,
2026-10-13: everything through the store's own systems, queries and command buffers, no
intermediate layer, no extra pointer to chase on the hot path, no lockpick into a chunk):

- **Reads.** A binding hands the engine a `store.Row` a block at a time: for every column
  bound, the array of its values as the store's own accessor gives it (`goke.OptColumn.Base`,
  `Comp.Slice` without the type parameter), nil where the block carries none. The engine
  indexes it (`Row.At`: nil check, `unsafe.Add(base, i*size)`) as `Slice(cur)[i]` does, and
  that is all it knows. The cost is a pointer a column a chunk, nothing an entity.
- **Writes in place** (a tag's bit, a knob under `Alter`, a slot of `Active`) are writes in
  place, through the same pointer.
- **Structural writes** (a message, a fact dropped, a family attached) go through the store's
  staging: `store.Facts.Stage` hands the engine the place reserved in goke's command buffer
  (`CmdBuf.Reserve`, the first half of `AddOne`) and the engine writes `*(*F)(p) = v` there, so
  a value is copied once and nothing escapes to the heap. The binding's systems write through
  their own buffers (`ECS.CmdBuf(runnable)`), which goke syncs with every other.
- **Services** (`effect.Runtime`, `store.Facts`) are interfaces where the path is dominated by
  a lookup anyway (`Seek`) or rare.

A change here is measured in gram's `bench/behaviour_test.go` (plans, rules, effects a tick),
baseline and working tree alternately.
