// Use to remember things.
package brain

import (
	"sort"
	"sync"
	"time"

	"github.com/SaidinWoT/timespan"
)

type Brain interface {
	Remember(at time.Time, key, thing string) error
	Get(key string) (string, bool)
	GetInPeriod(period timespan.Span) []string
	Forget(key string) error
}

type thing struct {
	At      time.Time
	Key     string
	Payload string
}

type things []thing

func (t things) Len() int           { return len(t) }
func (t things) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t things) Less(i, j int) bool { return t[i].At.Before(t[j].At) }

type brain struct {
	l *sync.RWMutex

	things      map[string]thing
	periodIndex []thing
}

func NewBrain() *brain {
	return &brain{
		things:      make(map[string]thing),
		periodIndex: []thing{},
		l:           &sync.RWMutex{},
	}
}

func (b *brain) Remember(at time.Time, key, payload string) error {
	b.l.Lock()
	defer b.l.Unlock()

	t := thing{at, key, payload}
	b.things[key] = t
	b.periodIndex = append(b.periodIndex, t)
	sort.Sort(things(b.periodIndex))

	return nil
}

func (b brain) Get(key string) (string, bool) {
	b.l.RLock()
	defer b.l.RUnlock()

	thing, present := b.things[key]
	if present {
		return thing.Payload, present
	}

	return "", present
}

func (b brain) GetInPeriod(period timespan.Span) []string {
	b.l.RLock()
	defer b.l.RUnlock()

	i := sort.Search(
		len(b.periodIndex),
		func(i int) bool {
			return period.ContainsTime(b.periodIndex[i].At)
		},
	)

	rv := []string{}
	for ; i < len(b.periodIndex); i++ {
		if !period.ContainsTime(b.periodIndex[i].At) {
			break
		}
		rv = append(rv, b.periodIndex[i].Payload)
	}

	return rv
}

func (b *brain) Forget(key string) error {
	b.l.Lock()
	defer b.l.Unlock()

	if _, present := b.things[key]; !present {
		return nil
	}

	delete(b.things, key)
	for i := 0; i < len(b.periodIndex); i++ {
		if b.periodIndex[i].Key == key {
			b.periodIndex = append(
				b.periodIndex[:i],
				b.periodIndex[i+1:]...,
			)
			break
		}
	}

	return nil
}
