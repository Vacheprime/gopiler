package gopiler

import (
	"strings"
)

// Define constants for token classes
const (
	INITNOCLASS int32 = 0
	INITNOREPR  int32 = 0
	EOF         int32 = 256
	DIGIT       int32 = 257
	OPERATOR    int32 = 258
)

// Define struct for token type
type Token struct {
	Class int32
	Repr  rune
}

func isLayoutChar(chr rune) bool {
	if chr == ' ' || chr == '\t' || chr == '\n' {
		return true
	}
	return false
}

func isDigit(chr rune) bool {
	return chr >= 48 && chr <= 57
}

func isOperator(chr rune) bool {
	return chr == 42 || chr == 43
}

func GetTokens(code string) []Token {
	// Create the buffered reader
	reader := strings.NewReader(code)

	// Initialize the slice of tokens
	tokens := make([]Token, 0, 25)

	// Loop over every char
	for {
		// Create the token
		token := Token{INITNOCLASS, INITNOREPR}

		// Read the next rune
		r, _, err := reader.ReadRune()

		// Check if EoF
		if err != nil {
			token.Class = EOF
			tokens = append(tokens, token)
			break
		}

		// If if whitespace
		if isLayoutChar(r) {
			continue
		}

		// Set representation
		token.Repr = r

		// Identify rune class
		if isDigit(r) {
			token.Class = DIGIT
		} else if isOperator(r) {
			token.Class = OPERATOR
		} else {
			token.Class = r
		}

		tokens = append(tokens, token)
	}
	return tokens
}
