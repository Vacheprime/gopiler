package lexer

import (
	"errors"
	"io"
	"testing"

	pw "github.com/Vacheprime/gopiler/lexer/regex/powerset"
)

var (
	defs []Definition = []Definition{
		{
			Identifier: "WHITESPACE",
			DefType:    CLASS,
			InsType:    IGNORE,
		},
		{
			Identifier: "INTEGER",
			DefType:    CLASS,
			InsType:    LEX,
		},
		{
			Identifier: "FLOAT",
			DefType:    CLASS,
			InsType:    LEX,
		},
		{
			Identifier: "ASSIGNMENT",
			DefType:    CLASS,
			InsType:    LEX,
		},
		{
			Identifier: "DTYPE_INT",
			DefType:    CLASS,
			InsType:    LEX,
		},
		{
			Identifier: "DTYPE_FLOAT",
			DefType:    CLASS,
			InsType:    LEX,
		},
		{
			Identifier: "IDENTIFIER",
			DefType:    CLASS,
			InsType:    LEX,
		},
		{
			Identifier: "SEMICOLON",
			DefType:    CLASS,
			InsType:    LEX,
		},
		{
			Identifier: "NEWLINE",
			DefType:    CLASS,
			InsType:    IGNORE,
		},
	}
)

type matchPair struct {
	match pw.ReMatch
	err   error
}

type tokenResult struct {
	tk  Token
	err error
}

type mockMatcher struct {
	matches []matchPair
	currPos int
}

func (m *mockMatcher) MatchNext() (nextMatch pw.ReMatch, err error) {
	if len(m.matches) == 0 {
		return pw.ReMatch{
			StartIndex: m.currPos,
			EndIndex:   m.currPos,
			Match:      "",
			Labels:     []string{},
			IsMatching: false,
		}, io.EOF
	}
	pair := m.matches[0]
	m.currPos = pair.match.EndIndex + 1
	m.matches = m.matches[1:]
	return pair.match, pair.err
}

func (m *mockMatcher) CurrentPosition() (position int) {
	return m.currPos
}

func (m *mockMatcher) Close() error {
	return nil
}

func TestNextToken(t *testing.T) {
	var testCases = []struct {
		name           string
		ex             string // Example string for demonstration. Not actually used in test.
		matcher        pw.SequentialMatcher
		definitions    []Definition
		expectedTokens []Token
	}{
		{
			name: "[Token identification]: Identifier as keyword.",
			ex:   "int abcdefg",
			matcher: &mockMatcher{
				matches: []matchPair{
					{
						match: pw.ReMatch{
							StartIndex: 0,
							EndIndex:   2,
							Match:      "int",
							Labels:     []string{"IDENTIFIER", "DTYPE_INT"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 3,
							EndIndex:   3,
							Match:      " ",
							Labels:     []string{"WHITESPACE"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 4,
							EndIndex:   10,
							Match:      "abcdefg",
							Labels:     []string{"IDENTIFIER"},
							IsMatching: true,
						},
						err: nil,
					},
				},
			},
			definitions: defs,
			expectedTokens: []Token{
				{
					Repr:   "int",
					TkType: DTYPE_INT,
					Pos: Position{
						Line: 0,
						Col:  0,
					},
				},
				{
					Repr:   "abcdefg",
					TkType: IDENTIFIER,
					Pos: Position{
						Line: 0,
						Col:  4,
					},
				},
				{
					TkType: EOF,
				},
			},
		},
		{
			name: "[Instruction type handling] Consecutive ignore tokens followed by EOF.",
			ex:   "   \n   \n   \n   ",
			matcher: &mockMatcher{
				matches: []matchPair{
					{
						match: pw.ReMatch{
							StartIndex: 0,
							EndIndex:   2,
							Match:      "   ",
							Labels:     []string{"WHITESPACE"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 3,
							EndIndex:   3,
							Match:      "\n",
							Labels:     []string{"NEWLINE"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 4,
							EndIndex:   6,
							Match:      "   ",
							Labels:     []string{"WHITESPACE"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 7,
							EndIndex:   7,
							Match:      "\n",
							Labels:     []string{"NEWLINE"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 8,
							EndIndex:   10,
							Match:      "   ",
							Labels:     []string{"WHITESPACE"},
							IsMatching: true,
						},
						err: nil,
					},
				},
			},
			definitions: defs,
			expectedTokens: []Token{
				{
					TkType: EOF,
				},
			},
		},
		{
			name: "[Unmatched sequence] ERROR token for unmatched string.",
			ex:   "abc#",
			matcher: &mockMatcher{
				matches: []matchPair{
					{
						match: pw.ReMatch{
							StartIndex: 0,
							EndIndex:   2,
							Match:      "abc",
							Labels:     []string{"IDENTIFIER"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 3,
							EndIndex:   3,
							Match:      "#",
							Labels:     []string{},
							IsMatching: false,
						},
						err: nil,
					},
				},
			},
			definitions: defs,
			expectedTokens: []Token{
				{
					Repr:   "abc",
					TkType: IDENTIFIER,
					Pos:    Position{Line: 0, Col: 0},
				},
				{
					Repr:   "#",
					TkType: ERROR,
					Pos:    Position{Line: 0, Col: 3},
				},
				{TkType: EOF},
			},
		},
		{
			name: "[Matcher EOF Contract] Valid match along with EOF.",
			ex:   "abc",
			matcher: &mockMatcher{
				matches: []matchPair{
					{
						match: pw.ReMatch{
							StartIndex: 0,
							EndIndex:   2,
							Match:      "abc",
							Labels:     []string{"IDENTIFIER"},
							IsMatching: true,
						},
						err: io.EOF,
					},
				},
			},
			definitions: defs,
			expectedTokens: []Token{
				{
					Repr:   "abc",
					TkType: IDENTIFIER,
					Pos: Position{
						Line: 0,
						Col:  0,
					},
				},
				{
					TkType: EOF,
				},
			},
		},
		{
			name: "[Matcher EOF Contract] Invalid match along with EOF.",
			ex:   "%^&",
			matcher: &mockMatcher{
				matches: []matchPair{
					{
						match: pw.ReMatch{
							StartIndex: 0,
							EndIndex:   2,
							Match:      "%^&",
							Labels:     []string{},
							IsMatching: false,
						},
						err: io.EOF,
					},
				},
			},
			definitions: defs,
			expectedTokens: []Token{
				{
					Repr:   "%^&",
					TkType: ERROR,
					Pos: Position{
						Line: 0,
						Col:  0,
					},
				},
				{
					TkType: EOF,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			lexer := NewLexer(testCase.matcher, testCase.definitions)
			result := []Token{}
			for {
				tk, err := lexer.NextToken()
				if err != nil {
					t.Fatalf("Got error %s while expecting no errors", err)
				}
				result = append(result, tk)
				if tk.TkType == EOF {
					break
				}
			}
			nbrExpected := len(testCase.expectedTokens)
			nbrGot := len(result)
			if nbrGot != nbrExpected {
				t.Fatalf("Expected %d tokens, but got %d", nbrExpected, nbrGot)
			}
			for idx, expTk := range testCase.expectedTokens {
				tkGot := result[idx]
				if tkGot != expTk {
					t.Errorf("Got token %+v while expecting %+v", tkGot, expTk)
				}
			}
		})
	}
}

func TestNextTokenSubsequentEOF(t *testing.T) {
	const SUBSEQUENT_CALLS int = 3
	var testCases = []struct {
		name           string
		matcher        pw.SequentialMatcher
		definitions    []Definition
		expectedTokens []Token
	}{
		{
			name: "[EOF after success] Subsequent EOFs after successful lex.",
			matcher: &mockMatcher{
				matches: []matchPair{
					{
						match: pw.ReMatch{
							StartIndex: 0,
							EndIndex:   1,
							Match:      "xy",
							Labels:     []string{"IDENTIFIER"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 2,
							EndIndex:   2,
							Match:      "=",
							Labels:     []string{"ASSIGNMENT"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 3,
							EndIndex:   3,
							Match:      "",
							Labels:     []string{},
							IsMatching: false,
						},
						err: io.EOF,
					},
				},
			},
			definitions: defs,
			expectedTokens: []Token{
				{
					Repr:   "xy",
					TkType: IDENTIFIER,
					Pos:    Position{Line: 0, Col: 0},
				},
				{
					Repr:   "=",
					TkType: ASSIGNMENT,
					Pos:    Position{Line: 0, Col: 2},
				},
				{TkType: EOF},
				{TkType: EOF},
				{TkType: EOF},
			},
		},
		{
			name: "[EOF after error token] Subsequent EOFs following ERROR token.",
			matcher: &mockMatcher{
				matches: []matchPair{
					{
						match: pw.ReMatch{
							StartIndex: 0,
							EndIndex:   1,
							Match:      "xy",
							Labels:     []string{"IDENTIFIER"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 2,
							EndIndex:   3,
							Match:      "&^",
							Labels:     []string{},
							IsMatching: false,
						},
						err: io.EOF,
					},
				},
			},
			definitions: defs,
			expectedTokens: []Token{
				{
					Repr:   "xy",
					TkType: IDENTIFIER,
					Pos:    Position{Line: 0, Col: 0},
				},
				{
					Repr:   "&^",
					TkType: ERROR,
					Pos:    Position{Line: 0, Col: 2},
				},
				{TkType: EOF},
				{TkType: EOF},
				{TkType: EOF},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			lexer := NewLexer(testCase.matcher, testCase.definitions)
			result := []Token{}
			eofCount := 0
			for eofCount < SUBSEQUENT_CALLS {
				tk, err := lexer.NextToken()
				if err != nil {
					t.Fatalf("Unexpected error %s while expecting none", err)
				}
				result = append(result, tk)
				if tk.TkType == EOF {
					eofCount++
				}
			}
			nbrExpected := len(testCase.expectedTokens)
			nbrGot := len(result)
			if nbrGot != nbrExpected {
				t.Fatalf("Expected %d tokens, but got %d", nbrExpected, nbrGot)
			}
			for idx, expTk := range testCase.expectedTokens {
				tkGot := result[idx]

				if tkGot != expTk {
					t.Errorf("Got token %+v while expecting %+v", tkGot, expTk)
				}
			}
		})
	}
}

func TestFailedNextToken(t *testing.T) {
	ErrIO := errors.New("IO error")
	var testCases = []struct {
		name           string
		matcher        pw.SequentialMatcher
		definitions    []Definition
		expectedTokens []tokenResult
	}{
		{
			name: "[Invalid Encoding] Invalid encoding error.",
			matcher: &mockMatcher{
				matches: []matchPair{
					{
						match: pw.ReMatch{},
						err:   pw.ErrInvalidUTF8Sequence,
					},
				},
			},
			definitions: defs,
			expectedTokens: []tokenResult{
				{
					tk:  Token{},
					err: pw.ErrInvalidUTF8Sequence,
				},
			},
		},
		{
			name: "[Unexpected error] Unexpected error.",
			matcher: &mockMatcher{
				matches: []matchPair{
					{
						match: pw.ReMatch{},
						err:   ErrIO,
					},
				},
			},
			definitions: defs,
			expectedTokens: []tokenResult{
				{
					tk:  Token{},
					err: ErrIO,
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			lexer := NewLexer(testCase.matcher, testCase.definitions)
			result := []tokenResult{}
			for {
				tk, err := lexer.NextToken()
				tkPair := tokenResult{tk: tk, err: err}
				result = append(result, tkPair)
				if tk.TkType == EOF || err != nil {
					break
				}
			}
			nbrExpected := len(testCase.expectedTokens)
			nbrGot := len(result)
			if nbrGot != nbrExpected {
				t.Fatalf("Expected %d tokens, but got %d", nbrExpected, nbrGot)
			}
			for idx, expTk := range testCase.expectedTokens {
				tkGot := result[idx]
				if expTk.err != nil && !errors.Is(expTk.err, tkGot.err) {
					t.Errorf("Got error %s while expecting error %s", tkGot.err, expTk.err)
				}

				if expTk.err == nil && tkGot.tk != expTk.tk {
					t.Errorf("Got token %+v while expecting %+v", tkGot.tk, expTk.tk)
				}
			}
		})
	}
}
