package gopiler

import (
	"errors"
)

const (
	EXP_TYPE_DIGIT rune = 'D'
	EXP_TYPE_EXP   rune = 'E'
)

type TokenIterator struct {
	tokens       *[]Token
	currentIndex int
}

func CreateIterator(tokens *[]Token) TokenIterator {
	return TokenIterator{tokens: tokens, currentIndex: 0}
}

func (ti *TokenIterator) Next() *Token {
	if ti.currentIndex == len(*ti.tokens) {
		return nil
	}
	nextToken := (*ti.tokens)[ti.currentIndex]
	ti.currentIndex++
	return &nextToken
}

type Expression struct {
	Type       rune
	DigitValue int
	LeftExpr   *Expression
	RightExpr  *Expression
	Operator   rune
}

func parseExpression(tokenItr *TokenIterator) (*Expression, error) {
	// Instantiate the expression
	expr := Expression{Type: EXP_TYPE_EXP, DigitValue: 0}

	// PARSE LEFT EXPRESSION
	// Get the next token. Should be opening parenthesis or digit
	token := tokenItr.Next()
	if token == nil {
		return nil, errors.New("Token missing. Token should be parenthesis or digit.")
	}

	// Check class
	switch token.Class {
	case DIGIT:
		expr.LeftExpr = &Expression{Type: EXP_TYPE_DIGIT, DigitValue: int(token.Repr - '0')}
	case '(':
		// Parse left expr
		leftExpr, err := parseExpression(tokenItr)
		if err != nil {
			return nil, err
		}
		expr.LeftExpr = leftExpr
	default:
		return nil, errors.New("Left expression must be a digit or an expression.")
	}

	// PARSE MIDDLE EXPR
	token = tokenItr.Next()
	if token == nil {
		return nil, errors.New("Token missing. Operator expected.")
	}

	// Middle expression MUST be an operator
	if token.Class != OPERATOR {
		return nil, errors.New("Token expected to be an operator.")
	}

	// Set the operator
	expr.Operator = token.Repr

	// PARSE RIGHT EXPR
	// Get next token, should be parenthesis or digit
	token = tokenItr.Next()
	if token == nil {
		return nil, errors.New("Token missing. Token should be parenthesis or digit.")
	}

	// Check class
	switch token.Class {
	case DIGIT:
		expr.RightExpr = &Expression{Type: EXP_TYPE_DIGIT, DigitValue: int(token.Repr - '0')}
	case '(':
		// Parse Right expr
		rightExpr, err := parseExpression(tokenItr)
		if err != nil {
			return nil, err
		}
		expr.RightExpr = rightExpr
	default:
		return nil, errors.New("Right expression must be a digit or an expression.")
	}

	// Consume closing parenthesis
	token = tokenItr.Next()
	if token.Class != ')' {
		return nil, errors.New("Expression must end with a closing parenthesis.")
	}

	// Return final expression
	return &expr, nil
}

func ParseTokens(tokens *[]Token) (*Expression, error) {
	// Create the token iterator
	tokenItr := CreateIterator(tokens)

	// Validate first expression start
	firstTk := tokenItr.Next()
	if firstTk == nil {
		return nil, errors.New("No top-level expression. Code must start with '('.")
	}

	if firstTk.Class != '(' {
		return nil, errors.New("No top-level expression. Code must start with '('.")
	}

	// Start parsing
	return parseExpression(&tokenItr)
}
