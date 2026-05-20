package regex

import (
	"testing"
)

func TestExplicitConcatenation(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		want  string
	}{
		{"abc should be a&b&c&", "abc", "a&b&c"},
		{"a*b should be a*b&", "a*b", "a*&b"},
		{"(a*)* should be (a*)*", "(a*)*", "(a*)*"},
		{"(a)(b) should be (a)&(b)", "(a)(b)", "(a)&(b)"},
		{"(a*)*b((a+b)*e) should be (a*)*&b&((a+&b)*&e)", "(a*)*b((a+b)*e)", "(a*)*&b&((a+&b)*&e)"},
		{"(a|b) should be (a|b)", "(a|b)", "(a|b)"},
		{"(a+b)|a* should be (a+&b)|a*", "(a+b)|a*", "(a+&b)|a*"},
		{"(ab|(cd|e))*e?f* should be (a&b|(c&d|e))*&e?&f*", "(ab|(cd|e))*e?f*", "(a&b|(c&d|e))*&e?&f*"},
		{"[a-zA-Z][a] should be [a-zA-Z]&[a]", "[a-zA-Z][a]", "[a-zA-Z]&[a]"},
		{"[abc]*|ag should be [abc]*|a&g", "[abc]*|ag", "[abc]*|a&g"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tks, err := tokenizeRegex(tt.input)
			if err != nil {
				t.Errorf("got error on tokenization %s", err)
			}
			res := prepareRegexString(tks)
			stringified := tokensToString(res)
			if stringified != tt.want {
				t.Errorf("got %s, want %s", stringified, tt.want)
			}
		})
	}
}
