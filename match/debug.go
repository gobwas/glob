package match

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
)

var i = new(int32)

func logf(f string, args ...interface{}) {
	n := int(atomic.LoadInt32(i))
	fmt.Fprint(os.Stderr,
		strings.Repeat("  ", n),
		fmt.Sprintf("(%d) ", n),
		fmt.Sprintf(f, args...),
		"\n",
	)
}

func enter() {
	atomic.AddInt32(i, 1)
}

func leave() {
	atomic.AddInt32(i, -1)
}

func Graphviz(pattern string, m Matcher) string {
	return fmt.Sprintf(`digraph G {graph[label="%s"];%s}`, pattern, graphviz(m, fmt.Sprintf("%x", rand.Int63())))
}

func graphviz(m Matcher, id string) string {
	buf := &bytes.Buffer{}

	switch v := m.(type) {
	case Tree:
		fmt.Fprintf(buf, `"%s"[label="%s"];`, id, v.value)
		for _, m := range []Matcher{v.left, v.right} {
			switch n := m.(type) {
			case nil:
				rnd := rand.Int63()
				fmt.Fprintf(buf, `"%x"[label="<nil>"];`, rnd)
				fmt.Fprintf(buf, `"%s"->"%x";`, id, rnd)

			default:
				sub := fmt.Sprintf("%x", rand.Int63())
				fmt.Fprintf(buf, `"%s"->"%s";`, id, sub)
				fmt.Fprintf(buf, graphviz(n, sub))
			}
		}

	case Container:
		fmt.Fprintf(buf, `"%s"[label="*AnyOf"];`, id)
		for _, m := range v.Content() {
			rnd := rand.Int63()
			fmt.Fprintf(buf, graphviz(m, fmt.Sprintf("%x", rnd)))
			fmt.Fprintf(buf, `"%s"->"%x";`, id, rnd)
		}

	case EveryOf:
		fmt.Fprintf(buf, `"%s"[label="EveryOf"];`, id)
		for _, m := range v.ms {
			rnd := rand.Int63()
			fmt.Fprintf(buf, graphviz(m, fmt.Sprintf("%x", rnd)))
			fmt.Fprintf(buf, `"%s"->"%x";`, id, rnd)
		}

	default:
		fmt.Fprintf(buf, `"%s"[label="%s"];`, id, m)
	}

	return buf.String()
}
