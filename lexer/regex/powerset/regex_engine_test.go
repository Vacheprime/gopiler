package powerset

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
)

type matchResult struct {
	Match ReMatch
	Err   error
}

type mockDFA struct {
	transitions     []map[rune]int
	acceptingStates []int
	finalLabels     map[int][]string
}

func (dfa mockDFA) Start() int {
	return 0
}

func (dfa mockDFA) TransitionState(state int, r rune) (nextState int) {
	nextState, ok := dfa.transitions[state][r]
	if !ok {
		return -1
	}
	return nextState
}

func (dfa mockDFA) IsAccepting(state int) bool {
	return slices.Contains(dfa.acceptingStates, state)
}

func (dfa mockDFA) AcceptingLabels(state int) []string {
	labels, ok := dfa.finalLabels[state]
	if !ok {
		return []string{}
	}
	return labels
}

func areMatchesEqual(m1 ReMatch, m2 ReMatch) bool {
	if m1.StartIndex != m2.StartIndex || m1.EndIndex != m2.EndIndex || m1.Match != m2.Match {
		return false
	}
	if len(m1.Labels) != len(m2.Labels) {
		return false
	}
	for _, l := range m1.Labels {
		if !slices.Contains(m2.Labels, l) {
			return false
		}
	}
	return true
}

func TestMatchNext(t *testing.T) {
	var testCases = []struct {
		name            string
		input           string
		dfa             DFA
		expectedMatches []matchResult
	}{
		{
			name:  "Multiple matches same label",
			input: "ab0ab0ab0",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{'b': 2},
					{'0': 3},
					{},
				},
				acceptingStates: []int{3},
				finalLabels: map[int][]string{
					3: {"ab0"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   2,
						Match:      "ab0",
						Labels:     []string{"ab0"},
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 3,
						EndIndex:   5,
						Match:      "ab0",
						Labels:     []string{"ab0"},
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 6,
						EndIndex:   8,
						Match:      "ab0",
						Labels:     []string{"ab0"},
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "Multiple matches different labels",
			input: "aba",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1, 'b': 2},
					{},
					{},
				},
				acceptingStates: []int{1, 2},
				finalLabels: map[int][]string{
					1: {"a"},
					2: {"b"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   0,
						Match:      "a",
						Labels:     []string{"a"},
						IsMatching: true,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   1,
						Match:      "b",
						Labels:     []string{"b"},
						IsMatching: true,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 2,
						EndIndex:   2,
						Match:      "a",
						Labels:     []string{"a"},
						IsMatching: true,
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "Empty string no match.",
			input: "",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{},
				},
				acceptingStates: []int{1},
				finalLabels: map[int][]string{
					1: {"a"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   0,
						Match:      "",
						IsMatching: false,
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "Full string match.",
			input: "abc",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{'b': 2},
					{'c': 3},
					{},
				},
				acceptingStates: []int{3},
				finalLabels: map[int][]string{
					3: {"abc"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   2,
						Match:      "abc",
						Labels:     []string{"abc"},
						IsMatching: true,
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "Incomplete match followed by full match.",
			input: "abbbd",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1, 'b': 2},
					{'b': 2},
					{'b': 3},
					{'d': 4},
					{},
				},
				acceptingStates: []int{4},
				finalLabels: map[int][]string{
					4: {"abbd or bbd"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   3,
						Match:      "abbb",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   3,
						Match:      "bbb",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 2,
						EndIndex:   4,
						Match:      "bbd",
						Labels:     []string{"abbd or bbd"},
						IsMatching: true,
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "Incomplete match due to EOF followed by match.",
			input: "abcd",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1, 'b': 6},
					{'b': 2},
					{'c': 3},
					{'d': 4},
					{'e': 5},
					{},
					{'c': 7},
					{'d': 8},
					{},
				},
				acceptingStates: []int{5, 8},
				finalLabels: map[int][]string{
					5: {"abcde"},
					8: {"bcd"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   3,
						Match:      "abcd",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   3,
						Match:      "bcd",
						Labels:     []string{"bcd"},
						IsMatching: true,
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "Overlapping matches separated by no match.",
			input: "abaccba",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1, 'c': 1},
					{'b': 2},
					{'a': 3},
					{'c': 4},
					{'c': 5},
					{'d': 6},
					{},
				},
				acceptingStates: []int{3, 6},
				finalLabels: map[int][]string{
					3: {"partial"},
					6: {"full"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   2,
						Match:      "aba",
						Labels:     []string{"partial"},
						IsMatching: true,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 3,
						EndIndex:   4,
						Match:      "cc",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 4,
						EndIndex:   6,
						Match:      "cba",
						Labels:     []string{"partial"},
						IsMatching: true,
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "Failed match at start.",
			input: "a",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{},
				},
				acceptingStates: []int{},
				finalLabels:     map[int][]string{},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   0,
						Match:      "a",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   1,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "Partial match followed by incomplete matches.",
			input: "abcdef",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{'b': 2},
					{'c': 3},
					{'d': 4},
					{'e': 5},
					{'g': 6},
					{},
				},
				acceptingStates: []int{3, 6},
				finalLabels: map[int][]string{
					3: {"partial"},
					6: {"full"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   2,
						Match:      "abc",
						Labels:     []string{"partial"},
						IsMatching: true,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 3,
						EndIndex:   3,
						Match:      "d",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 4,
						EndIndex:   4,
						Match:      "e",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 5,
						EndIndex:   5,
						Match:      "f",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 6,
						EndIndex:   6,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "Incomplete match due to EOF.",
			input: "ab",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{'b': 2},
					{'c': 3},
					{},
				},
				acceptingStates: []int{3},
				finalLabels: map[int][]string{
					3: {"abc"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   1,
						Match:      "ab",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   1,
						Match:      "b",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 2,
						EndIndex:   2,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "Partial match cut off by EOF.",
			input: "abc",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{'b': 2},
					{'c': 3},
					{'d': 4},
					{},
				},
				acceptingStates: []int{2, 4},
				finalLabels: map[int][]string{
					2: {"partial"},
					4: {"full"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   1,
						Match:      "ab",
						Labels:     []string{"partial"},
						IsMatching: true,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 2,
						EndIndex:   2,
						Match:      "c",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 3,
						EndIndex:   3,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := io.NopCloser(strings.NewReader(testCase.input))
			var matcher SequentialMatcher = NewDFAMatcher(reader, testCase.dfa)
			matches := []matchResult{}
			for {
				match, err := matcher.MatchNext()
				res := matchResult{
					Match: match,
					Err:   err,
				}
				matches = append(matches, res)
				if errors.Is(err, io.EOF) {
					break
				}
			}
			nbrMatches := len(matches)
			expectedMatches := len(testCase.expectedMatches)
			if expectedMatches != nbrMatches {
				t.Fatalf("Got %d matches while expecting %d matches", nbrMatches, expectedMatches)
			}
			for idx, expMatchRes := range testCase.expectedMatches {
				matchRes := matches[idx]
				if !areMatchesEqual(expMatchRes.Match, matchRes.Match) {
					t.Errorf("Actual match %+v does not equal expected match %+v", matchRes.Match, expMatchRes.Match)
				}
				if !errors.Is(matchRes.Err, expMatchRes.Err) {
					t.Errorf("Actual match error %+v does not match expected error %+v", matchRes.Err, expMatchRes.Err)
				}
			}
		})
	}
}

func TestInvalidEncoding(t *testing.T) {
	dfa := mockDFA{
		transitions: []map[rune]int{
			{'a': 1},
			{},
		},
		acceptingStates: []int{1},
		finalLabels: map[int][]string{
			1: {"a"},
		},
	}
	var testCases = []struct {
		name            string
		input           []byte
		dfa             DFA
		expectedMatches []matchResult
	}{
		{
			name:  "Invalid encoding at start.",
			input: []byte{0xff, 'a', 'a'},
			dfa:   dfa,
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   0,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: ErrInvalidUTF8Sequence,
				},
			},
		},
		{
			name:  "Invalid encoding in middle.",
			input: []byte{'a', 0xff, 'a'},
			dfa:   dfa,
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   0,
						Match:      "a",
						Labels:     []string{"a"},
						IsMatching: true,
					},
					Err: ErrInvalidUTF8Sequence,
				},
			},
		},
		{
			name:  "Invalid encoding at end.",
			input: []byte{'a', 'a', 0xff},
			dfa:   dfa,
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   0,
						Match:      "a",
						Labels:     []string{"a"},
						IsMatching: true,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   1,
						Match:      "a",
						Labels:     []string{"a"},
						IsMatching: true,
					},
					Err: ErrInvalidUTF8Sequence,
				},
			},
		},
		{
			name:  "Invalid encoding during successful match.",
			input: []byte{'a', 'b', 'c', 0xff},
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{'b': 2},
					{'c': 3},
					{'e': 4},
					{},
				},
				acceptingStates: []int{3, 4},
				finalLabels: map[int][]string{
					3: {"partial"},
					4: {"full"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   2,
						Match:      "abc",
						Labels:     []string{"partial"},
						IsMatching: true,
					},
					Err: ErrInvalidUTF8Sequence,
				},
			},
		},
		{
			name:  "Invalid encoding during partial match.",
			input: []byte{'a', 'b', 'c', 0xff, 'c'},
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{'b': 2},
					{'c': 3},
					{'d': 4},
					{},
				},
				acceptingStates: []int{2, 4},
				finalLabels: map[int][]string{
					2: {"partial"},
					4: {"full"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   1,
						Match:      "ab",
						Labels:     []string{"partial"},
						IsMatching: true,
					},
					Err: ErrInvalidUTF8Sequence,
				},
			},
		},
		{
			name:  "Invalid encoding during unsuccessful matches.",
			input: []byte{'b', 'b', 0xff},
			dfa:   dfa,
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   0,
						Match:      "b",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   1,
						Match:      "b",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 2,
						EndIndex:   2,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: ErrInvalidUTF8Sequence,
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := io.NopCloser(strings.NewReader(string(testCase.input)))
			var matcher SequentialMatcher = NewDFAMatcher(reader, testCase.dfa)
			matches := []matchResult{}
			for {
				match, err := matcher.MatchNext()
				res := matchResult{
					Match: match,
					Err:   err,
				}
				matches = append(matches, res)
				if errors.Is(err, io.EOF) || errors.Is(err, ErrInvalidUTF8Sequence) {
					break
				}
			}
			nbrMatches := len(matches)
			expectedMatches := len(testCase.expectedMatches)
			if expectedMatches != nbrMatches {
				t.Fatalf("Got %d matches while expecting %d matches", nbrMatches, expectedMatches)
			}
			for idx, expMatchRes := range testCase.expectedMatches {
				matchRes := matches[idx]
				if !areMatchesEqual(expMatchRes.Match, matchRes.Match) {
					t.Errorf("Actual match %+v does not equal expected match %+v", matchRes.Match, expMatchRes.Match)
				}
				if !errors.Is(matchRes.Err, expMatchRes.Err) {
					t.Errorf("Actual match error %+v does not match expected error %+v", matchRes.Err, expMatchRes.Err)
				}
			}
		})
	}
}

type faultyReader struct {
	data []byte
}

func NewFaultyReader(input string) *faultyReader {
	return &faultyReader{
		data: []byte(input),
	}
}

func (fr *faultyReader) Read(p []byte) (n int, err error) {
	requested := len(p)
	read := copy(p, fr.data)
	fr.data = fr.data[:read]
	if read < requested {
		return read, errors.New("some IO error")
	}
	return read, nil
}

func TestUnexpectedError(t *testing.T) {
	dfa := mockDFA{
		transitions: []map[rune]int{
			{'a': 1},
			{},
		},
		acceptingStates: []int{1},
		finalLabels: map[int][]string{
			1: {"a"},
		},
	}
	var testCases = []struct {
		name            string
		input           string
		dfa             DFA
		expectedMatches []matchResult
	}{
		{
			name:  "IO error at start.",
			input: "",
			dfa:   dfa,
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   0,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: ErrUnexpectedReaderError,
				},
			},
		},
		{
			name:  "IO error after match.",
			input: "aaa",
			dfa:   dfa,
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   0,
						Match:      "a",
						Labels:     []string{"a"},
						IsMatching: true,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   1,
						Match:      "a",
						Labels:     []string{"a"},
						IsMatching: true,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 2,
						EndIndex:   2,
						Match:      "a",
						Labels:     []string{"a"},
						IsMatching: true,
					},
					Err: ErrUnexpectedReaderError,
				},
			},
		},
		{
			name:  "IO error after partial match.",
			input: "abc",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{'b': 2},
					{'c': 3},
					{'d': 4},
					{},
				},
				acceptingStates: []int{2, 4},
				finalLabels: map[int][]string{
					2: {"partial"},
					4: {"full"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   1,
						Match:      "ab",
						Labels:     []string{"partial"},
						IsMatching: true,
					},
					Err: ErrUnexpectedReaderError,
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := io.NopCloser(NewFaultyReader(testCase.input))
			var matcher SequentialMatcher = NewDFAMatcher(reader, testCase.dfa)
			matches := []matchResult{}
			for {
				match, err := matcher.MatchNext()
				res := matchResult{
					Match: match,
					Err:   err,
				}
				matches = append(matches, res)
				if errors.Is(err, ErrUnexpectedReaderError) {
					break
				} else if err != nil {
					t.Fatalf("Got err %v, while only expecting ErrUnexpectedReaderError", err)
				}
			}
			nbrMatches := len(matches)
			expectedMatches := len(testCase.expectedMatches)
			if expectedMatches != nbrMatches {
				t.Fatalf("Got %d matches while expecting %d matches", nbrMatches, expectedMatches)
			}
			for idx, expMatchRes := range testCase.expectedMatches {
				matchRes := matches[idx]
				if !areMatchesEqual(expMatchRes.Match, matchRes.Match) {
					t.Errorf("Actual match %+v does not equal expected match %+v", matchRes.Match, expMatchRes.Match)
				}
				if !errors.Is(matchRes.Err, expMatchRes.Err) {
					t.Errorf("Actual match error %+v does not match expected error %+v", matchRes.Err, expMatchRes.Err)
				}
			}
		})
	}
}

func TestSubsequentEOF(t *testing.T) {
	const SUBSEQUENT_CALLS int = 3
	var testCases = []struct {
		name            string
		input           string
		dfa             DFA
		expectedMatches []matchResult
	}{
		{
			name:  "EOFs after successfull match consuming all preceding characters.",
			input: "a",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{},
				},
				acceptingStates: []int{1},
				finalLabels: map[int][]string{
					1: {"a"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   0,
						Match:      "a",
						Labels:     []string{"a"},
						IsMatching: true,
					},
					Err: io.EOF,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   1,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   1,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   1,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "EOFs after partial match consuming part of preceding characters.",
			input: "abcd",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{'b': 2},
					{'c': 3},
					{'d': 4},
					{'e': 5},
					{},
				},
				acceptingStates: []int{2, 5},
				finalLabels: map[int][]string{
					2: {"partial"},
					5: {"full"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   1,
						Match:      "ab",
						Labels:     []string{"partial"},
						IsMatching: true,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 2,
						EndIndex:   2,
						Match:      "c",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 3,
						EndIndex:   3,
						Match:      "d",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 4,
						EndIndex:   4,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
				{
					Match: ReMatch{
						StartIndex: 4,
						EndIndex:   4,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
				{
					Match: ReMatch{
						StartIndex: 4,
						EndIndex:   4,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
				{
					Match: ReMatch{
						StartIndex: 4,
						EndIndex:   4,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
			},
		},
		{
			name:  "EOFs after incomplete match consuming none of the preceding characters.",
			input: "abc",
			dfa: mockDFA{
				transitions: []map[rune]int{
					{'a': 1},
					{'b': 2},
					{'c': 3},
					{'d': 4},
					{},
				},
				acceptingStates: []int{4},
				finalLabels: map[int][]string{
					4: {"abcd"},
				},
			},
			expectedMatches: []matchResult{
				{
					Match: ReMatch{
						StartIndex: 0,
						EndIndex:   2,
						Match:      "abc",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 1,
						EndIndex:   1,
						Match:      "b",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 2,
						EndIndex:   2,
						Match:      "c",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: nil,
				},
				{
					Match: ReMatch{
						StartIndex: 3,
						EndIndex:   3,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
				{
					Match: ReMatch{
						StartIndex: 3,
						EndIndex:   3,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
				{
					Match: ReMatch{
						StartIndex: 3,
						EndIndex:   3,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
				{
					Match: ReMatch{
						StartIndex: 3,
						EndIndex:   3,
						Match:      "",
						Labels:     []string{},
						IsMatching: false,
					},
					Err: io.EOF,
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := io.NopCloser(strings.NewReader(testCase.input))
			var matcher SequentialMatcher = NewDFAMatcher(reader, testCase.dfa)
			matches := []matchResult{}
			EOFCount := 0
			for {
				match, err := matcher.MatchNext()
				res := matchResult{
					Match: match,
					Err:   err,
				}
				matches = append(matches, res)
				if err == nil && EOFCount > 0 {
					t.Fatal("Got non-EOF match following an EOF match.")
				}
				if errors.Is(err, io.EOF) {
					EOFCount++
				} else if err != nil {
					t.Fatalf("Got err %v, while only expecting io.EOF", err)
				}
				if EOFCount == SUBSEQUENT_CALLS+1 {
					break
				}
			}
			nbrMatches := len(matches)
			expectedMatches := len(testCase.expectedMatches)
			if expectedMatches != nbrMatches {
				t.Fatalf("Got %d matches while expecting %d matches", nbrMatches, expectedMatches)
			}
			for idx, expMatchRes := range testCase.expectedMatches {
				matchRes := matches[idx]
				if !areMatchesEqual(expMatchRes.Match, matchRes.Match) {
					t.Errorf("Actual match %+v does not equal expected match %+v", matchRes.Match, expMatchRes.Match)
				}
				if !errors.Is(matchRes.Err, expMatchRes.Err) {
					t.Errorf("Actual match error %+v does not match expected error %+v", matchRes.Err, expMatchRes.Err)
				}
			}
		})
	}
}
