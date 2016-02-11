package brain

import (
	"testing"
	"time"
)

func TestNewBrain(t *testing.T) {
	brain := NewBrain()
	if brain == nil {
		t.Errorf("create brain failed")
	}
}

func TestRememberAndGet(t *testing.T) {
	brain := NewBrain()
	expectedKey := "foo"
	expectedThing := "bar"
	newThing := "baz"

	var (
		err     error
		present bool
		thing   string
	)

	// Remember then get
	err = brain.Remember(time.Now(), expectedKey, expectedThing)
	if err != nil {
		t.Error(err)
	}
	thing, present = brain.Get(expectedKey)
	if !present {
		t.Errorf("unexpected false")
	}

	if thing != expectedThing {
		t.Errorf("expected %s, got %s", expectedThing, thing)
	}

	// Get for missing
	_, present = brain.Get("invalid-key")
	if present {
		t.Errorf("unexpected true")
	}

	// Override key
	err = brain.Remember(time.Now(), expectedKey, newThing)
	if err != nil {
		t.Error(err)
	}

	thing, present = brain.Get(expectedKey)
	if !present {
		t.Errorf("unexpected false")
	}

	if thing != newThing {
		t.Errorf("expected %s, got %s", newThing, thing)
	}
}
