# glob.[go](https://golang.org)

[![GoDoc][godoc-image]][godoc-url] [![Build Status][travis-image]][travis-url]

> Simple globbing library.

## Install

```shell
    go get github.com/gobwas/glob
```

## Example

```go

package main

import "github.com/gobwas/glob"

func main() {
    var g glob.Glob
    
    // create simple glob
    g = glob.New("*.github.com")
    g.Match("api.github.com") // true
    
    // create new glob with set of delimiters as ["."]
    g = glob.New("api.*.com", ".")
    g.Match("api.github.com") // true
    g.Match("api.gi.hub.com") // false
    
    // create new glob with set of delimiters as ["."]
    // but now with super wildcard
    g = glob.New("api.**.com", ".")
    g.Match("api.github.com") // true
    g.Match("api.gi.hub.com") // true
        
    // create glob with single symbol wildcard
    g = glob.New("?at")
    g.Match("cat") // true
    g.Match("fat") // true
    g.Match("at") // false
    
    // create glob with single symbol wildcard and delimiters ["f"]
    g = glob.New("?at", "f")
    g.Match("cat") // true
    g.Match("fat") // false
    g.Match("at") // false 
}

```

## Performance

In comparison with [go-glob](https://github.com/ryanuber/go-glob), it is ~2.5x faster (on my personal Mac),
because my impl compiles patterns for future usage. If you will not use compiled `glob.Glob` object,
and do `g := glob.New(pattern); g.Match(...)` every time, then your code will be about ~3x slower.

Run `go test bench=.` from source root to see the benchmarks:

Test | Operations | Speed
-----|------------|------
github.com/gobwas/glob | 20000000 | 150 ns/op
github.com/ryanuber/go-glob | 10000000 | 375 ns/op

Also, there are few simple optimizations, that help to test much faster patterns like `*abc`, `abc*` or `a*c`:

Test | Operations | Speed
-----|------------|------
prefix | 200000000 | 8.78 ns/op
suffix | 200000000 | 9.46 ns/op
prefix-suffix | 100000000 | 16.3 ns/op

[godoc-image]: https://godoc.org/github.com/gobwas/glob?status.svg
[godoc-url]: https://godoc.org/github.com/gobwas/glob
[travis-image]: https://travis-ci.org/gobwas/glob.svg?branch=master
[travis-url]: https://travis-ci.org/gobwas/glob