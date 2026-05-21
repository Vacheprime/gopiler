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
			symInfo := BuildSymbolInformation(expr)
			//symbols := DetermineStartSymbols(expr, occurences)
			res := symbolsToString(symInfo.StartSymbols)
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
		{"a|b should be ba", "a|b", "ab"},
		{"a*b should be b", "a*b", "b"},
		{"(a|b)*c should be c", "(a|b)*c", "c"},
		{"(a(ab)*)*|(ba)* should be ab", "(a(ab)*)*|(ba)*", "baa"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := re.RegexToParseTree(tt.input)
			if err != nil {
				t.Errorf("got error while creating parse tree %s", err)
			}
			symInfo := BuildSymbolInformation(expr)
			//symbols := DetermineStartSymbols(expr, occurences)
			res := symbolsToString(symInfo.FinalSymbols)
			if res != tt.want {
				t.Errorf("got %s, want %s", res, tt.want)
			}
		})
	}
}

func TestDetermineOccurences(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		want  int
	}{
		{"abc should be 3", "abc", 3},
		{"a|b should be 2", "a|b", 2},
		{"a*b should be 2", "a*b", 2},
		{"(a|b)*c should be 3", "(a|b)*c", 3},
		{"(a(ab)*)*|(ba)* should be 5", "(a(ab)*)*|(ba)*", 5},
		{"[a-z] should be 1", "[a-z]", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := re.RegexToParseTree(tt.input)
			if err != nil {
				t.Errorf("got error while creating parse tree %s", err)
			}

			symInfo := BuildSymbolInformation(expr)
			res := len(symInfo.Occurences)
			if res != tt.want {
				t.Errorf("got %d, want %d", res, tt.want)
			}
		})
	}
}
