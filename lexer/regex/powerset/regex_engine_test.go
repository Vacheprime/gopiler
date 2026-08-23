package powerset

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
)

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
		expectedMatches []ReMatch
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
			expectedMatches: []ReMatch{
				{
					StartIndex: 0,
					EndIndex:   2,
					Match:      "ab0",
					Labels:     []string{"ab0"},
				},
				{
					StartIndex: 3,
					EndIndex:   5,
					Match:      "ab0",
					Labels:     []string{"ab0"},
				},
				{
					StartIndex: 6,
					EndIndex:   8,
					Match:      "ab0",
					Labels:     []string{"ab0"},
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
			expectedMatches: []ReMatch{
				{
					StartIndex: 0,
					EndIndex:   0,
					Match:      "a",
					Labels:     []string{"a"},
				},
				{
					StartIndex: 1,
					EndIndex:   1,
					Match:      "b",
					Labels:     []string{"b"},
				},
				{
					StartIndex: 2,
					EndIndex:   2,
					Match:      "a",
					Labels:     []string{"a"},
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
			expectedMatches: []ReMatch{},
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
			expectedMatches: []ReMatch{
				{
					StartIndex: 0,
					EndIndex:   2,
					Match:      "abc",
					Labels:     []string{"abc"},
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
			expectedMatches: []ReMatch{
				{
					StartIndex: 0,
					EndIndex:   3,
					Match:      "abbb",
					Labels:     []string{},
				},
				{
					StartIndex: 1,
					EndIndex:   3,
					Match:      "bbb",
					Labels:     []string{},
				},
				{
					StartIndex: 2,
					EndIndex:   4,
					Match:      "bbd",
					Labels:     []string{"abbd or bbd"},
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
			expectedMatches: []ReMatch{
				{
					StartIndex: 0,
					EndIndex:   2,
					Match:      "aba",
					Labels:     []string{"partial"},
				},
				{
					StartIndex: 3,
					EndIndex:   4,
					Match:      "cc",
					Labels:     []string{},
				},
				{
					StartIndex: 4,
					EndIndex:   6,
					Match:      "cba",
					Labels:     []string{"partial"},
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
			expectedMatches: []ReMatch{
				{
					StartIndex: 0,
					EndIndex:   0,
					Match:      "a",
					Labels:     []string{},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := io.NopCloser(strings.NewReader(testCase.input))
			matcher := NewDFAMatcher(reader, testCase.dfa)
			matches := []ReMatch{}
			for {
				match, ok := matcher.MatchNext()
				if ok {
					matches = append(matches, match)
				} else {
					err := matcher.Error()
					if errors.Is(err, ErrEOF) {
						break
					}
					matches = append(matches, match)
				}
			}
			nbrMatches := len(matches)
			expectedMatches := len(testCase.expectedMatches)
			if expectedMatches != nbrMatches {
				t.Fatalf("Got %d matches while expecting %d matches", nbrMatches, expectedMatches)
			}
			for idx, expMatch := range testCase.expectedMatches {
				match := matches[idx]
				if !areMatchesEqual(expMatch, match) {
					t.Errorf("Actual match %+v does not equal expected match %+v", match, expMatch)
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
		nbrValidMatches int
	}{
		{
			name:            "Invalid encoding at start.",
			input:           []byte{0xff, 'a', 'a'},
			dfa:             dfa,
			nbrValidMatches: 0,
		},
		{
			name:            "Invalid encoding in middle.",
			input:           []byte{'a', 0xff, 'a'},
			dfa:             dfa,
			nbrValidMatches: 1,
		},
		{
			name:            "Invalid encoding at end.",
			input:           []byte{'a', 'a', 0xff},
			dfa:             dfa,
			nbrValidMatches: 2,
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
			},
			nbrValidMatches: 1,
		},
		{
			name:            "Invalid encoding during unsuccessful matches.",
			input:           []byte{'b', 'b', 0xff},
			dfa:             dfa,
			nbrValidMatches: 0,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := io.NopCloser(strings.NewReader(string(testCase.input)))
			matcher := NewDFAMatcher(reader, testCase.dfa)
			nbrActualMatches := 0
			for {
				_, ok := matcher.MatchNext()
				if ok {
					nbrActualMatches++
				} else {
					err := matcher.Error()
					if errors.Is(err, ErrEOF) {
						t.Fatalf("Expecting encoding errors, got EOF.")
					} else if errors.Is(err, ErrUnexpectedIOError) {
						t.Fatalf("Expecting encoding errors, got IO error.")
					} else if errors.Is(err, ErrInvalidUTF8Sequence) {
						break
					}
				}
			}
			if nbrActualMatches != testCase.nbrValidMatches {
				t.Errorf("Expected %d before encoding error, got %d", testCase.nbrValidMatches, nbrActualMatches)
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
		nbrValidMatches int
	}{
		{
			name:            "IO error at start.",
			input:           "",
			dfa:             dfa,
			nbrValidMatches: 0,
		},
		{
			name:            "IO error after match.",
			input:           "aaa",
			dfa:             dfa,
			nbrValidMatches: 3,
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
			nbrValidMatches: 1,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			reader := io.NopCloser(NewFaultyReader(testCase.input))
			matcher := NewDFAMatcher(reader, testCase.dfa)
			nbrMatches := 0
			for {
				_, ok := matcher.MatchNext()
				if ok {
					nbrMatches++
				} else {
					err := matcher.Error()
					if !errors.Is(err, ErrUnexpectedIOError) {
						t.Fatalf("Expected IO error, got %s", err.Error())
					}
					break
				}
			}
			if nbrMatches != testCase.nbrValidMatches {
				t.Fatalf("Expected %d matches, got %d", testCase.nbrValidMatches, nbrMatches)
			}
		})
	}
}
