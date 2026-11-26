package glob

import "testing"

func TestUnmatchedTermsCloseIsSyntaxError(t *testing.T) {
	for _, pattern := range []string{"}", "a}", "}a"} {
		if p, err := Compile(pattern); err == nil {
			t.Errorf("Compile(%q) succeeded with %v; want syntax error", pattern, p)
		}
	}
}
