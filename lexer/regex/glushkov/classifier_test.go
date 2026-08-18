package glushkov

import (
	"slices"
	"testing"

	gp "github.com/Vacheprime/gopiler"
	re "github.com/Vacheprime/gopiler/lexer/regex"
)

type ComparisonType int

const (
	EQUAL ComparisonType = iota
	NOT_EQUAL
)

type Comparison struct {
	LeftSide    rune
	RightSide   rune
	CompareType ComparisonType
}

func (cc Comparison) CompareClassIds(classifier SymbolClassifier) bool {
	leftSideId := classifier.Classify(cc.LeftSide)
	rightSideId := classifier.Classify(cc.RightSide)
	switch cc.CompareType {
	case EQUAL:
		if leftSideId != rightSideId {
			return false
		}
	case NOT_EQUAL:
		if leftSideId == rightSideId {
			return false
		}
	}
	return true
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

// createSymbolOccurrences is a helper for creating symbols based on a slice of slice of
// of character ranges.
//
// The token and label of every symbol is empty and the ID corresponds to the index.
func createSymbolOccurrences(ranges [][]re.CharacterRange) SymbolOccurrences {
	occurrences := map[*re.RegexToken]re.Symbol{}
	for idx, r := range ranges {
		sym := createSymbol(r, idx)
		tk := re.RegexToken{}
		occurrences[&tk] = sym
	}
	return occurrences
}

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

func TestDetermineOwnership(t *testing.T) {
	var testCases = []struct {
		name         string
		equiRanges   []re.CharacterRange
		alphabet     [][]re.CharacterRange
		ownershipSet [][]uint64
	}{
		{
			name: "Unique ownerships",
			equiRanges: []re.CharacterRange{
				{Start: 'a', End: 'z'},
				{Start: '0', End: '9'},
			},
			alphabet: [][]re.CharacterRange{
				{
					{Start: 'a', End: 'z'},
				},
				{
					{Start: '0', End: '9'},
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
			alphabet: [][]re.CharacterRange{
				{
					{Start: 'a', End: 'z'},
				},
				{
					{Start: 'f', End: 'z'},
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
			occurrences := createSymbolOccurrences(testCase.alphabet)
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

func TestDetermineEquivalenceClasses(t *testing.T) {
	var testCases = []struct {
		name               string
		alphabet           [][]re.CharacterRange
		equiRanges         []re.CharacterRange
		expectedNbrClasses int
	}{
		{
			name: "Single equivalence classes",
			alphabet: [][]re.CharacterRange{
				{
					{Start: 'a', End: 'z'},
				},
				{
					{Start: '0', End: '9'},
				},
			},
			equiRanges: []re.CharacterRange{
				{Start: '0', End: '9'},
				{Start: 'a', End: 'z'},
			},
			expectedNbrClasses: 2,
		},
		{
			name: "Same ownership for ranges",
			alphabet: [][]re.CharacterRange{
				{
					{Start: 'a', End: 'z'},
				},
				{
					{Start: 'i', End: 'i'},
				},
			},
			equiRanges: []re.CharacterRange{
				{Start: 'a', End: 'h'},
				{Start: 'i', End: 'i'},
				{Start: 'j', End: 'z'},
			},
			expectedNbrClasses: 2, // Only
		},
		{
			name: "Same ownership with distinct classes",
			alphabet: [][]re.CharacterRange{
				{
					{Start: 'a', End: 'e'},
				},
				{
					{Start: 'c', End: 'z'},
				},
			},
			equiRanges: []re.CharacterRange{
				{Start: 'a', End: 'b'},
				{Start: 'c', End: 'e'},
				{Start: 'f', End: 'z'},
			},
			expectedNbrClasses: 3,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			occurrences := createSymbolOccurrences(testCase.alphabet)
			ownerships := determineOwnership(testCase.equiRanges, occurrences)
			equiClasses, _ := determineEquivalenceClasses(ownerships, occurrences)
			maxClassId := -1
			for _, equiClass := range equiClasses {
				if equiClass.ClassId > maxClassId {
					maxClassId = equiClass.ClassId
				}
			}
			nbrClasses := maxClassId + 1
			if nbrClasses != testCase.expectedNbrClasses {
				t.Errorf("Expected %d classes, got %d", testCase.expectedNbrClasses, nbrClasses)
			}
		})
	}
}

func TestBuildClassifier(t *testing.T) {
	var testCases = []struct {
		name             string
		alphabet         [][]re.CharacterRange
		expectedClassIds []Comparison
		totalEquiClasses int
	}{
		{
			name: "Distinct Symbols.",
			alphabet: [][]re.CharacterRange{
				{
					{Start: 'a', End: 'z'},
				},
				{
					{Start: '0', End: '9'},
				},
			},
			expectedClassIds: []Comparison{
				{
					LeftSide:    'a',
					RightSide:   'z',
					CompareType: EQUAL,
				},
				{
					LeftSide:    '0',
					RightSide:   '9',
					CompareType: EQUAL,
				},
				{
					LeftSide:    'z',
					RightSide:   '0',
					CompareType: NOT_EQUAL,
				},
			},
			totalEquiClasses: 2,
		},
		{
			name: "Overlapping symbols in the middle.",
			alphabet: [][]re.CharacterRange{
				{
					{Start: 'a', End: 'z'},
				},
				{
					{Start: 'i', End: 'i'},
				},
			},
			expectedClassIds: []Comparison{
				{
					LeftSide:    'h',
					RightSide:   'j',
					CompareType: EQUAL,
				},
				{
					LeftSide:    'i',
					RightSide:   'h',
					CompareType: NOT_EQUAL,
				},
			},
			totalEquiClasses: 2,
		},
		{
			name: "Complex overlap",
			alphabet: [][]re.CharacterRange{
				{
					{Start: 'a', End: 'm'},
				},
				{
					{Start: 'c', End: 'l'},
				},
				{
					{Start: 'f', End: 'n'},
				},
			},
			expectedClassIds: []Comparison{
				{
					LeftSide:    'a',
					RightSide:   'c',
					CompareType: NOT_EQUAL,
				},
				{
					LeftSide:    'c',
					RightSide:   'f',
					CompareType: NOT_EQUAL,
				},
				{
					LeftSide:    'f',
					RightSide:   'm',
					CompareType: NOT_EQUAL,
				},
				{
					LeftSide:    'n',
					RightSide:   'm',
					CompareType: NOT_EQUAL,
				},
			},
			totalEquiClasses: 5,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			occurrences := createSymbolOccurrences(testCase.alphabet)
			classifier, err := BuildClassifier(occurrences)
			// Of course, there should be no error.
			if err != nil {
				t.Errorf("Encountered unexpected error %s", err.Error())
			}
			// The total number of equivalence ranges is deterministic and should match
			if classifier.Total() != testCase.totalEquiClasses {
				t.Errorf("Expected %d equivalence ranges, got %d", testCase.totalEquiClasses, classifier.Total())
			}
			compareStrings := map[ComparisonType]string{
				EQUAL:     "equal",
				NOT_EQUAL: "not equal",
			}
			// Testing classification by comparing various class IDs from different characters.
			for _, cmp := range testCase.expectedClassIds {
				if !cmp.CompareClassIds(classifier) {
					t.Errorf("Failed comparison: class id of %q must be %s when compared to class id of %q",
						cmp.LeftSide,
						compareStrings[cmp.CompareType],
						cmp.RightSide,
					)
				}
			}
			// Test that all equivalence ranges associated with a symbol, when simplified (merged),
			// correspond exactly to the initial ranges of the symbol.
			for _, sym := range occurrences {
				classIds := classifier.GetClassIdsFromSymbol(sym)
				allRanges := []re.CharacterRange{}
				for _, id := range classIds {
					allRanges = append(allRanges, classifier.GetEquiRangesFromClassId(id)...)
				}
				mergedRanges := re.MergeCharacterRanges(allRanges)
				if len(mergedRanges) != len(sym.CharSet.Ranges) {
					t.Errorf("Equivalence ranges %+v do not fully represent symbol ranges, %+v", mergedRanges, sym.CharSet.Ranges)
				}
				for _, r := range sym.CharSet.Ranges {
					if !slices.Contains(mergedRanges, r) {
						t.Errorf("Equivalence range %+v is not contained in equivalence ranges %+v", r, mergedRanges)
					}
				}
			}
		})
	}
}
