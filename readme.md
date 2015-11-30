# glob.[go](https://golang.org)

Simple globbing library.

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

In comparison with [go-glob](https://github.com/ryanuber/go-glob), it is ~2.7x faster (on my personal Mac),
because my impl compiles patterns for future usage. If you will not use compiled `glob.Glob` object,
and do `g := glob.New(pattern); g.Match(...)` every time, then your code will be about ~3x slower.

Run `go test bench=.` from source root to see the benchmarks:

Test | Operations | Speed
-----|------------|------
github.com/gobwas/glob | 20000000 | 165 ns/op
github.com/ryanuber/go-glob | 10000000 | 452 ns/op