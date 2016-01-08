package glob

import (
	"strings"
)

func indexByteNonEscaped(source string, needle, escape byte, shift int) int {
	i := strings.IndexByte(source, needle)
	if i == -1 {
		return -1
	}
	if i == 0 {
		return shift
	}

	if source[i-1] != escape {
		return i + shift
	}

	sh := i + 1
	return indexByteNonEscaped(source[sh:], needle, escape, sh)
}
