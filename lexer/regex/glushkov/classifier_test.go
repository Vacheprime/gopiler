package glushkov

import (
	"slices"
	"testing"

	gp "github.com/Vacheprime/gopiler"
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

func createSymbol(ranges []re.CharacterRange, id int) re.Symbol {
	charSet := re.CharacterSet{Ranges: ranges}
	return re.Symbol{
		SymbolId: id,
		Token:    nil,
		CharSet:  &charSet,
		Label:    "",
	}
}

func createSymbolOccurrences(symbols []re.Symbol) SymbolOccurrences {
	occurrences := map[*re.RegexToken]re.Symbol{}
	for _, sym := range symbols {
		// create tk
		tk := re.RegexToken{}
		occurrences[&tk] = sym
	}
	return occurrences
}

func TestDetermineOwnership(t *testing.T) {
	type alphabetSym struct {
		ranges []re.CharacterRange
		id     int
	}
	var testCases = []struct {
		name         string
		equiRanges   []re.CharacterRange
		alphabet     []alphabetSym
		ownershipSet [][]uint64
	}{
		{
			name: "Unique ownerships",
			equiRanges: []re.CharacterRange{
				{Start: 'a', End: 'z'},
				{Start: '0', End: '9'},
			},
			alphabet: []alphabetSym{
				{
					ranges: []re.CharacterRange{
						{Start: 'a', End: 'z'},
					},
					id: 0,
				},
				{
					ranges: []re.CharacterRange{
						{Start: '0', End: '9'},
					},
					id: 1,
				},
			},
			ownershipSet: [][]uint64{
				{1},
				{2},
			},
		},
		{
			name: "Multiple Ownerships",
			equiRanges: []re.CharacterRange{
				{Start: 'a', End: 'e'},
				{Start: 'f', End: 'z'},
			},
			alphabet: []alphabetSym{
				{
					ranges: []re.CharacterRange{
						{Start: 'a', End: 'z'},
					},
					id: 0,
				},
				{
					ranges: []re.CharacterRange{
						{Start: 'f', End: 'z'},
					},
					id: 1,
				},
			},
			ownershipSet: [][]uint64{
				{1},
				{3},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			symbols := []re.Symbol{}
			for _, a := range testCase.alphabet {
				symbols = append(symbols, createSymbol(a.ranges, a.id))
			}
			occurrences := createSymbolOccurrences(symbols)
			ownershipSets := determineOwnership(testCase.equiRanges, occurrences)
			nbrSets := len(ownershipSets)
			nbrRanges := len(testCase.equiRanges)
			if nbrSets != nbrRanges {
				t.Errorf("Expected %d ranges, got %d", nbrRanges, nbrSets)
			}
			for idx, equiRange := range testCase.equiRanges {
				expectedBitSet := gp.NewBitSetFromInt(testCase.ownershipSet[idx])
				resultBitSet, ok := ownershipSets[equiRange]
				if !ok {
					t.Errorf("Could not locate range %+v in ownership sets", equiRange)
				}
				if !resultBitSet.Equals(expectedBitSet) {
					t.Errorf("Got %+v, want %+v", resultBitSet, expectedBitSet)
				}
			}
		})
	}
}
