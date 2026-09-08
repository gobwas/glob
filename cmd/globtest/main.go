package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/gobwas/glob"
	"github.com/gobwas/glob/internal/debug"
)

func benchString(r testing.BenchmarkResult) string {
	nsop := r.NsPerOp()
	ns := fmt.Sprintf("%10d ns/op", nsop)
	allocs := "0"
	if r.N > 0 {
		if nsop < 100 {
			// The format specifiers here make sure that
			// the ones digits line up for all three possible formats.
			if nsop < 10 {
				ns = fmt.Sprintf("%13.2f ns/op", float64(r.T.Nanoseconds())/float64(r.N))
			} else {
				ns = fmt.Sprintf("%12.1f ns/op", float64(r.T.Nanoseconds())/float64(r.N))
			}
		}

		allocs = fmt.Sprintf("%d", r.MemAllocs/uint64(r.N))
	}

	return fmt.Sprintf("%8d\t%s\t%s allocs", r.N, ns, allocs)
}

// pointAt renders the pattern with a pointer at the offset of the syntax
// error, followed by its reason:
//
//	{a,b
//	----^ unclosed `{`
func pointAt(pattern string, err *glob.SyntaxError) string {
	// The offset is in bytes, while the pointer is drawn in columns.
	col := utf8.RuneCountInString(pattern[:min(err.Offset, len(pattern))])
	return fmt.Sprintf("%s\n%s^ %s\n\n", pattern, strings.Repeat("-", col), err.Reason)
}

func main() {
	var (
		pattern      = flag.String("p", "", "pattern to draw")
		sep          = flag.String("s", "", "comma separated list of separators")
		fixture      = flag.String("f", "", "fixture")
		benchCompile = flag.Bool("bench-compile", false, "benchmark compilation time")
		benchMatch   = flag.Bool("bench-match", false, "benchmark matching time")
		verbose      = flag.Bool("v", false, "print the pattern, fixture, compiled matcher tree and result; without it only the exit status tells: 0 on match, 1 otherwise")
	)

	// Expose testing package flags.
	testing.Init()

	flag.Parse()

	if *pattern == "" {
		flag.Usage()
		os.Exit(2)
	}

	var separators []rune
	for _, c := range strings.Split(*sep, ",") {
		if len(c) == 0 {
			continue
		}
		r, w := utf8.DecodeRuneInString(c)
		if len(c) > w {
			fmt.Fprintln(os.Stderr, "only single charactered separators are allowed")
			os.Exit(2)
		}
		separators = append(separators, r)
	}

	g, err := glob.Compile(*pattern, separators...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not compile pattern:", err)
		var syntaxErr *glob.SyntaxError
		if errors.As(err, &syntaxErr) {
			fmt.Fprint(os.Stderr, "\n"+pointAt(*pattern, syntaxErr))
		}
		os.Exit(2)
	}
	matched := g.Match(*fixture)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	if *verbose {
		fmt.Fprintf(w, "pattern:\t%s\n", g)
		fmt.Fprintf(w, "fixture:\t%s\n", *fixture)
		fmt.Fprintf(w, "matchers:\t%s\n", debug.Tree(g))
		fmt.Fprintf(w, "result:\t%t\n", matched)
		// Flush before the benchmarks: they take a while, and the result
		// must not wait for them.
		w.Flush()
	}

	if *benchCompile {
		b := testing.Benchmark(func(b *testing.B) {
			for b.Loop() {
				glob.Compile(*pattern, separators...)
			}
		})
		fmt.Fprintf(w, "compile:\t%s\n", benchString(b))
	}
	if *benchMatch {
		b := testing.Benchmark(func(b *testing.B) {
			for b.Loop() {
				g.Match(*fixture)
			}
		})
		fmt.Fprintf(w, "match:\t%s\n", benchString(b))
	}
	w.Flush()

	if !matched {
		os.Exit(1)
	}
}
