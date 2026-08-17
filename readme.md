# glob.[go](https://golang.org)

[![GoDoc][godoc-image]][godoc-url] [![CI][ci-image]][ci-url]

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
    var g *glob.Pattern

    // create simple glob
    g = glob.MustCompile("*.github.com")
    g.Match("api.github.com") // true

    // quote meta characters and then create simple glob
    g = glob.MustCompile(glob.QuoteMeta("*.github.com"))
    g.Match("*.github.com") // true

    // create new glob with set of delimiters as ["."]
    g = glob.MustCompile("api.*.com", '.')
    g.Match("api.github.com") // true
    g.Match("api.gi.hub.com") // false

    // create new glob with set of delimiters as ["."]
    // but now with super wildcard
    g = glob.MustCompile("api.**.com", '.')
    g.Match("api.github.com") // true
    g.Match("api.gi.hub.com") // true

    // create glob with single symbol wildcard
    g = glob.MustCompile("?at")
    g.Match("cat") // true
    g.Match("fat") // true
    g.Match("at") // false

    // create glob with single symbol wildcard and delimiters ['f']
    g = glob.MustCompile("?at", 'f')
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

`Compile` reports malformed patterns with a `*glob.SyntaxError` carrying the
byte offset and the reason:

```go
_, err := glob.Compile("{a,b")
// err: glob: syntax error at 4: unclosed `{`
```

A compiled `Pattern` captures what it was compiled from, so it can be passed
around instead of the raw arguments and inspected when needed (`String()` makes
it a `fmt.Stringer`, like `regexp.Regexp`):

```go
g := glob.MustCompile("*.github.com", '.')
g.String()     // "*.github.com"
g.Separators() // []rune{'.'}
```

## Syntax

Syntax is inspired by [standard wildcards](http://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm),
except that `**` is aka super-asterisk, that do not sensitive for separators.

```
pattern:
    { term }

term:
    `*`         matches any sequence of non-separator characters
    `**`        matches any sequence of characters
    `?`         matches any single non-separator character
    `[` [ `!` ] class `]`
                character class; `!` negates it
    `{` pattern-list `}`
                pattern alternatives
    c           matches character c (c != `*`, `**`, `?`, `\`, `[`, `{`, `}`)
    `\` c       matches character c

class:
    lo `-` hi   matches character c for lo <= c <= hi
    { c }       matches any of the listed characters (c != `\`, `]`;
                `\` c matches c, `-` is literal here); must be non-empty

pattern-list:
    pattern { `,` pattern }
                comma-separated (without spaces) patterns
```

### Separators

The separators are not part of the pattern syntax -- they are configured
once, at compilation time, as the extra arguments of `Compile`:

```go
g := glob.MustCompile("api.*.com", '.', '/')
```

They only limit the wildcards: `*` and `?` never match a separator, while
`**` matches across them; the literals and the character classes are not
affected. With no separators given, `*` and `**` are equivalent. A compiled
`*glob.Pattern` keeps its separators for all matches -- to match the same
pattern with different separators, compile it again.

## Performance

This library is created for compile-once patterns. This means, that
compilation could take time, but strings matching is done faster, than in
case when always parsing template.

If you will not use compiled `*glob.Pattern` object, and do
`g := glob.MustCompile(pattern); g.Match(...)` every time, then your code
will be much more slower.

`Match` performs zero allocations and is safe for concurrent use. Common
pattern shapes (literals, prefixes, suffixes, substrings) are recognized at
compile time and matched with plain string comparisons; the backtracking
engine behind the rest is differentially fuzzed against the `regexp` package
(see `FuzzMatchRegexp`).

Run `go test -bench=.` from source root to see the benchmarks (the numbers
below are from an Apple M4):

Pattern | Fixture | Match | Speed (ns/op)
--------|---------|-------|--------------
`[a-z][!a-x]*cat*[h][!b]*eyes*` | `my cat has very bright eyes` | `true` | 142
`[a-z][!a-x]*cat*[h][!b]*eyes*` | `my dog has very bright eyes` | `false` | 45
`https://*.google.*` | `https://account.google.com` | `true` | 16
`https://*.google.*` | `https://google.com` | `false` | 14
`{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}` | `http://yahoo.com` | `true` | 61
`{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}` | `http://google.com` | `false` | 71
`{https://*gobwas.com,http://exclude.gobwas.com}` | `https://safe.gobwas.com` | `true` | 23
`{https://*gobwas.com,http://exclude.gobwas.com}` | `http://safe.gobwas.com` | `false` | 32
`google.com` | `google.com` | `true` | 4.5
`google.com` | `gobwas.com` | `false` | 3.5
`abc*` | `abcdef` | `true` | 4.0
`abc*` | `af` | `false` | 6.2
`*def` | `abcdef` | `true` | 4.2
`*def` | `af` | `false` | 2.9
`ab*ef` | `abcdef` | `true` | 6.2
`ab*ef` | `af` | `false` | 2.9

The same things with the `regexp` package -- not to pick on it (it is a
general-purpose engine with much stronger guarantees), but as a reference
for how the glob-shaped specialization pays off per pattern:

Pattern | Fixture | Match | Speed (ns/op) | glob is
--------|---------|-------|---------------|--------
`^[a-z][^a-x].*cat.*[h][^b].*eyes.*$` | `my cat has very bright eyes` | `true` | 511 | 3.6x faster
`^[a-z][^a-x].*cat.*[h][^b].*eyes.*$` | `my dog has very bright eyes` | `false` | 225 | 5.0x faster
`^https://.*\.google\..*$` | `https://account.google.com` | `true` | 276 | 17x faster
`^https://.*\.google\..*$` | `https://google.com` | `false` | 151 | 11x faster
`^(https://.*\.google\..*\|.*yandex\..*\|.*yahoo\..*\|.*mail\.ru)$` | `http://yahoo.com` | `true` | 385 | 6.3x faster
`^(https://.*\.google\..*\|.*yandex\..*\|.*yahoo\..*\|.*mail\.ru)$` | `http://google.com` | `false` | 549 | 7.7x faster
`^(https://.*gobwas\.com\|http://exclude\.gobwas\.com)$` | `https://safe.gobwas.com` | `true` | 218 | 9.4x faster
`^(https://.*gobwas\.com\|http://exclude\.gobwas\.com)$` | `http://safe.gobwas.com` | `false` | 45 | 1.4x faster
`^google\.com$` | `google.com` | `true` | 31 | 7.0x faster
`^google\.com$` | `gobwas.com` | `false` | 17 | 4.8x faster
`^abc.*$` | `abcdef` | `true` | 41 | 10x faster
`^abc.*$` | `af` | `false` | 1.3 | 4.7x slower
`^.*def$` | `abcdef` | `true` | 72 | 17x faster
`^.*def$` | `af` | `false` | 1.3 | 2.2x slower
`^ab.*ef$` | `abcdef` | `true` | 77 | 12x faster
`^ab.*ef$` | `af` | `false` | 1.3 | 2.2x slower

(The three `slower` rows are the tiny-mismatch cases. Both engines reject
them with the same literal check; `regexp` just reaches it through less
call overhead. In absolute terms it is 2ns vs 6ns -- negligible either
way.)

[godoc-image]: https://pkg.go.dev/badge/github.com/gobwas/glob.svg
[godoc-url]: https://pkg.go.dev/github.com/gobwas/glob
[ci-image]: https://github.com/gobwas/glob/actions/workflows/ci.yml/badge.svg?branch=master
[ci-url]: https://github.com/gobwas/glob/actions/workflows/ci.yml
