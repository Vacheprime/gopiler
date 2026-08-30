package lexer

import (
	"errors"
	"io"
	"slices"

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
}

type Token struct {
	Repr   string
	TkType TokenType
	Pos    Position
}

type Position struct {
	Line int
	Col  int
}

// Lexer tokenizes source code.
type Lexer interface {
	NextToken() (token Token, err error)
}

type SMLexer struct {
	matcher     pw.SequentialMatcher
	definitions []Definition
	nextMatch   *pw.ReMatch
	lastNLIdx   int
	nlCount     int
}

func NewLexer(sm pw.SequentialMatcher, defs []Definition) *SMLexer {
	return &SMLexer{
		matcher:     sm,
		definitions: defs,
		nextMatch:   nil,
		lastNLIdx:   0,
		nlCount:     0,
	}
}

func (l *SMLexer) NextToken() (token Token, err error) {
	var match pw.ReMatch
	if l.nextMatch != nil {
		match = *l.nextMatch
		l.nextMatch = nil
		err = nil
	} else {
		match, err = l.matcher.MatchNext()
	}

	if errors.Is(err, io.EOF) && match.Match == "" {
		token.TkType = EOF
		return token, nil
	}

	token.TkType = ERROR
	if !match.IsMatching {
		nextMatch, err := l.skipSameNoMatches(match)
		if errors.Is(err, io.EOF) {
			if match.Match == "" {
				token.TkType = EOF
				return token, nil
			}
		} else if err != nil {
			return token, err
		}
		l.nextMatch = &nextMatch
	} else {
		def := l.mustGetSourceDefinition(match.Labels)
		if def.InsType == IGNORE {
			nextMatch, err := l.skipIgnoreTokens()
			if errors.Is(err, io.EOF) {
				if match.Match == "" {
					token.TkType = EOF
					return token, nil
				}
			} else if err != nil {
				return token, err
			}
			l.nextMatch = &nextMatch
			return l.NextToken()
		}
		token.TkType = stringToTokenType[def.Identifier]
	}
	token.Repr = match.Match
	token.Pos = Position{
		Line: l.nlCount,
		Col:  match.StartIndex - l.lastNLIdx,
	}
	if token.TkType == NEWLINE {
		l.nlCount++
		l.lastNLIdx = match.EndIndex
	}
	return token, nil
}

func (l *SMLexer) mustGetSourceDefinition(matchLabels []string) (def Definition) {
	for _, def := range l.definitions {
		if slices.Contains(matchLabels, def.Identifier) {
			return def
		}
	}
	panic("match label does not map to any token definition.")
}

func (l *SMLexer) skipSameNoMatches(firstMatch pw.ReMatch) (nextMatch pw.ReMatch, err error) {
	for {
		nextMatch, err = l.matcher.MatchNext()
		if errors.Is(err, io.EOF) {
			if nextMatch.Match == "" {
				return nextMatch, err
			}
		} else if err != nil {
			return nextMatch, err
		}
		if nextMatch.IsMatching || nextMatch.StartIndex > firstMatch.EndIndex {
			break
		}
	}
	return nextMatch, nil
}

func (l *SMLexer) skipIgnoreTokens() (nextMatch pw.ReMatch, err error) {
	for {
		nextMatch, err = l.matcher.MatchNext()
		if errors.Is(err, io.EOF) {
			if nextMatch.Match == "" {
				return nextMatch, err
			}
		} else if err != nil {
			return nextMatch, err
		}
		if !nextMatch.IsMatching {
			return nextMatch, nil
		}
		def := l.mustGetSourceDefinition(nextMatch.Labels)
		if def.InsType == IGNORE {
			continue
		}
		return nextMatch, nil
	}
}
