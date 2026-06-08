package lexer

import (
	"errors"
	"testing"
)

func TestRegexSubstitution(t *testing.T) {
	definitions := [][]definition{
		{definition{"digit", "[0-9]", FRAG, 0}},
	}
	var tests = []struct {
		name        string
		input       string
		definitions []definition
		want        string
	}{
		{"{digit}* should be [0-9]*", "{digit}*", definitions[0], "[0-9]*"},
		{"{  digit }* should be [0-9]*", "{  digit }*", definitions[0], "[0-9]*"},
		{`\{digit}*|{digit} should be {digit}*|[0-9]`, `\{digit}*|{digit}`, definitions[0], `{digit}*|[0-9]`},
		{"abc should be abc", "abc", definitions[0], "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, err := substituteRegexPlaceholders(tt.input, tt.definitions)
			if err != nil {
				t.Fatalf("got error on substitution %s", err)
			}
			if sub != tt.want {
				t.Errorf("got %s, want %s", sub, tt.want)
			}
		})
	}
}

func TestRegexSubstitutionInvalid(t *testing.T) {
	definitions := [][]definition{
		{definition{"digit", "[0-9]", FRAG, 0}},
	}
	var tests = []struct {
		name        string
		input       string
		definitions []definition
		wantMsg     string
	}{
		{"{digit* should be Unclosed", "{digit*", definitions[0], ErrUnclosedRePlaceholder},
		{`{di git}*|{digit} should be Invalid Name`, `{di git}*|{digit}`, definitions[0], ErrInvalidRePlaceholderName},
		{"{digit}a should be Class not defined", "{digit}a", []definition{}, ErrClassDefNotDefined},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := substituteRegexPlaceholders(tt.input, tt.definitions)
			if err == nil {
				t.Fatalf("error expected, but got none.")
			}
			var parseErr *ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("Expected ParseError, got %T", err)
			}
			if parseErr.Message != tt.wantMsg {
				t.Errorf("got %s, want %s", parseErr.Message, tt.wantMsg)
			}
		})
	}

}
