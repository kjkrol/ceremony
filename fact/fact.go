// Package fact is what happens to an entity: a fact is told to it and carried until it goes,
// a loop of a plan begins on it or waits for it; a message is a fact with its sender.
package fact

import (
	"time"

	"github.com/kjkrol/uid"
)

// Fact marks a type as a fact an entity carries: a component that comes and goes, told by
// whoever perceives. A type is one by its method Fact, a mark never called, confirmed where it
// is defined: var _ fact.Fact = Arrived{}.
type Fact interface{ Fact() }

// Subject is a fact or a moment about another entity, whom an Aimed command or a message
// given on it goes to; false for nobody just now.
type Subject interface{ Subject() (uid.UID64, bool) }

// MaxChain is how many times a message of one type may be passed on, the first sender's
// counted.
const MaxChain = 4

// AskLife is how long a message waits to be taken up before it is dropped.
const AskLife = 3 * time.Second

// From is what a message carries of its sender. A command embedding it, given with
// plan.Other.Do, lands on that entity as a fact in place of going to a queue, the engine
// filling in who sent it; one without a plan takes none. A loop of the command's type answers
// it, plan.Other.Do again.
type From struct {
	By   uid.UID64
	At   time.Duration // on the clock
	Hops uint8         // how many times the message was passed on
}

// Fact: a message is a fact of whom it is sent to.
func (From) Fact() {}

// Subject is the sender.
func (f From) Subject() (uid.UID64, bool) { return f.By, true }

// Message is the sender's part of a command embedding From, for the engine to fill in.
func (f *From) Message() *From { return f }

// Sent is the sender's part as a fact carried reads it.
func (f From) Sent() From { return f }

// Message is a command embedding From: sent to the other as a fact in place of a queue.
type Message interface{ Message() *From }

// Sent is a fact that is a message, read where it landed.
type Sent interface{ Sent() From }
