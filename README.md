# future

The `future` package aims to provide a simple interface for consuming return values and/or errors from a `Future`.

It provides the `SettableFuture` struct and a `Future` interface for consumers.
This distinction allows explicit decoupling of value providers and consumers.

The following example demonstrates the interfaces:

```go
settableFuture := future.New[int]()
settableFuture.Set(1, nil)

var f future.Future = settableFuture
<-f.Done()
fmt.Println(f.Value())  // prints 1
fmt.Println(f.Err())    // prints <nil>
fmt.Println(f.Result()) // prints 1 <nil>
fmt.Println(f.Wait(context.Background())) // prints 1 <nil>

f.Then(func(i int) {
    fmt.Println(i) // prints 1
}).Catch(func (e error) {
    fmt.Println(e) // doesn't run
})

```

See [godoc](https://godoc.org/github.com/neilisaac/future) for full package documentation.
