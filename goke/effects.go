package goke

import (
	"time"

	"github.com/kjkrol/ceremony/effect"
	"github.com/kjkrol/ceremony/tag"
	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

// Effects runs a register of effects over goke: the register itself, to define and name them,
// and their running on the entities (effect.Runtime), a system once a step counting them down,
// granting, altering and restoring.
type Effects struct {
	*effect.Effects
	system *effectSystem
	module *module
}

// NewEffects runs the effects of reg.
func NewEffects(reg *effect.Effects) *Effects {
	e := &Effects{Effects: reg}
	e.system = newEffectSystem(reg)
	e.module = &module{system: e.system}
	return e
}

var _ effect.Runtime = (*Effects)(nil)

// Cast puts ef on id for as long as its Spec says, landing with the effects' next pass; cast
// again, it is refreshed unless it stacks.
func (e *Effects) Cast(id uid.UID64, ef effect.Effect) { e.CastFor(id, ef, e.Effects.Lasts(ef)) }

// CastFor is Cast for a set time of the clock's; effect.Forever for one that lasts until
// Dispel.
func (e *Effects) CastFor(id uid.UID64, ef effect.Effect, d time.Duration) {
	e.system.cast(id, ef.ID(), d)
}

// Dispel ends ef on id with the effects' next pass.
func (e *Effects) Dispel(id uid.UID64, ef effect.Effect) { e.system.dispel(id, ef.ID()) }

// Has reports whether id is under ef.
func (e *Effects) Has(id uid.UID64, ef effect.Effect) bool { return e.system.has(id, ef.ID()) }

// Module is the effects as a goke.Module, for the world to install.
func (e *Effects) Module() goke.Module { return e.module }

type module struct {
	system   *effectSystem
	runnable goke.Runnable
}

var _ goke.Module = (*module)(nil)

// once, however many times the module is registered; the system writes through its own buffer
func (m *module) RegSystems(ecs *goke.ECS) {
	if m.runnable == nil {
		m.runnable = ecs.RegSys(m.system)
		m.system.cb = ecs.CmdBuf(m.runnable)
	}
}

func (m *module) RunPlan(ctx goke.RunCtx, d time.Duration) {
	ctx.Run(m.runnable, d)
	ctx.Sync()
}

func (m *module) SetupSystems() []goke.System { return nil }

func (m *module) LoadComps() []goke.CompToken {
	return []goke.CompToken{goke.LoadComp[Active](), goke.LoadComp[Wide](),
		goke.LoadComp[tag.Tags[effect.States]]()}
}

func (m *module) Persisted() []any { return []any{&m.system.originals.byEntity} }
