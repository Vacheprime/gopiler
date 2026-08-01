package regex

import (
	"errors"
	"slices"
	"strconv"
	"unicode"

	gp "github.com/Vacheprime/gopiler"
)

var (
	ErrUnmatchedParenthesis  = errors.New("right parenthesis unmatched.")
	ErrUnmatchedBracket      = errors.New("left bracket of character class unmatched.")
	ErrInvalidCharClass      = errors.New("invalid character class,")
	ErrMissingOperands       = errors.New("missing operand(s) in expression.")
	ErrUnknownToken          = errors.New("unknown token encountered while building parse tree.")
	ErrMalformedRegex        = errors.New("regex is malformed and could not be parsed.")
	ErrInvalidEscapeSequence = errors.New("regex contains an invalid escape sequence.")
	ErrUnknownMetaClass      = errors.New("unknown meta class.")
)

var VALID_ESCAPE_CHARS []rune = []rune{
	'[',
	']',
	'(',
	')',
}

var VALID_METACLASSES []rune = []rune{
	'w', // Word character [a-zA-Z0-9_]
	'W',
	's', // Whitespace
	'S',
	'.', // Any char except newline [^\n\r]
}

type RegexTokenType int

const (
	SINGLE_CHAR RegexTokenType = iota
	CHAR_CLASS
	META_CHAR
	QUANTIFIER
	OPERATOR
	LEFT_PARENTHESIS
	RIGHT_PARENTHESIS
)

type ExpressionType int

const (
	BINARY_EXPR ExpressionType = iota
	UNARY_EXPR
	ATOMIC
)

type CharacterClass struct {
	Singulars        []Singular
	Ranges           []CharacterRange
	NegatedSingulars []Singular
	NegatedRanges    []CharacterRange
	IsNegated        bool
}

// Equals checks if a CharacterClass is equal to this CharacterClass.
func (cc1 *CharacterClass) Equals(cc2 CharacterClass) (isEqual bool) {
	if len(cc1.Singulars) != len(cc2.Singulars) ||
		len(cc1.Ranges) != len(cc2.Ranges) {
		return false
	}
	// Need to use ContainsFunc because singulars and ranges may be located
	// at different indexes in both arrays.
	for _, s := range cc1.Singulars {
		if !slices.ContainsFunc(cc2.Singulars, func(s2 Singular) bool {
			return s == s2
		}) {
			return false
		}
	}

	for _, r := range cc1.Ranges {
		if !slices.ContainsFunc(cc2.Ranges, func(r2 CharacterRange) bool {
			return r == r2
		}) {
			return false
		}
	}
	return true
}

func (cc *CharacterClass) Matches(r rune) (matches bool) {
	isInAnyRegular := len(cc.Singulars) == 0 && len(cc.Ranges) == 0
	isOutsideAllNegated := true
	// Check if in any regular
	for _, sing := range cc.Singulars {
		if sing.Char == r {
			isInAnyRegular = true
			break
		}
	}
	if !isInAnyRegular {
		for _, cr := range cc.Ranges {
			if cr.Start <= r && cr.End >= r {
				isInAnyRegular = true
				break
			}
		}
	}
	// Check if in any negated
	for _, sing := range cc.NegatedSingulars {
		if sing.Char == r {
			isOutsideAllNegated = false
			break
		}
	}
	for _, cr := range cc.NegatedRanges {
		if cr.Start <= r && cr.End >= r {
			isOutsideAllNegated = false
			break
		}
	}
	if !cc.IsNegated {
		return isInAnyRegular || isOutsideAllNegated
	}
	return isInAnyRegular && isOutsideAllNegated
}

type CharacterRange struct {
	Start rune
	End   rune
}

type Singular struct {
	Char rune
}

func NewCharacterClass(tk RegexToken) (CharacterClass, error) {
	// Handle meta characters
	if tk.Class == META_CHAR {
		return getMetaClass(tk.Repr[0])
	}
	charClass := CharacterClass{
		Ranges:           []CharacterRange{},
		Singulars:        []Singular{},
		NegatedSingulars: []Singular{},
		NegatedRanges:    []CharacterRange{},
	}
	embeddedMetaClasses := []CharacterClass{}
	isNegated := false
	for i := 1; i < len(tk.Repr)-1; i++ {
		curr := tk.Repr[i]
		next := tk.Repr[i+1]
		// Handle negated classes
		if i == 1 && curr == '^' {
			isNegated = true
			charClass.IsNegated = true
			continue
		}
		if curr == '\\' {
			c, isMeta, consumed, err := parseEscapeSequence(tk.Repr[i:])
			if err != nil {
				return CharacterClass{}, ErrInvalidEscapeSequence
			}
			if isMeta {
				metaClass, err := getMetaClass(c)
				if err != nil {
					return CharacterClass{}, err
				}
				embeddedMetaClasses = append(embeddedMetaClasses, metaClass)
			} else {
				sing := Singular{Char: c}
				if isNegated {
					charClass.NegatedSingulars = append(charClass.NegatedSingulars, sing)
				} else {
					charClass.Singulars = append(charClass.NegatedSingulars, sing)
				}
			}
			i += consumed - 1
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
			cr := CharacterRange{curr, end}
			if isNegated {
				charClass.NegatedRanges = append(charClass.NegatedRanges, cr)
			} else {
				charClass.Ranges = append(charClass.Ranges, cr)
			}
			i += 2 // Skip 2 to land after char range
			continue
		}
		sing := Singular{Char: curr}
		if isNegated {
			charClass.NegatedSingulars = append(charClass.NegatedSingulars, sing)
		} else {
			charClass.Singulars = append(charClass.Singulars, sing)
		}
	}
	// Process the metaclasses
	for _, metaClass := range embeddedMetaClasses {
		// Regular singulars
		for _, sing := range metaClass.Singulars {
			// Check if already present
			if slices.ContainsFunc(charClass.Singulars, func(s Singular) bool {
				return s.Char == sing.Char
			}) {
				continue
			}
			if isNegated {
				charClass.NegatedSingulars = append(charClass.NegatedSingulars, sing)
			} else {
				charClass.Singulars = append(charClass.Singulars, sing)
			}
		}
		// Regular Ranges
		for _, metaRange := range metaClass.Ranges {
			if slices.ContainsFunc(charClass.Ranges, func(r CharacterRange) bool {
				return r.Start == metaRange.Start && r.End == metaRange.End
			}) {
				continue
			}
			if isNegated {
				charClass.NegatedRanges = append(charClass.NegatedRanges, metaRange)
			} else {
				charClass.Ranges = append(charClass.Ranges, metaRange)
			}
		}
		// Negated singulars
		for _, sing := range metaClass.NegatedSingulars {
			if slices.ContainsFunc(charClass.NegatedSingulars, func(s Singular) bool {
				return s.Char == sing.Char
			}) {
				continue
			}
			if isNegated {
				charClass.Singulars = append(charClass.Singulars, sing)
			} else {
				charClass.NegatedSingulars = append(charClass.NegatedSingulars, sing)
			}
		}
		// Negated ranges
		for _, metaRange := range metaClass.NegatedRanges {
			if slices.ContainsFunc(charClass.NegatedRanges, func(r CharacterRange) bool {
				return r.Start == metaRange.Start && r.End == metaRange.End
			}) {
				continue
			}
			if isNegated {
				charClass.Ranges = append(charClass.Ranges, metaRange)
			} else {
				charClass.NegatedRanges = append(charClass.NegatedRanges, metaRange)
			}
		}
	}
	return charClass, nil
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
			isOptional = true // Don't require any so optional
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
		case SINGLE_CHAR, META_CHAR, CHAR_CLASS:
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

// prepareRegexString adds explicit concatenation tokens to the regex.
func prepareRegexString(tokens []RegexToken) []RegexToken {
	preparedTokens := make([]RegexToken, 0)
	for i := 0; i < len(tokens)-1; i++ {
		curr := tokens[i]
		next := tokens[i+1]

		if (next.Class == SINGLE_CHAR || next.Class == CHAR_CLASS || next.Class == META_CHAR || next.Class == LEFT_PARENTHESIS) &&
			(curr.Class == RIGHT_PARENTHESIS || curr.Class == SINGLE_CHAR || curr.Class == CHAR_CLASS || curr.Class == META_CHAR || curr.Class == QUANTIFIER) {
			preparedTokens = append(preparedTokens, curr, RegexToken{OPERATOR, []rune{'&'}})
		} else {
			preparedTokens = append(preparedTokens, curr)
		}
	}
	preparedTokens = append(preparedTokens, tokens[len(tokens)-1])
	return preparedTokens
}

// tokenizeRegex tokenizes a regex string.
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
			tokens = append(tokens, RegexToken{META_CHAR, []rune{c}})
		case c == '\\':
			c, isMeta, consumed, err := parseEscapeSequence(chars[i:])
			if err != nil {
				return nil, err
			}
			var tk RegexToken
			if isMeta {
				// Process meta class
				tk.Class = META_CHAR
				tk.Repr = []rune{c}
			} else {
				tk.Class = SINGLE_CHAR
				tk.Repr = []rune{c}
			}
			tokens = append(tokens, tk)
			// Skip consumed
			i += consumed - 1
		default:
			tokens = append(tokens, RegexToken{SINGLE_CHAR, []rune{c}})
		}
	}
	return tokens, nil
}

// parseEscapeSequence parses an escape sequence given a slice of runes.
//
// This method assumes that the first character in the slice is the backslash that
// starts the escape sequence.
func parseEscapeSequence(chars []rune) (char rune, isMetaClass bool, consumed int, err error) {
	if len(chars) <= 1 {
		return 0, false, 0, ErrInvalidEscapeSequence
	}
	next := chars[1]
	found := slices.Index(VALID_ESCAPE_CHARS, next)
	if found > -1 {
		// Treat as single char
		return VALID_ESCAPE_CHARS[found], false, 2, nil
	}
	found = slices.Index(VALID_METACLASSES, next)
	if found > -1 {
		// Treat as meta class
		return VALID_METACLASSES[found], true, 2, nil
	}
	// Attempt to parse unicode escape
	v, _, t, err := strconv.UnquoteChar(string(chars), 0)
	if err != nil {
		return 0, false, 0, ErrInvalidEscapeSequence
	}
	return v, false, len(chars) - len(t), nil
}

func getMetaClass(char rune) (CharacterClass, error) {
	isNegated := unicode.IsUpper(char)
	switch char {
	case 'w', 'W':
		ranges := []CharacterRange{
			{Start: 'a', End: 'z'},
			{Start: 'A', End: 'Z'},
			{Start: '0', End: '9'},
		}
		singulars := []Singular{
			{Char: '_'},
		}
		if isNegated {
			return CharacterClass{NegatedRanges: ranges, NegatedSingulars: singulars}, nil
		}
		return CharacterClass{Ranges: ranges, Singulars: singulars}, nil
	case 's', 'S':
		singulars := []Singular{
			{Char: '\n'},
			{Char: '\r'},
			{Char: '\t'},
			{Char: '\f'},
			{Char: '\v'},
			{Char: ' '},
		}
		if isNegated {
			return CharacterClass{NegatedSingulars: singulars}, nil
		}
		return CharacterClass{Singulars: singulars}, nil
	case '.':
		singulars := []Singular{
			{Char: '\n'},
			{Char: '\r'},
		}
		return CharacterClass{NegatedSingulars: singulars}, nil
	}
	return CharacterClass{}, ErrUnknownMetaClass
}

func tokenizeCharClass(i int, chars *[]rune) (*RegexToken, int, error) {
	c := (*chars)[i]
	repr := []rune{c}
	charsRead := 0
	// Read all runes in char class
	nextChars := (*chars)[i+1:]
	for j := range nextChars {
		charsRead++
		r := nextChars[j]
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
