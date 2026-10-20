package command_test

import (
	"reflect"
	"testing"

	"github.com/kjkrol/ceremony/command"
	"github.com/kjkrol/uid"
)

type (
	walk   struct{ Far uint64 }
	sent   struct{ As walk }
	wave   struct{}
	sing   struct{}
	loose  struct{ got []command.Command }
	issued = command.Issued[walk]
)

func (walk) Command() {}
func (sent) Command() {}
func (wave) Command() {}
func (sing) Command() {}

func (s sent) Routed() command.Command { return s.As }

func (*loose) Accepts() reflect.Type                     { return reflect.TypeFor[wave]() }
func (q *loose) Put(_ command.Issuer, c command.Command) { q.got = append(q.got, c) }
func (q *loose) PutFrom(_ uid.UID64, c command.Command)  { q.got = append(q.got, c) }
func (q *loose) Empty() bool                             { return len(q.got) == 0 }

func TestCarrier_PutFromOfGoesWherePutFromDoes(t *testing.T) {
	var c command.Carrier
	var walks command.Queue[walk]
	var waves loose
	if err := c.Carry(&walks, &waves); err != nil {
		t.Fatal(err)
	}
	const me = uid.UID64(1 << 40)
	var erased command.Command = walk{Far: 3}
	for i, ok := range []bool{
		c.PutFromOf(me, walk{Far: 1}),
		c.PutFromOf(me, sent{As: walk{Far: 2}}),
		c.PutFromOf(me, erased),
		c.PutFromOf(me, wave{}),
	} {
		if !ok {
			t.Errorf("command %d was not taken", i)
		}
	}
	if c.PutFromOf(me, sing{}) {
		t.Error("a command no queue takes was taken")
	}
	var none *command.Carrier
	if none.PutFromOf(me, walk{}) {
		t.Error("no carrier took a command")
	}
	var got []issued
	walks.Drain(func(i issued) { got = append(got, i) })
	want := []issued{
		{Entity: me, ByEntity: true, Command: walk{Far: 1}},
		{Entity: me, ByEntity: true, Command: walk{Far: 2}},
		{Entity: me, ByEntity: true, Command: walk{Far: 3}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("walks got %v, want %v", got, want)
	}
	if len(waves.got) != 1 || waves.got[0] != (wave{}) {
		t.Errorf("waves got %v", waves.got)
	}
	put := func() {
		c.PutFromOf(me, walk{Far: 1 << 20})
		walks.Drain(func(issued) {})
	}
	put()
	if n := testing.AllocsPerRun(100, put); n != 0 {
		t.Errorf("a command into its Queue allocated %v times", n)
	}
}
