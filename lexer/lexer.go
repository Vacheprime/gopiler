package lexer

import (
	"errors"
	"io"
	"slices"

	utils "github.com/Vacheprime/gopiler"
	pw "github.com/Vacheprime/gopiler/lexer/regex/powerset"
)

type TokenType int

const (
	IDENTIFIER TokenType = iota
	INTEGER
	FLOAT
	ASSIGNMENT
	NEWLINE
	DTYPE_INT
	DTYPE_FLOAT
	WHITESPACE
	SEMICOLON

	ERROR
	EOF
)

var stringToTokenType map[string]TokenType = map[string]TokenType{
	"IDENTIFIER":  IDENTIFIER,
	"INTEGER":     INTEGER,
	"FLOAT":       FLOAT,
	"ASSIGNMENT":  ASSIGNMENT,
	"NEWLINE":     NEWLINE,
	"DTYPE_INT":   DTYPE_INT,
	"DTYPE_FLOAT": DTYPE_FLOAT,
	"WHITESPACE":  WHITESPACE,
	"SEMICOLON":   SEMICOLON,
}

type Token struct {
	Repr   string
	TkType TokenType
	Pos    Position
}

// Position represents the position of a token.
type Position struct {
	Line int
	Col  int // Column could be the start character or end character.
}

// TokenStream provides primitives for reading tokens from source code.
type TokenStream interface {
	// NextToken returns and consumes the next token in the stream or an error from the underlying
	// matcher.
	NextToken() (token Token, err error)

	// PeekToken returns without consuming the nth token in the stream starting
	// from the current position where n = 0 is the next token.
	//
	// EOF is returned for every n beyond the available tokens in the stream.
	PeekToken(n int) (token Token, err error)
}

type Lexer struct {
	matcher     pw.SequentialMatcher
	definitions []Definition
	tokenBuffer utils.Queue[Token]

	// TODO: Group these two fields
	lastNLIdx int
	nlCount   int
}

func NewLexer(sm pw.SequentialMatcher, defs []Definition) *Lexer {
	return &Lexer{
		matcher:     sm,
		definitions: defs,
		lastNLIdx:   0,
		nlCount:     0,
	}
}

func (l *Lexer) NextToken() (token Token, err error) {
	for {
		if !l.tokenBuffer.IsEmpty() {
			token, _ = l.tokenBuffer.Dequeue()
			return token, nil
		}

		token, insType, strtOffset, err := l.getTokenFromMatcher()
		if err != nil {
			return token, err
		}

		if token.TkType == NEWLINE {
			l.nlCount++
			l.lastNLIdx = strtOffset + len(token.Repr)
		}

		if insType == IGNORE {
			continue
		}
		return token, err
	}
}

func (l *Lexer) PeekToken(n int) (token Token, err error) {
	tk, err := l.tokenBuffer.Peek(n)
	if err == nil {
		return tk, nil
	}

	count := 0
	for {
		token, insType, strtOffset, err := l.getTokenFromMatcher()
		if err != nil {
			return token, err
		}

		if token.TkType == NEWLINE {
			l.nlCount++
			l.lastNLIdx = strtOffset + len(token.Repr)
		}

		if insType == IGNORE {
			continue
		}
		l.tokenBuffer.Enqueue(token)

		if count == n {
			return token, nil
		}
		count++
	}
}

func (l *Lexer) getTokenFromMatcher() (tk Token, insType instructionType, strtOffset int, err error) {
	match, err := l.matcher.MatchNext()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return tk, UNDEFINED, match.StartIndex, err // Return unexpected errors (Invalid encoding or other)
		}
		if isMatchMeaningless(match) {
			tk.TkType = EOF
			return tk, UNDEFINED, match.StartIndex, nil
		}
	}

	tk.Repr = match.Match
	tk.Pos = Position{
		Line: l.nlCount,
		Col:  match.StartIndex - l.lastNLIdx,
	}

	insType = UNDEFINED
	if match.IsMatching {
		definition := l.mustGetSourceDefinition(match.Labels)
		tk.TkType = stringToTokenType[definition.Identifier]
		insType = definition.InsType
	} else {
		tk.TkType = ERROR
	}

	return tk, insType, match.StartIndex, nil
}

func (l *Lexer) mustGetSourceDefinition(matchLabels []string) (def Definition) {
	for _, def := range l.definitions {
		if slices.Contains(matchLabels, def.Identifier) {
			return def
		}
	}
	panic("match label does not map to any token definition.")
}

func isMatchMeaningless(m pw.ReMatch) bool {
	return m.Match == ""
}
