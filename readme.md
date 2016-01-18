# glob.[go](https://golang.org)

[![GoDoc][godoc-image]][godoc-url] [![Build Status][travis-image]][travis-url]

> Go Globbing Library.

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
    g = glob.MustCompile("*.github.com")
    g.Match("api.github.com") // true
    
    // create new glob with set of delimiters as ["."]
    g = glob.MustCompile("api.*.com", ".")
    g.Match("api.github.com") // true
    g.Match("api.gi.hub.com") // false
    
    // create new glob with set of delimiters as ["."]
    // but now with super wildcard
    g = glob.MustCompile("api.**.com", ".")
    g.Match("api.github.com") // true
    g.Match("api.gi.hub.com") // true
        
    // create glob with single symbol wildcard
    g = glob.MustCompile("?at")
    g.Match("cat") // true
    g.Match("fat") // true
    g.Match("at") // false
    
    // create glob with single symbol wildcard and delimiters ["f"]
    g = glob.MustCompile("?at", "f")
    g.Match("cat") // true
    g.Match("fat") // false
    g.Match("at") // false 
    
    // create glob with character-list matchers 
    g = glob.MustCompile("[abc]at")
    g.Match("cat") // true
    g.Match("bat") // true
    g.Match("fat") // false
    g.Match("at") // false
    
    // create glob with character-list matchers 
    g = glob.MustCompile("[!abc]at")
    g.Match("cat") // false
    g.Match("bat") // false
    g.Match("fat") // true
    g.Match("at") // false 
    
    // create glob with character-range matchers 
    g = glob.MustCompile("[a-c]at")
    g.Match("cat") // true
    g.Match("bat") // true
    g.Match("fat") // false
    g.Match("at") // false
    
    // create glob with character-range matchers 
    g = glob.MustCompile("[!a-c]at")
    g.Match("cat") // false
    g.Match("bat") // false
    g.Match("fat") // true
    g.Match("at") // false 
    
    
    // create glob with pattern-alternatives list 
    g = glob.MustCompile("{cat,bat,[fr]at}")
    g.Match("cat") // true
    g.Match("bat") // true
    g.Match("fat") // true
    g.Match("rat") // true
    g.Match("at") // false 
    g.Match("zat") // false 
}

```

## Performance

This library is created for compile-once patterns. This means, that compilation could take time, but 
strings matching is done faster, than in case when always parsing template.

If you will not use compiled `glob.Glob` object, and do `g := glob.MustCompile(pattern); g.Match(...)` every time, then your code will be much more slower.

Run `go test -bench=.` from source root to see the benchmarks:

Pattern | Fixture | Operations | Speed (ns/op)
--------|---------|------------|--------------
`[a-z][!a-x]*cat*[h][!b]*eyes*` | | 50000 | 26497
`[a-z][!a-x]*cat*[h][!b]*eyes*` | `my cat has very bright eyes` | 2000000 | 615
`https://*.google.*` | `https://account.google.com` | 10000000 | 121
`{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}` | `http://yahoo.com` | 10000000 | 167
`{https://*gobwas.com,http://exclude.gobwas.com}` | `https://safe.gobwas.com` | 50000000 | 24.7 
`abc*` | `abcdef` | 200000000 | 9.49
`*def` | `abcdef` | 200000000 | 9.60 ns/op
`ab*ef` | `abcdef` | 100000000 | 15.2

[godoc-image]: https://godoc.org/github.com/gobwas/glob?status.svg
[godoc-url]: https://godoc.org/github.com/gobwas/glob
[travis-image]: https://travis-ci.org/gobwas/glob.svg?branch=master
[travis-url]: https://travis-ci.org/gobwas/glob