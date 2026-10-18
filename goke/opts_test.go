package goke_test

import (
	"testing"
	"time"

	"github.com/kjkrol/ceremony/command"
	"github.com/kjkrol/ceremony/cond"
	"github.com/kjkrol/ceremony/effect"
	cgoke "github.com/kjkrol/ceremony/goke"
	"github.com/kjkrol/ceremony/internal/steps"
	"github.com/kjkrol/ceremony/plan"
	"github.com/kjkrol/ceremony/store"
	"github.com/kjkrol/ceremony/tag"
	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

type (
	far  struct{}
	ping struct{}
)

func (ping) Command() {}

// one entity of comps near, one with far besides
func nearAndFar(si *goke.SysInit, comps ...goke.Addable) (near, away uid.UID64) {
	var mark goke.Comp[far]
	f := si.NewFactory(comps...)
	f.Create(1)
	f.Next()
	near = f.Cursor.IDs[0]
	f = si.NewFactory(append(comps, &mark)...)
	f.Create(1)
	f.Next()
	return near, f.Cursor.IDs[0]
}

func TestPlans_GoOverTheEntitiesTheirOptionsLeave(t *testing.T) {
	var carrier command.Carrier
	var pings command.Queue[ping]
	if err := carrier.Carry(&pings); err != nil {
		t.Fatal(err)
	}
	ps := plan.NewPlans(func() time.Duration { return 0 }, 0, nil, &carrier)
	ps.Register(plan.New("ping", plan.Loop[any]{
		Name: "ping",
		When: cond.Self.If(cond.Always),
		Then: plan.Self.Do(ping{}),
	}))
	ecs := goke.New()
	var mind goke.Comp[store.Mind]
	var near, away uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		near, away = nearAndFar(si, &mind)
		q := si.NewQueryBuilder(&mind).Build()
		for q.All(); q.Next(); {
			for i := range mind.Slice(q.Cursor()) {
				mind.Slice(q.Cursor())[i] = store.Mind{Plan: steps.PlanID("ping")}
			}
		}
	}})
	run := ecs.RegSys(cgoke.NewPlans(ps, goke.Exclude[far]()))
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(run, d)
		ctx.Sync()
	})
	given := map[uid.UID64]int{}
	for range 3 {
		ecs.Tick(100 * time.Millisecond)
		pings.Drain(func(i command.Issued[ping]) { given[i.Entity]++ })
	}
	if given[near] == 0 || given[away] != 0 {
		t.Errorf("pings: the near entity %d, the far %d; want the near's plan alone going",
			given[near], given[away])
	}
}

func TestEffects_CountDownByTheStepTheirPaceGives(t *testing.T) {
	next := effect.Changed
	reg := effect.New(func(string) tag.Tag[effect.States] { t := next; next++; return t })
	reg.Define("burn", effect.Spec{effect.Lasts(time.Second)})
	burn := reg.Named("burn")
	var step uint64
	fx := cgoke.NewEffects(reg, goke.PacedBy(func() uint64 { return step }, goke.Every[far](2)))
	ecs := goke.New()
	var hp goke.Comp[health]
	var active goke.Comp[cgoke.Active]
	var q *goke.Query
	var nearID, awayID uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		nearID, awayID = nearAndFar(si, &hp)
		q = si.NewQueryBuilder(&active).Build()
	}})
	fx.Module().RegSystems(ecs)
	cast := true
	casts := ecs.RegSys(goke.SystemFn{OnUpdate: func(*goke.CmdBuf, time.Duration) {
		if cast {
			fx.Cast(nearID, burn)
			fx.Cast(awayID, burn)
			cast = false
		}
	}})
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(casts, d)
		fx.Module().RunPlan(ctx, d)
		step++
	})
	left := func() map[uid.UID64]time.Duration {
		got := map[uid.UID64]time.Duration{}
		for q.All(); q.Next(); {
			cur := q.Cursor()
			for i, id := range cur.IDs {
				got[id], _ = active.Slice(cur)[i].Left(burn)
			}
		}
		return got
	}
	for range 4 {
		ecs.Tick(100 * time.Millisecond)
	}
	var near, far []time.Duration
	for range 2 {
		before := left()
		ecs.Tick(100 * time.Millisecond)
		after := left()
		near = append(near, before[nearID]-after[nearID])
		far = append(far, before[awayID]-after[awayID])
	}
	if near[0] != 100*time.Millisecond || near[1] != 100*time.Millisecond {
		t.Errorf("the near entity counted %v off by step, want 100ms each", near)
	}
	if far[0]+far[1] != 200*time.Millisecond || far[0] != 0 && far[1] != 0 {
		t.Errorf("the far entity going every second step counted %v off, want 0 and 200ms", far)
	}
}
