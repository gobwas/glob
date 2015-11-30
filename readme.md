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