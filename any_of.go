package glob

// AnyOf represents a collection of globs
type AnyOf struct {
	Globs Globs
}

// NewAnyOf returns a new AnyOf from a list of globs
func NewAnyOf(g ...Glob) AnyOf {
	return AnyOf{Globs(g)}
}

// Add adds a glob to the AnyOf collection
func (a *AnyOf) Add(g Glob) {
	a.Globs = append(a.Globs, g)
}

// Match checks every glob until one matches returning true.
// If none matches it returns false.
func (a AnyOf) Match(s string) bool {
	for _, m := range a.Globs {
		if m.Match(s) {
			return true
		}
	}

	return false
}
