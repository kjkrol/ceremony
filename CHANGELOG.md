# Changelog

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
