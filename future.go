package future

import (
	"context"
	"fmt"
)

// Future is the consumer interface for futures.
type Future[T any] interface {
	Done() <-chan struct{}
	Err() error
	Value() T
	Result() (T, error)
	Wait(context.Context) (T, error)
	Then(func(T) (T, error)) Future[T]
	Catch(func(error)) Future[T]
}

// SettableFuture is the complete interface, including the Set method.
// It may be downcast to a Future.
// This is sometimes referred to as a promise, but that would be
// confusing in a library called future.
type SettableFuture[T any] struct {
	value T
	err   error
	done  chan struct{}
}

// New creates a new Future that can be Set or waited on.
func New[T any]() *SettableFuture[T] {
	return &SettableFuture[T]{
		done: make(chan struct{}),
	}
}

// Set provides the value or error associated for a Future.
// Set may only be called once, or it will panic.
func (f *SettableFuture[T]) Set(value T, err error) Future[T] {
	f.value = value
	f.err = err
	close(f.done)
	return f
}

// Done returns a channel which is closed when the result is set.
// This mimics the context.Context interface.
func (f *SettableFuture[T]) Done() <-chan struct{} {
	return f.done
}

// Err returns an error after a value is set.
// This is useful in cases where the result value is not needed.
// This mimics the context.Context interface.
func (f *SettableFuture[T]) Err() error {
	return f.err
}

// Value returns the value after a value is set.
func (f *SettableFuture[T]) Value() T {
	return f.value
}

// Result returns the result and error after a value is set.
func (f *SettableFuture[T]) Result() (T, error) {
	return f.value, f.err
}

// Wait blocks until the provided context expires or the Future's value is set,
// then returns the associated value and error.
// It returns the context's Err() value if the context expires first.
func (f *SettableFuture[T]) Wait(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var nilValue T
		return nilValue, ctx.Err()
	case <-f.Done():
		return f.Result()
	}
}

// Then blocks until the Future's value is set, then either
// returns the existing future if an error was set, or
// invokes the callback and returns a new Future wrapping its result.
func (f *SettableFuture[T]) Then(callback func(T) (T, error)) Future[T] {
	<-f.Done()
	if err := f.Err(); err != nil {
		return f
	}
	return New[T]().Set(callback(f.Value()))
}

// Catch blocks until the Future's value is set, then invokes
// the callback if a non-nil error was set.
// Catch returns after the callback was invoked or skipped.
func (f *SettableFuture[T]) Catch(callback func(err error)) Future[T] {
	<-f.Done()
	if err := f.Err(); err != nil {
		callback(err)
	}
	return f
}

// String converts the settable future to a string representation,
// otherwise most testing frameworks will throw a data race
func (f *SettableFuture[T]) String() string {
	return fmt.Sprintf("SettableFuture<%p>", f)
}
