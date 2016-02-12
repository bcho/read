package brain

import (
	"testing"
	"time"

	"github.com/SaidinWoT/timespan"
)

func parseTime(layout, str string) time.Time {
	rv, err := time.Parse(layout, str)
	if err != nil {
		panic(err)
	}
	return rv
}

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

func TestGetInPeriod(t *testing.T) {
	var (
		err    error
		span   timespan.Span
		things []string
	)

	brain := NewBrain()
	err = brain.Remember(parseTime("2006-01-02", "2015-01-01"), "t1", "t1")
	if err != nil {
		t.Error(err)
	}
	err = brain.Remember(parseTime("2006-01-02", "2015-02-01"), "t2", "t2")
	if err != nil {
		t.Error(err)
	}
	err = brain.Remember(parseTime("2006-01-02", "2015-02-01"), "t3", "t3")
	if err != nil {
		t.Error(err)
	}
	err = brain.Remember(parseTime("2006-01-02", "2015-02-02"), "t4", "t4")
	if err != nil {
		t.Error(err)
	}

	// Cover everythings
	span = timespan.New(
		parseTime("2006-01-02", "2015-01-01"),
		time.Duration(365*24)*time.Hour,
	)
	things = brain.GetInPeriod(span)
	if len(things) != 4 {
		t.Errorf("expected 4 things, got: %v", things)
	}

	// Cover nothings (from left side)
	span = timespan.New(
		parseTime("2006-01-02", "2014-01-01"),
		time.Duration(365*24)*time.Hour,
	)
	things = brain.GetInPeriod(span)
	if len(things) != 0 {
		t.Errorf("expected 0 things, got: %v", things)
	}

	// Cover nothings (from right side)
	span = timespan.New(
		parseTime("2006-01-02", "2016-01-01"),
		time.Duration(365*24)*time.Hour,
	)
	things = brain.GetInPeriod(span)
	if len(things) != 0 {
		t.Errorf("expected 0 things, got: %v", things)
	}

	// Partial cover
	span = timespan.New(
		parseTime("2006-01-02", "2015-02-01"),
		time.Duration(1)*time.Hour,
	)
	things = brain.GetInPeriod(span)
	if len(things) != 2 {
		t.Errorf("expected 2 things, got: %v", things)
	}
	if things[0] != "t2" {
		t.Errorf("expected 't2', got: %s", things[0])
	}
	if things[1] != "t3" {
		t.Errorf("expected 't3', got: %s", things[1])
	}
}

func TestForget(t *testing.T) {
	var (
		err             error
		retrievedThing  string
		retrievedThings []string
		present         bool
	)

	at := time.Now()
	key := "foo"
	thing := "bar"
	brain := NewBrain()
	err = brain.Remember(at, key, thing)
	if err != nil {
		t.Error(err)
	}

	retrievedThing, present = brain.Get(key)
	if !present {
		t.Errorf("expected true")
	}
	if retrievedThing != thing {
		t.Errorf("expected %s, got: %s", thing, retrievedThing)
	}

	err = brain.Forget(key)
	if err != nil {
		t.Error(err)
	}

	retrievedThing, present = brain.Get(key)
	if present {
		t.Errorf("expected false")
	}
	span := timespan.New(at, time.Duration(1)*time.Hour)
	retrievedThings = brain.GetInPeriod(span)
	if len(retrievedThings) != 0 {
		t.Errorf("expected 0 things, got: %v", retrievedThings)
	}
}

func TestEach(t *testing.T) {
	var err error

	brain := NewBrain()
	err = brain.Remember(parseTime("2006-01-02", "2015-01-01"), "k", "t")
	if err != nil {
		t.Error(err)
	}
	err = brain.Remember(parseTime("2006-01-02", "2015-02-01"), "k", "t")
	if err != nil {
		t.Error(err)
	}
	err = brain.Remember(parseTime("2006-01-02", "2015-02-01"), "k", "t")
	if err != nil {
		t.Error(err)
	}
	err = brain.Remember(parseTime("2006-01-02", "2015-02-02"), "k", "t")
	if err != nil {
		t.Error(err)
	}

	err = brain.Each(func(at time.Time, key, thing string) error {
		if key != "k" || thing != "t" {
			t.Errorf("invalid thing: %s %s", key, thing)
		}

		return nil
	})
	if err != nil {
		t.Error(err)
	}

	// break
	counter := 0
	err = brain.Each(func(at time.Time, key, thing string) error {
		counter++
		if key != "k" || thing != "t" {
			t.Errorf("invalid thing: %s %s", key, thing)
		}

		return EachBreak
	})
	if counter != 1 {
		t.Errorf("expected 1, got: %d", counter)
	}
	if err != EachBreak {
		t.Error(err)
	}
}
