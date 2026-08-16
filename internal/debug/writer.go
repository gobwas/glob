package debug

import (
	"fmt"
	"io"
	"strings"
)

type FuncWriter func([]byte) (int, error)

func (f FuncWriter) Write(p []byte) (int, error) {
	return f(p)
}

func WriteCounter(w io.Writer, n *int64) io.Writer {
	return FuncWriter(func(p []byte) (int, error) {
		m, err := w.Write(p)
		*n += int64(m)
		return m, err
	})
}

func MatcherName(v any) string {
	s := fmt.Sprintf("%T", v)
	return strings.ReplaceAll(
		strings.ReplaceAll(s, "*glob.", ""),
		"Matcher", "",
	)
}
