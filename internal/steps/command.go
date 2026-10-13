package steps

import (
	"fmt"
	"reflect"

	"github.com/kjkrol/ceremony/command"
	"github.com/kjkrol/ceremony/fact"
	"github.com/kjkrol/uid"
)

// NewOrder gives the entity the command cmd, or sends it as a message where cmd carries a From.
func NewOrder[C command.Command](cmd C) Step {
	return leaf{sign: "order[" + typeName[C]() + "]",
		make: func() exec { return &order[C]{cmd: cmd} }}
}

// Aimed is a command told the subject of the fact or moment it stands under as it is issued.
type Aimed interface{ Aim(who uid.UID64) }

type order[C command.Command] struct {
	basic
	cmd     C
	carried *carried[C] // a message's
}

func (*order[C]) instant() {}

func (i *order[C]) sends() bool {
	var v C
	_, ok := any(&v).(fact.Message)
	return ok
}

func (i *order[C]) refuses(any) string {
	if i.sends() {
		return "sends a message, which only a plan can"
	}
	return ""
}

func (i *order[C]) bind(s *Plans) {
	if i.sends() {
		i.carried = carriedOf[C](s)
	}
}

func (i *order[C]) tick(c *ctx, _ int, _ []int) Status {
	if !c.entity {
		return Failure
	}
	v := i.cmd
	if m, ok := any(&v).(fact.Message); ok {
		return i.send(c, &v, m.Message())
	}
	if a, ok := any(&v).(Aimed); ok && c.about {
		a.Aim(c.subject)
	} else if ok && c.names {
		return Failure
	}
	commands := c.pass.Commands
	if c.sys != nil {
		commands = c.sys.commands
	}
	if !commands.PutFrom(c.id, v) {
		panic(fmt.Sprintf("rule: no plugin handles %T — Use the one that defines it", v))
	}
	return Success
}

// to whom the branch acts on, never to the sender itself; a message passed on keeps count of
// its hops, one of another type begins its own count
func (i *order[C]) send(c *ctx, v *C, from *fact.From) Status {
	if c.id == c.me {
		return Failure
	}
	var hops uint8
	if c.hopsOf == reflect.TypeFor[C]() {
		hops = c.hops + 1
	}
	if hops >= fact.MaxChain {
		return Failure
	}
	*from = fact.From{By: c.me, At: c.now, Hops: hops}
	if !i.carried.putOn(c, c.id, *v) {
		return Failure
	}
	return Success
}
