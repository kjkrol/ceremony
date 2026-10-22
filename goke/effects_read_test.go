package goke_test

import (
	"sync"
	"testing"
	"time"

	"github.com/kjkrol/ceremony/effect"
	cgoke "github.com/kjkrol/ceremony/goke"
	"github.com/kjkrol/ceremony/tag"
	"github.com/kjkrol/goke/v3"
	"github.com/kjkrol/uid"
)

func TestEffectsRead_ForksReadWhatTheEffectsDoOnGoroutinesOfTheirOwn(t *testing.T) {
	next := effect.Changed
	reg := effect.New(func(string) tag.Tag[effect.States] { t := next; next++; return t })
	reg.Define("burn", effect.Spec{effect.Lasts(time.Minute)})
	burn := reg.Named("burn")
	fx := cgoke.NewEffects(reg)
	early := fx.Read().Fork() // before the effects' system is set up: its queries come then
	ecs := goke.New()
	var hp goke.Comp[health]
	var ids []uid.UID64
	ecs.Setup(goke.SystemFn{OnInit: func(si *goke.SysInit) {
		f := si.NewFactory(&hp)
		f.Create(64)
		for f.Next() {
			ids = append(ids, f.Cursor.IDs...)
		}
	}})
	fx.Module().RegSystems(ecs)
	casts := ecs.RegSys(goke.SystemFn{OnUpdate: func(*goke.CmdBuf, time.Duration) {
		for i, id := range ids {
			if i%3 == 0 {
				fx.Cast(id, burn)
			}
		}
	}})
	ecs.SetPlan(func(ctx goke.RunCtx, d time.Duration) {
		ctx.Run(casts, d)
		fx.Module().RunPlan(ctx, d)
	})
	ecs.Tick(100 * time.Millisecond)
	ecs.Tick(100 * time.Millisecond)
	want := make([]bool, len(ids))
	for i, id := range ids {
		want[i] = fx.Has(id, burn)
	}
	reads := []*cgoke.EffectsRead{early}
	for range 7 {
		reads = append(reads, fx.Read().Fork())
	}
	got := make([][]bool, len(reads))
	var wg sync.WaitGroup
	for g, r := range reads {
		wg.Go(func() {
			got[g] = make([]bool, len(ids))
			for range 50 {
				for i, id := range ids {
					got[g][i] = r.Has(id, burn)
				}
			}
		})
	}
	wg.Wait()
	burning := 0
	for i := range ids {
		if want[i] {
			burning++
		}
		for g := range reads {
			if got[g][i] != want[i] {
				t.Fatalf("fork %d read entity %d as burning %v, the effects %v", g, i, got[g][i],
					want[i])
			}
		}
	}
	if burning == 0 || burning == len(ids) {
		t.Errorf("%d of %d burning: the test proves nothing", burning, len(ids))
	}
}
