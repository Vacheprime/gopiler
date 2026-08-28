package lexer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"unicode"

	pw "github.com/Vacheprime/gopiler/lexer/regex/powerset"
)

const (
	ErrInvalidClassDefinition   string = "Invalid syntax in token class or fragment definition."
	ErrClassRedefined           string = "Class fragment or definition was redefined."
	ErrInvalidRePlaceholder     string = "Invalid syntax for the regex placeholder."
	ErrInvalidRePlaceholderName string = "Invalid identifier for regex placeholder."
	ErrClassDefNotDefined       string = "Class definition not defined."
	ErrUnclosedRePlaceholder    string = "Regex placeholder is never closed."

	ErrTokenDefinitionUnrecognised string = "Definition name is not a valid token."
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

type Lexer struct {
	matcher     pw.SequentialMatcher
	definitions []Definition
	nextMatch   *pw.ReMatch
	lastNLIdx   int
	nlCount     int
}

func NewLexer(sm pw.SequentialMatcher, defs []Definition) *Lexer {
	return &Lexer{
		matcher:     sm,
		definitions: defs,
		nextMatch:   nil,
		lastNLIdx:   0,
		nlCount:     0,
	}
}

func (l *Lexer) NextToken() (token Token, err error) {
	var match pw.ReMatch
	if l.nextMatch != nil {
		match = *l.nextMatch
		l.nextMatch = nil
		err = nil
	} else {
		match, err = l.matcher.MatchNext()
	}

	if errors.Is(err, io.EOF) {
		token.TkType = EOF
		return token, nil
	}

	// Aggregate no matches
	if !match.IsMatching {
		for {
			nextMatch, err := l.matcher.MatchNext()
			if err != nil {
				return token, err
			}
			if nextMatch.IsMatching {
				l.nextMatch = &nextMatch
				break
			}
			if nextMatch.StartIndex <= match.EndIndex {
				continue
			}
			l.nextMatch = &nextMatch
			break
		}
		token.Repr = match.Match
		token.TkType = ERROR
		token.Pos = Position{
			Line: l.nlCount,
			Col:  match.StartIndex,
		}
		return token, nil
	}
	def, err := l.getSourceDefinition(match.Labels)
	if err != nil {
		return token, err
	}
	tkType, ok := stringToTokenType[def.Identifier]
	if !ok {
		return token, errors.New(ErrTokenDefinitionUnrecognised)
	}
	token.TkType = tkType
	token.Repr = match.Match
	pos := Position{
		Line: l.nlCount,
		Col:  match.StartIndex - l.lastNLIdx,
	}
	token.Pos = pos
	if token.TkType == NEWLINE {
		l.nlCount++
		l.lastNLIdx = match.EndIndex
	}

	return token, nil
}

func (l *Lexer) getSourceDefinition(matchLabels []string) (def Definition, err error) {
	for _, def := range l.definitions {
		if slices.Contains(matchLabels, def.Identifier) {
			return def, nil
		}
	}
	return Definition{}, errors.New(ErrTokenDefinitionUnrecognised)
}

/* Parsing related errors. */
type ParseError struct {
	Message string
	Line    int
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf("%s Line: %d", pe.Message, pe.Line)
}

type definitionType int

const (
	CLASS definitionType = iota
	FRAG
)

type instructionType int

const (
	LEX instructionType = iota
	IGNORE
)

/* Encompasses token Definition information. */
type Definition struct {
	Identifier string
	Regex      string
	DefType    definitionType
	InsType    instructionType
}

/*
	Parse token definitions from a file.
	Token definitions returned do not contain the fragment definitions.

TODO: Instead of file path, request a Reader or ReadCloser.
*/
func ParseDefinitions(patternFilePath string) ([]Definition, error) {
	file, err := os.Open(patternFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	classDefs := []Definition{}
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}
		def, err := parseDefinition(line)
		if err != nil {
			err.Line = lineCount
			return nil, err
		}
		found := slices.ContainsFunc(classDefs, func(d Definition) bool {
			if d.Identifier == def.Identifier {
				return true
			}
			return false
		})
		if found {
			return nil, &ParseError{ErrClassRedefined, lineCount}
		}
		substituted, err := substituteRegexPlaceholders(def.Regex, classDefs)
		if err != nil {
			err.Line = lineCount
			return nil, err
		}
		def.Regex = substituted
		classDefs = append(classDefs, def)
		def.Regex = substituted
	}
	return filterDefinitionsByType(classDefs, CLASS), nil
}

/* Filters definitions by the definition type. */
func filterDefinitionsByType(defs []Definition, defType definitionType) []Definition {
	filtered := make([]Definition, 0, len(defs))
	for _, d := range defs {
		if d.DefType == defType {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

/* Parses a definition from a line of input. */
func parseDefinition(line string) (Definition, *ParseError) {
	components := strings.Fields(line)
	if len(components) != 2 {
		return Definition{}, &ParseError{Message: ErrInvalidClassDefinition}
	}
	className := components[0]
	regexDef := components[1]
	defType := CLASS
	insType := LEX
	switch className[0] {
	case '!':
		defType = FRAG
		className = className[1:]
	case '?':
		insType = IGNORE
		className = className[1:]
	}
	return Definition{className, regexDef, defType, insType}, nil
}

/*
Given a regex and a slice of regex definitions, this function replaces the placeholders of a regex with their appropriate definitions.
*/
func substituteRegexPlaceholders(regex string, classDefs []Definition) (string, *ParseError) {
	chars := []rune(regex)
	var substitutedRegexBuilder strings.Builder
	for i := 0; i < len(chars); i++ {
		c := chars[i]
		if c == '\\' {
			if i+1 < len(chars) && chars[i+1] != '{' {
				substitutedRegexBuilder.WriteRune('\\')
				substitutedRegexBuilder.WriteRune(chars[i+1])
			} else {
				substitutedRegexBuilder.WriteRune(chars[i+1])
			}
			i++ // Next char is escaped
		} else if c == '{' {
			if i+1 >= len(chars) {
				return "", &ParseError{Message: ErrInvalidRePlaceholder}
			}
			placeholderId, n, err := parseRegexPlaceholder(chars[i+1:])
			if err != nil {
				return "", err
			}
			foundIdx := slices.IndexFunc(classDefs, func(d Definition) bool {
				return d.Identifier == placeholderId
			})
			if foundIdx == -1 {
				return "", &ParseError{Message: ErrClassDefNotDefined}
			}
			substituteRegex := classDefs[foundIdx].Regex
			substitutedRegexBuilder.WriteString(substituteRegex)
			i += n
		} else {
			substitutedRegexBuilder.WriteRune(c)
		}

	}
	return substitutedRegexBuilder.String(), nil
}

/* Parses a regex placeholder, returning the placeholder name and the number of characters consumed. */
func parseRegexPlaceholder(nextChars []rune) (string, int, *ParseError) {
	var identifierBuilder strings.Builder
	charCount := 0
	wasClosed := false
	for _, c := range nextChars {
		charCount++
		if c == '}' {
			wasClosed = true
			break
		}
		identifierBuilder.WriteRune(c)
	}
	if !wasClosed {
		return "", 0, &ParseError{Message: ErrUnclosedRePlaceholder}
	}
	identifier := strings.TrimSpace(identifierBuilder.String())
	if strings.IndexFunc(identifier, isWhitespace) != -1 {
		return "", 0, &ParseError{Message: ErrInvalidRePlaceholderName}
	}
	return identifier, charCount, nil
}

/* Checks if a rune is a whitespace excluding newlines */
func isWhitespace(r rune) bool {
	return unicode.IsSpace(r) && r != '\n'
}
