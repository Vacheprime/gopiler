package glushkov

import (
	"strings"
	"testing"

	re "github.com/Vacheprime/gopiler/lexer/regex"
)

func symbolsToString(symbols []Symbol) string {
	var b strings.Builder
	for i := range symbols {
		s := symbols[i]
		b.WriteString(string(*s.Repr))
	}
	return b.String()
}

func TestDetermineStartSymbols(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		want  string
	}{
		{"abc should be a", "abc", "a"},
		{"a|b should be ab", "a|b", "ab"},
		{"a*b should be ab", "a*b", "ab"},
		{"(a|b)*c should be abc", "(a|b)*c", "abc"},
		{"(a(ab)*)*|(ba*) should be ab", "(a(ab)*)*|(ba*)", "ab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := re.RegexToParseTree(tt.input)
			if err != nil {
				t.Errorf("got error while creating parse tree %s", err)
			}
			symbols := determineStartSymbols(expr)
			res := symbolsToString(symbols)
			if res != tt.want {
				t.Errorf("got %s, want %s", res, tt.want)
			}
		})
	}
}
