package glob

// EveryOf represents a collection of globs
type EveryOf struct {
	Globs Globs
}

// NewEveryOf returns a new EveryOf from a list of globs
func NewEveryOf(g ...Glob) EveryOf {
	return EveryOf{Globs(g)}
}

// Add adds a glob to the EveryOf collection
func (a *EveryOf) Add(g Glob) {
	a.Globs = append(a.Globs, g)
}

// Match checks every glob until one doesn't matches returning false.
// If every glob matches it returns true.
func (a EveryOf) Match(s string) bool {
	for _, m := range a.Globs {
		if !m.Match(s) {
			return false
		}
	}

	return true
}
