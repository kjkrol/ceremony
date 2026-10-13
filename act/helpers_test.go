package act_test

import (
	"github.com/kjkrol/ceremony"
	"github.com/kjkrol/ceremony/cond"
	"github.com/kjkrol/ceremony/tag"
)

var testTags = map[string]tag.Tag[cond.Roles]{}

// Role is the role named name, one tag a name.
func Role(name string) *ceremony.Part {
	t, ok := testTags[name]
	if !ok {
		t = tag.Tag[cond.Roles](len(testTags))
		testTags[name] = t
	}
	return ceremony.NewPart(name, t)
}
