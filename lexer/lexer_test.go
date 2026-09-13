package lexer

import (
	"errors"
	"io"
	"testing"

	"github.com/Vacheprime/gopiler"
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

func TestPeekToken(t *testing.T) {
	var testCases = []struct {
		name        string
		ex          string // Example string for demonstration. Not actually used in test.
		matcher     pw.SequentialMatcher
		definitions []Definition
		peekIndex   int
		peekedToken Token // Token returned by peek.
		nextToken   Token // Token returned by one call to NextToken following call to PeekToken.
	}{
		{
			name: "[Peek Next] Peeking the next token.",
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
			peekIndex:   0,
			peekedToken: Token{
				Repr:   "abc",
				TkType: IDENTIFIER,
				Pos:    Position{},
			},
			nextToken: Token{
				Repr:   "abc",
				TkType: IDENTIFIER,
				Pos:    Position{},
			},
		},
		{
			name: "[Peek far ahead within bounds] Peek multiple tokens ahead while staying within bounds of available tokens.",
			ex:   "int a = 5",
			matcher: &mockMatcher{
				matches: []matchPair{
					{
						match: pw.ReMatch{
							StartIndex: 0,
							EndIndex:   2,
							Match:      "int",
							Labels:     []string{"DTYPE_INT"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 4,
							EndIndex:   4,
							Match:      "a",
							Labels:     []string{"IDENTIFIER"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 6,
							EndIndex:   6,
							Match:      "=",
							Labels:     []string{"ASSIGNMENT"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 8,
							EndIndex:   8,
							Match:      "5",
							Labels:     []string{"INTEGER"},
							IsMatching: true,
						},
						err: io.EOF,
					},
				},
			},
			definitions: defs,
			peekIndex:   2,
			peekedToken: Token{
				Repr:   "=",
				TkType: ASSIGNMENT,
				Pos:    Position{Line: 0, Col: 6},
			},
			nextToken: Token{
				Repr:   "int",
				TkType: DTYPE_INT,
				Pos:    Position{},
			},
		},
		{
			name: "[Peek far ahead outside of bounds] Peek further then the available tokens in the stream.",
			ex:   "abc =",
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
							StartIndex: 4,
							EndIndex:   4,
							Match:      "=",
							Labels:     []string{"ASSIGNMENT"},
							IsMatching: true,
						},
						err: io.EOF,
					},
				},
			},
			definitions: defs,
			peekIndex:   10,
			peekedToken: Token{TkType: EOF},
			nextToken: Token{
				Repr:   "abc",
				TkType: IDENTIFIER,
				Pos:    Position{},
			},
		},
		{
			name: "[Peek ignored token] Peeking at a token that has the ignore instruction.",
			ex:   "abc   \nint",
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
							EndIndex:   5,
							Match:      "   ",
							Labels:     []string{"WHITESPACE"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 6,
							EndIndex:   6,
							Match:      "\n",
							Labels:     []string{"NEWLINE"},
							IsMatching: true,
						},
						err: nil,
					},
					{
						match: pw.ReMatch{
							StartIndex: 7,
							EndIndex:   9,
							Match:      "int",
							Labels:     []string{"DTYPE_INT"},
							IsMatching: true,
						},
						err: io.EOF,
					},
				},
			},
			definitions: defs,
			peekIndex:   1,
			peekedToken: Token{
				Repr:   "int",
				TkType: DTYPE_INT,
				Pos:    Position{Line: 1, Col: 0},
			},
			nextToken: Token{
				Repr:   "abc",
				TkType: IDENTIFIER,
				Pos:    Position{},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			lexer := NewLexer(testCase.matcher, testCase.definitions)
			resPeeked, err := lexer.PeekToken(testCase.peekIndex)
			if err != nil {
				t.Fatalf("Unexpected error while peeking: %s", err)
			}
			resNext, err := lexer.NextToken()
			if err != nil {
				t.Fatalf("Unexpected error while consuming next: %s", err)
			}
			if resPeeked != testCase.peekedToken {
				t.Errorf("Peeked token %+v does not match expected peek token %+v", resPeeked, testCase.peekedToken)
			}
			if resNext != testCase.nextToken {
				t.Errorf("Next token %+v does not match expected next token %+v", resNext, testCase.nextToken)
			}
		})
	}
}

func TestNegativePeekIndex(t *testing.T) {
	matcher := &mockMatcher{
		matches: []matchPair{
			{
				match: pw.ReMatch{
					StartIndex: 0,
					EndIndex:   1,
					Match:      "xy",
					Labels:     []string{"IDENTIFIER"},
					IsMatching: true,
				},
				err: io.EOF,
			},
		},
	}
	lexer := NewLexer(matcher, defs)
	n := -1
	_, err := lexer.PeekToken(n)
	if !errors.Is(err, gopiler.ErrNegativeIndex) {
		t.Fatalf("Unexpected error %s while expecting %s", err, gopiler.ErrNegativeIndex)
	}
}

func TestFailPeekToken(t *testing.T) {
	var testCases = []struct {
		name        string
		matcher     pw.SequentialMatcher
		definitions []Definition
		peekIndex   int
		expectedErr error
	}{
		{
			name: "[Encoding Error] Encoding error encountered while trying to peek.",
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
						match: pw.ReMatch{},
						err:   pw.ErrInvalidUTF8Sequence,
					},
				},
			},
			definitions: defs,
			peekIndex:   1,
			expectedErr: pw.ErrInvalidUTF8Sequence,
		},
		{
			name: "[Unexpected Error] Unexpected error while matching.",
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
						match: pw.ReMatch{},
						err:   pw.ErrUnexpectedReaderError,
					},
				},
			},
			definitions: defs,
			peekIndex:   1,
			expectedErr: pw.ErrUnexpectedReaderError,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			lexer := NewLexer(testCase.matcher, testCase.definitions)
			_, err := lexer.PeekToken(testCase.peekIndex)
			if err == nil {
				t.Fatalf("Got no error while expecting error: %s", testCase.expectedErr)
			}
			if !errors.Is(err, testCase.expectedErr) {
				t.Errorf("Got error %s while expeting %s", err, testCase.expectedErr)
			}
		})
	}
}
