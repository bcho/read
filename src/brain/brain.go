// Use to remember things.
package brain

import (
	"sync"
	"time"
)

type Brain interface {
	Remember(at time.Time, key, thing string) error
	Get(key string) (string, bool)
}

type thing struct {
	At      time.Time
	Payload string
}

type brain struct {
	things     map[string]thing
	thingsLock *sync.RWMutex
}

func NewBrain() *brain {
	return &brain{
		things:     make(map[string]thing),
		thingsLock: &sync.RWMutex{},
	}
}

func (b *brain) Remember(at time.Time, key, payload string) error {
	b.thingsLock.Lock()
	defer b.thingsLock.Unlock()

	t := thing{at, payload}
	b.things[key] = t

	return nil
}

func (b brain) Get(key string) (string, bool) {
	b.thingsLock.RLock()
	defer b.thingsLock.RUnlock()

	thing, present := b.things[key]
	if present {
		return thing.Payload, present
	}

	return "", present
}
