package regex

import (
	"errors"

	gp "github.com/Vacheprime/gopiler"
)

var (
	ErrUnmatchedParenthesis = errors.New("right parenthesis unmatched.")
	ErrUnmatchedBracket     = errors.New("left bracket of character class unmatched.")
	ErrInvalidCharClass     = errors.New("invalid character class")
	ErrMissingOperands      = errors.New("missing operand(s) in expression")
	ErrUnknownToken         = errors.New("unknown token encountered while building AST")
	ErrMalformedRegex       = errors.New("regex is malformed and could not be parsed")
)

type RegexTokenType int
type ExpressionType int

const (
	SINGLE_CHAR RegexTokenType = iota
	CHAR_CLASS
	ANY_CHAR
	QUANTIFIER
	OPERATOR
	LEFT_PARENTHESIS
	RIGHT_PARENTHESIS
)

const (
	BINARY_EXPR ExpressionType = iota
	UNARY_EXPR
	ATOMIC
)

type CharacterClass struct {
	Singulars []rune
	Ranges    []CharacterRange
	IsNegated bool
}

type CharacterRange struct {
	Start rune
	End   rune
}

func NewCharacterClass(tk RegexToken) (CharacterClass, error) {
	acceptedRanges := []CharacterRange{}
	acceptedSingulars := []rune{}
	isNegated := false
	for i := 1; i < len(tk.Repr)-1; i++ {
		curr := tk.Repr[i]
		next := tk.Repr[i+1]
		// Handle exclude classes
		if i == 1 && curr == '^' {
			isNegated = true
			continue
		}
		if next == '-' {
			// Check for end of range
			if i+2 >= len(tk.Repr) {
				return CharacterClass{}, ErrInvalidCharClass
			}
			// Process the end of range
			end := tk.Repr[i+2]
			if end < curr {
				return CharacterClass{}, ErrInvalidCharClass
			}
			// Append range
			acceptedRanges = append(acceptedRanges, CharacterRange{curr, end})
			i += 2 // Skip 2 to land after char range
			continue
		}
		acceptedSingulars = append(acceptedSingulars, curr)
	}
	return CharacterClass{acceptedSingulars, acceptedRanges, isNegated}, nil
}

type RegexToken struct {
	Class RegexTokenType
	Repr  []rune
}

func tokensToString(tks []RegexToken) string {
	f := []rune{}
	for i := range tks {
		f = append(f, tks[i].Repr...)
	}
	return string(f)
}

type Expression struct {
	Type     ExpressionType
	Operator rune
	LExpr    *Expression
	RExpr    *Expression
	Atom     *RegexToken
}

func IsOptionalExpr(e Expression) bool {
	var isOptional bool
	switch e.Type {
	case ATOMIC:
		// Atomic expressions aren't considered optional
	case UNARY_EXPR:
		if e.Operator == '*' || e.Operator == '?' {
			isOptional = true // Don't require any so not optional
		}
	case BINARY_EXPR:
		return IsOptionalExpr(*e.LExpr) && IsOptionalExpr(*e.RExpr)
	}
	return isOptional
}

/*
Transforms a regex into a parse tree.
*/
func RegexToParseTree(regex string) (*Expression, error) {
	// Get tokens as postfix
	tks, err := RegexToPostfix(regex)
	if err != nil {
		return nil, err
	}
	exprStack := gp.NewStack[*Expression]()
	for i := range tks {
		tk := tks[i]
		switch tk.Class {
		case SINGLE_CHAR, ANY_CHAR, CHAR_CLASS:
			e := Expression{ATOMIC, 0, nil, nil, &tk}
			exprStack.Push(&e)
		case OPERATOR:
			if exprStack.Len() < 2 {
				return nil, ErrMissingOperands
			}
			e2, _ := exprStack.Pop()
			e1, _ := exprStack.Pop()
			e3 := Expression{BINARY_EXPR, tk.Repr[0], e1, e2, nil}
			exprStack.Push(&e3)
		case QUANTIFIER:
			if exprStack.Len() == 0 {
				return nil, ErrMissingOperands
			}
			e1, _ := exprStack.Pop()
			e2 := Expression{UNARY_EXPR, tk.Repr[0], e1, nil, nil}
			exprStack.Push(&e2)
		default:
			return nil, ErrUnknownToken
		}
	}
	if exprStack.Len() != 1 {
		return nil, ErrMalformedRegex
	}
	e, _ := exprStack.Pop()
	return e, nil
}

/*
Convert a simple regex into postfix notation using
a variation of the shunting-yard algorithm.

Operator precedence:
Parentheses () - First
Concatenation ab - Second
Alternation | - Third

Quantifiers *,+,? - Directly appended to output since they are unary operators

Since concatenation is implicit, the special character &
will be used to explicitely show this operation.
*/
func RegexToPostfix(regex string) ([]RegexToken, error) {
	tokens, err := tokenizeRegex(regex)
	if err != nil {
		return nil, err
	}
	tokens = prepareRegexString(tokens)
	outputQueue := make([]RegexToken, 0)
	operatorStack := gp.NewStack[RegexToken]()

	for i := range tokens {
		currToken := tokens[i]
		switch {
		case currToken.Class == LEFT_PARENTHESIS:
			operatorStack.Push(currToken)
		case currToken.Class == RIGHT_PARENTHESIS:
			for {
				topOp, err := operatorStack.Pop()
				if err != nil {
					return nil, ErrUnmatchedParenthesis
				}
				if topOp.Class == LEFT_PARENTHESIS {
					break
				}
				outputQueue = append(outputQueue, topOp)
			}
		case currToken.Class == QUANTIFIER:
			outputQueue = append(outputQueue, currToken)
		case currToken.Class == OPERATOR && currToken.Repr[0] == '&':
			for {
				topOp, err := operatorStack.Peak()
				if err != nil {
					operatorStack.Push(currToken)
					break
				}
				if topOp.Class == QUANTIFIER || topOp.Repr[0] == '&' {
					operatorStack.Pop()
					outputQueue = append(outputQueue, topOp)
					continue
				}
				operatorStack.Push(currToken)
				break
			}
		case currToken.Class == OPERATOR && currToken.Repr[0] == '|':
			for {
				topOp, err := operatorStack.Peak()
				if err != nil {
					operatorStack.Push(currToken)
					break
				}
				if topOp.Class == QUANTIFIER || topOp.Class == OPERATOR {
					operatorStack.Pop()
					outputQueue = append(outputQueue, topOp)
					continue
				}
				operatorStack.Push(currToken)
				break
			}
		default:
			outputQueue = append(outputQueue, currToken)
		}
	}
	// Append the rest of the operators
	for {
		topOp, err := operatorStack.Pop()
		if err != nil {
			break
		}
		if topOp.Class == LEFT_PARENTHESIS {
			return nil, ErrUnmatchedParenthesis
		}
		outputQueue = append(outputQueue, topOp)
	}
	return outputQueue, nil
}

/*
Prepares a regex for postfix conversion by adding
'&' to denote concatenation.
*/
func prepareRegexString(tokens []RegexToken) []RegexToken {
	preparedTokens := make([]RegexToken, 0)
	for i := 0; i < len(tokens)-1; i++ {
		curr := tokens[i]
		next := tokens[i+1]

		if (next.Class == SINGLE_CHAR || next.Class == CHAR_CLASS || next.Class == LEFT_PARENTHESIS) &&
			(curr.Class == RIGHT_PARENTHESIS || curr.Class == SINGLE_CHAR || curr.Class == CHAR_CLASS || curr.Class == QUANTIFIER) {
			preparedTokens = append(preparedTokens, curr, RegexToken{OPERATOR, []rune{'&'}})
		} else {
			preparedTokens = append(preparedTokens, curr)
		}
	}
	preparedTokens = append(preparedTokens, tokens[len(tokens)-1])
	return preparedTokens
}

func tokenizeRegex(regex string) ([]RegexToken, error) {
	chars := []rune(regex)
	tokens := make([]RegexToken, 0)
	for i := 0; i < len(chars); i++ {
		c := chars[i]
		switch {
		case c == '(':
			tokens = append(tokens, RegexToken{LEFT_PARENTHESIS, []rune{c}})
		case c == ')':
			tokens = append(tokens, RegexToken{RIGHT_PARENTHESIS, []rune{c}})
		case isQuantifier(c):
			tokens = append(tokens, RegexToken{QUANTIFIER, []rune{c}})
		case isOp(c):
			tokens = append(tokens, RegexToken{OPERATOR, []rune{c}})
		case c == '[':
			tk, charsRead, err := tokenizeCharClass(i, &chars)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, *tk)
			// Skip charsRead
			i += charsRead
		case c == '.':
			tokens = append(tokens, RegexToken{ANY_CHAR, []rune{c}})
		default:
			tokens = append(tokens, RegexToken{SINGLE_CHAR, []rune{c}})
		}
	}
	return tokens, nil
}

func tokenizeCharClass(i int, chars *[]rune) (*RegexToken, int, error) {
	c := (*chars)[i]
	repr := []rune{c}
	skipNext := false
	charsRead := 0
	// Read all runes in char class
	nextChars := (*chars)[i+1:]
	for j := range nextChars {
		charsRead++
		if skipNext {
			skipNext = !skipNext
			continue
		}
		r := nextChars[j]
		// Handle escape
		if r == '\\' {
			// No next char
			if j == len(nextChars)-1 {
				return nil, -1, ErrInvalidCharClass
			}
			next := nextChars[j+1]
			repr = append(repr, next)
			skipNext = true
		}
		repr = append(repr, r)
		if r == ']' {
			break
		}
	}
	if repr[len(repr)-1] != ']' {
		return nil, -1, ErrUnmatchedBracket
	}
	if len(repr) == 2 {
		return nil, -1, ErrInvalidCharClass
	}
	return &RegexToken{CHAR_CLASS, repr}, charsRead, nil
}

func isOp(r rune) bool {
	return r == '|' || r == '&'
}

func isQuantifier(r rune) bool {
	return r == '*' || r == '+' || r == '?'
}
