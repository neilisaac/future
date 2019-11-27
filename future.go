package future

import (
	"context"
	"fmt"
	"reflect"
)

// Future is the consumer interface for futures.
type Future interface {
	Done() <-chan struct{}
	Err() error
	Value() interface{}
	Result() (interface{}, error)
	Wait(context.Context) (interface{}, error)
	Then(interface{}) Future
	Catch(func(error)) Future
}

// SettableFuture is the complete interface, including the Set method.
// It may be downcast to a Future.
// This is sometimes referred to as a promise, but that would be
// confusing in a library called future.
type SettableFuture struct {
	value interface{}
	err   error
	done  chan struct{}
}

// New creates a new Future that can be Set or waited on.
func New() *SettableFuture {
	return &SettableFuture{
		value: nil,
		err:   nil,
		done:  make(chan struct{}),
	}
}

// Set provides the value or error associated for a Future.
// Set may only be called once, or it will panic.
func (f *SettableFuture) Set(value interface{}, err error) {
	f.value = value
	f.err = err
	close(f.done)
}

// Done returns a channel which is closed when the result is set.
// This mimics the context.Context interface.
func (f *SettableFuture) Done() <-chan struct{} {
	return f.done
}

// Err returns an error after a value is set.
// This is useful in cases where the result value is not needed.
// This mimics the context.Context interface.
func (f *SettableFuture) Err() error {
	return f.err
}

// Value returns the value after a value is set.
func (f *SettableFuture) Value() interface{} {
	return f.value
}

// Result returns the result and error after a value is set.
func (f *SettableFuture) Result() (interface{}, error) {
	return f.value, f.err
}

// Wait blocks until the provided context expires or the Future's value is set,
// then returns the associated value and error.
// It returns the context's Err() value if the context expires first.
func (f *SettableFuture) Wait(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.Done():
		return f.Value(), f.Err()
	}
}

// Then blocks until the Future's value is set, then invokes
// the callback if a nil error was set.
// The callback must be a function accepting a single argument of the Future's value type.
// Then returns after the callback was invoked or skipped.
func (f *SettableFuture) Then(callback interface{}) Future {
	fnType := reflect.TypeOf(callback)
	fnValue := reflect.ValueOf(callback)
	if fnType.Kind() != reflect.Func {
		panic(fmt.Sprintf("callback %s is not a function", callback))
	}
	if fnType.NumIn() != 1 {
		panic(fmt.Sprintf("callback %s does not take exactly one argument", callback))
	}
	if fnType.NumOut() != 0 {
		panic(fmt.Sprintf("callback %s has more than 0 return values", callback))
	}

	<-f.Done()
	if err := f.Err(); err == nil {
		fnValue.Call([]reflect.Value{reflect.ValueOf(f.Value())})
	}
	return f
}

// Catch blocks until the Future's value is set, then invokes
// the callback if a non-nil error was set.
// Catch returns after the callback was invoked or skipped.
func (f *SettableFuture) Catch(callback func(err error)) Future {
	<-f.Done()
	if err := f.Err(); err != nil {
		callback(err)
	}
	return f
}
