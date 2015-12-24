package match

type Kind int
const(
	KindRaw Kind = iota
	KindMultipleSeparated
	KindMultipleSuper
	KindSingle
	KindComposite
	KindPrefix
	KindSuffix
	KindPrefixSuffix
	KindRangeBetween
	KindRangeList
)


type Matcher interface {
	Match(string) bool
	Search(string) (int, int, bool)
	Kind() Kind
}