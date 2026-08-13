package glushkov

import (
	"slices"
	"testing"

	re "github.com/Vacheprime/gopiler/lexer/regex"
)

func TestGetEquivalenceClasses(t *testing.T) {
	var testCases = []struct {
		name  string
		input []re.CharacterRange
		want  []re.CharacterRange
	}{
		{
			name: "Total overlap",
			input: []re.CharacterRange{
				{Start: 'a', End: 'z'},
				{Start: 'c', End: 'd'},
			},
			want: []re.CharacterRange{
				{Start: 'a', End: 'b'},
				{Start: 'c', End: 'd'},
				{Start: 'e', End: 'z'},
			},
		},
		{
			name: "Overlap at right edge",
			input: []re.CharacterRange{
				{Start: 'a', End: 'e'},
				{Start: 'e', End: 'z'},
			},
			want: []re.CharacterRange{
				{Start: 'a', End: 'd'},
				{Start: 'e', End: 'e'},
				{Start: 'f', End: 'z'},
			},
		},
		{
			name: "Overlap with characters in middle",
			input: []re.CharacterRange{

				{Start: 'a', End: 'e'},
				{Start: 'c', End: 'z'},
			},
			want: []re.CharacterRange{
				{Start: 'a', End: 'b'},
				{Start: 'c', End: 'e'},
				{Start: 'f', End: 'z'},
			},
		},
		{
			name: "Total overlap left",
			input: []re.CharacterRange{
				{Start: 'a', End: 'z'},
				{Start: 'a', End: 'b'},
			},
			want: []re.CharacterRange{
				{Start: 'a', End: 'b'},
				{Start: 'c', End: 'z'},
			},
		},
		{
			name: "Total overlap left bigger than current",
			input: []re.CharacterRange{
				{Start: 'a', End: 'b'},
				{Start: 'a', End: 'z'},
			},
			want: []re.CharacterRange{
				{Start: 'a', End: 'b'},
				{Start: 'c', End: 'z'},
			},
		},
		{
			name: "Total overlap right",
			input: []re.CharacterRange{
				{Start: 'a', End: 'z'},
				{Start: 'c', End: 'z'},
			},
			want: []re.CharacterRange{
				{Start: 'a', End: 'b'},
				{Start: 'c', End: 'z'},
			},
		},
		{
			name: "Multiple overlaps",
			input: []re.CharacterRange{
				{Start: 'a', End: 'f'},
				{Start: 'c', End: 'm'},
				{Start: 'e', End: 'n'},
			},
			want: []re.CharacterRange{
				{Start: 'a', End: 'b'},
				{Start: 'c', End: 'd'},
				{Start: 'e', End: 'f'},
				{Start: 'g', End: 'm'},
				{Start: 'n', End: 'n'},
			},
		},
		{
			name: "No overlaps",
			input: []re.CharacterRange{
				{Start: '0', End: '9'},
				{Start: 'A', End: 'Z'},
				{Start: 'a', End: 'z'},
			},
			want: []re.CharacterRange{
				{Start: '0', End: '9'},
				{Start: 'A', End: 'Z'},
				{Start: 'a', End: 'z'},
			},
		},
		{
			name:  "No ranges",
			input: []re.CharacterRange{},
			want:  []re.CharacterRange{},
		},
		{
			name: "One range",
			input: []re.CharacterRange{
				{Start: 'a', End: 'z'},
			},
			want: []re.CharacterRange{
				{Start: 'a', End: 'z'},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			res := getEquivalenceRanges(testCase.input)
			if !slices.Equal(testCase.want, res) {
				t.Errorf("got %+v, want %+v", res, testCase.want)
			}
		})
	}

}
