# CLAUDE.md

Guidance for Claude Code working in this repository.

## What this is

**ceremony** (module `github.com/kjkrol/ceremony`) is a language and an engine for the behaviour
of things that have state and are told what happens to them: entities carry facts and states,
play roles, obey rules at the moments a pass catches, run plans over time on their minds, and ask
for things with commands. README.md has the dictionary; `doc/ceremony.md` the language. The
store of state is goke, an ECS, for now: facts, states, roles and minds are components. A store
interface — where a fact, a state, a role and the clock come from — is the next step; the plans'
runner (`internal/steps.Plans`, reached through `act.Planned.Define`) is where it will go.

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
effect           Effects, Effect, Spec (Lasts, Stacking, Then, Grant, Alter, Described), Knob,
                 Active, Wide, States, Changed, the effects' system
command          Command, Issuer, Nobody, Issued, Queue, CommandQueue, Carrier, Routed, Unwrap
host             Rules, PairRules, StepRules, Tick, TickSource, Columns, Own, Side, SideOf,
                 Marks, About, Met, Placed, Aimed, Hosted, Host, ErrUnhosted, ErrHostBuilt
cond             Cond, Subject (Self, Other, Around, On), If, Not, Chance, Always, Word,
                 NewWord, Roles, Role
act              Branch, Subject (Self, Other, Around), Step, Steps, OneOf, Rule, Build,
                 Lasting, Planned, NewPlan, and the bridges for plan
plan             Loop, Part, Steps, OneOf, Entity (Self with Wait and Idle), Other, New
ceremony (root)  Part, NewPart, Plays, Played, Deliver, Command, Put, Take, Toggle, Trigger,
                 Target, Router, Label, LabelOf, Whom, Named, Group, ID, World
internal/steps   the engine: trees of steps, Instant (a rule's run), Plans and Mind (a plan's)
```

Imports go one way:

- `plan → act → cond → effect → tag`; `cond → fact`;
- `host → fact, command, effect, tag`;
- `act → host, command, comp, internal/steps`;
- the root `ceremony → act, plan, cond, effect, command, comp, host`;
- `internal/steps → command, effect, fact, goke`; `comp → tag, goke`; `effect → tag, goke`.

External: `goke/v3`, `uid`. **Nothing imports gram or anything of a game**: no `render`, no
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
`plan/toys_test.go` a rig of commands and facts on a bare goke ECS. **A new semantics test is
checked with its fix disabled**: before a test of a behaviour is kept, the code that gives the
behaviour is taken out (a guard — `GuardCame`, a `Timeout` — made to pass) and the test must
fail; a test that passes either way tests nothing.

## The store's part, for now

`plan/store.go` aliases the engine's `Mind`, `Plans`, `NewPlans`, `PlanID` and `MaxSteps` from
`internal/steps`: a user over goke needs them to run plans (the world registers `Mind` for the
saves, finds minds, and runs `Plans.System`). They are aliases, the one exception to "no type
aliases", because a component must be one type for goke and the saves; they move to the store
binding (`ceremony/goke`) as it is made, with `act.Planned.Define`.
