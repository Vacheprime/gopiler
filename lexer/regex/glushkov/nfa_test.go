package glushkov

import (
	"strings"
	"testing"

	gp "github.com/Vacheprime/gopiler"
	re "github.com/Vacheprime/gopiler/lexer/regex"
)

func symbolsToString(symbols gp.Set[Symbol]) string {
	var b strings.Builder
	for _, v := range symbols.Items() {
		b.WriteString(string(v.Token.Repr))
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
			alphabet := make(map[*re.RegexToken]Symbol)
			symbols := determineStartSymbols(expr, alphabet)
			res := symbolsToString(symbols)
			if res != tt.want {
				t.Errorf("got %s, want %s", res, tt.want)
			}
		})
	}
}

func TestDetermineFinalSymbols(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		want  string
	}{
		{"abc should be c", "abc", "c"},
		{"a|b should be ba", "a|b", "ba"},
		{"a*b should be b", "a*b", "b"},
		{"(a|b)*c should be c", "(a|b)*c", "c"},
		{"(a(ab)*)*|(ba)* should be ab", "(a(ab)*)*|(ba)*", "aba"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := re.RegexToParseTree(tt.input)
			if err != nil {
				t.Errorf("got error while creating parse tree %s", err)
			}
			alphabet := make(map[*re.RegexToken]Symbol)
			symbols := determineFinalSymbols(expr, alphabet)
			res := symbolsToString(symbols)
			if res != tt.want {
				t.Errorf("got %s, want %s", res, tt.want)
			}
		})
	}
}
