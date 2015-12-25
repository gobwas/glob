package glob

import (
	"fmt"
	"testing"
)

func TestParseString(t *testing.T) {
	lexer := newLexer("hello")
	fmt.Println(lexer.nextItem())
	fmt.Println(lexer.nextItem())
	fmt.Println(lexer.nextItem())
}
