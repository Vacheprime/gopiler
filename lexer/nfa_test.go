package lexer

import (
	"slices"
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := prepareRegexString(tt.input)
			ex := []rune(tt.want)
			if !slices.Equal(res, ex) {
				t.Errorf("got %s, want %s", string(res), tt.want)
			}
		})
	}
}
