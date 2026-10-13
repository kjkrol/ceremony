// Package hosts hands rules to the hosts of the moments they are of.
package ceremony

import (
	"errors"
	"fmt"

	"github.com/kjkrol/ceremony/act"
	"github.com/kjkrol/ceremony/host"
)

// Deliver hands each rule, a role's each of its own, to the first of among that takes it, the
// next tried while one refuses with host.ErrUnhosted. A rule none takes is host.ErrUnhosted
// naming the rule; any other error stops at once.
func Deliver(among []host.Host, rules ...act.Rule) error {
	for _, r := range rules {
		if role, ok := r.(*Part); ok {
			if err := Deliver(among, role.Rules()...); err != nil {
				return err
			}
			continue
		}
		if err := deliver(among, r); err != nil {
			return err
		}
	}
	return nil
}

func deliver(among []host.Host, r act.Rule) error {
	for _, h := range among {
		err := h.Add(r)
		if err == nil || !errors.Is(err, host.ErrUnhosted) {
			return err
		}
	}
	return fmt.Errorf("%w: no plugin in use catches the moment of the rule %v", host.ErrUnhosted, r)
}
