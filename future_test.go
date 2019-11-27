package future

import (
	"context"
	"errors"
	"testing"
)

func TestFutureBasics(t *testing.T) {
	f := New()
	select {
	case <-f.Done():
		t.Fatal("future should not be done after New")
	default:
	}

	f.Set(1, nil)
	select {
	case <-f.Done():
	default:
		t.Fatal("future should be done after Set")
	}

	if value := f.Value(); value != 1 {
		t.Errorf("wrong value returned: %#v", value)
	}

	if err := f.Err(); err != nil {
		t.Errorf("error should be nil: %#v", err)
	}

	if value, err := f.Result(); value != 1 || err != nil {
		t.Errorf("wrong result from Result call: %#v, %#v", value, err)
	}

	if value, err := f.Wait(context.Background()); value != 1 || err != nil {
		t.Errorf("wrong result from Wait call: %#v, %#v", value, err)
	}
}

func TestFutureWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var f Future = New()
	value, err := f.Wait(ctx)
	if value != nil {
		t.Errorf("value is not nil: %#v", value)
	}
	if err != context.Canceled {
		t.Errorf("error is not context.Canceled: %#v", err)
	}
}

func TestFutureThen(t *testing.T) {
	f := New()
	f.Set(1, nil)

	var got int

	f.Then(func(value int) {
		got = value
	}).Catch(func(err error) {
		t.Errorf("Catch called with %#v", err)
	})

	if got != 1 {
		t.Errorf("Then not called with 1: %d", got)
	}
}

func TestFutureCatch(t *testing.T) {
	f := New()
	f.Set(nil, errors.New("bad"))

	var got error

	f.Then(func(value interface{}) {
		t.Errorf("Then called with: %#v", value)
	}).Catch(func(err error) {
		got = err
	})

	if got.Error() != "bad" {
		t.Errorf("Catch called with incorrect error: %#v", got)
	}
}

func TestThenBadCallback(t *testing.T) {
	f := New()
	f.Set(1, nil)

	assertPanics(t, func() {
		f.Then(1)
	})

	assertPanics(t, func() {
		f.Then(func() {})
	})

	assertPanics(t, func() {
		f.Then(func(a, b int) {})
	})

	assertPanics(t, func() {
		f.Then(func(a int) int { return a })
	})
}

func assertPanics(t *testing.T, callback func()) (value interface{}) {
	defer func() {
		value = recover()
		if value == nil {
			t.Errorf("function did not panic")
		}
	}()

	callback()
	return
}
