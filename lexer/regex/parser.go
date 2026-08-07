package regex

import (
	"errors"
	"math"
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
	'd',
	'D',
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

// CharacterRange represents a range of accepted characters.
type CharacterRange struct {
	Start rune
	End   rune
}

// CharacterSet represents a regex character set.
// It is comprised of a list of character ranges that constitute the
// set of characters that matches the set.
type CharacterSet struct {
	Ranges []CharacterRange
}

// BuildCharacterSet creates a new character set from a slice of character ranges.
// This method is the preferred method to use when building a set
// since it simplifies and merges ranges to obtain the simplest set
// possible and it also keeps all ranges sorted by start value.
func BuildCharacterSet(ranges []CharacterRange) CharacterSet {
	if len(ranges) == 0 || len(ranges) == 1 {
		return CharacterSet{Ranges: ranges}
	}

	// Sort ranges by start value
	slices.SortFunc(ranges, func(c1 CharacterRange, c2 CharacterRange) int {
		return int(c1.Start) - int(c2.Start)
	})

	// Merge ranges
	simplifiedRanges := []CharacterRange{ranges[0]}
	for _, currentRange := range ranges[1:] {
		latestSimpRange := &simplifiedRanges[len(simplifiedRanges)-1]

		// Handle distinct ranges (no overlap and not adjacent)
		if currentRange.Start-1 > latestSimpRange.End {
			simplifiedRanges = append(simplifiedRanges, currentRange)
			continue
		}

		// Handle total overlap
		if currentRange.End <= latestSimpRange.End {
			continue
		}

		// Handle partial overlap by extending the range
		latestSimpRange.End = currentRange.End
	}
	return CharacterSet{Ranges: simplifiedRanges}
}

// Equals checks if a CharacterClass is equal to this CharacterClass.
func (cc1 *CharacterSet) Equals(cc2 CharacterSet) (isEqual bool) {
	if len(cc1.Ranges) != len(cc2.Ranges) {
		return false
	}
	// Need to loop over because ranges may be located
	// at different indexes in both arrays.
	for _, s := range cc1.Ranges {
		if !slices.Contains(cc2.Ranges, s) {
			return false
		}
	}
	return true
}

// NegateCharacterSet computes the negated set of the set given.
// This method assumes that the character ranges of the set given are sorted.
func NegateCharacterSet(cs CharacterSet) (negatedSet CharacterSet) {
	var intervalStart int32 = 0
	negatedSet = CharacterSet{[]CharacterRange{}}
	for _, cr := range cs.Ranges {
		intervalEnd := cr.Start - 1
		if intervalEnd < intervalStart {
			continue // Skip, no interval possible
		}
		negatedRange := CharacterRange{Start: intervalStart, End: intervalEnd}
		negatedSet.Ranges = append(negatedSet.Ranges, negatedRange)

		// Handle max size
		if cr.End == math.MaxInt32 {
			break // There cannot be a next interval
		}
		intervalStart = cr.End + 1
	}
	return negatedSet
}

// Matches determines whether a rune r matches the CharacterSet.
func (cc *CharacterSet) Matches(r rune) (matches bool) {
	for _, cs := range cc.Ranges {
		if cs.Start <= r && r <= cs.End {
			return true
		}
	}
	return false
}

// NewCharacterSet creates a new character set from a CHAR_CLASS or META_CLASS regex token.
func NewCharacterSet(tk RegexToken) (CharacterSet, error) {
	// Handle meta characters
	if tk.Class == META_CHAR {
		return getMetaClass(tk.Repr[0])
	}
	ranges := []CharacterRange{}
	isNegated := false

	for i := 1; i < len(tk.Repr)-1; i++ {
		curr := tk.Repr[i]
		next := tk.Repr[i+1]
		// Handle negated classes
		if i == 1 && curr == '^' {
			isNegated = true
			continue
		}
		if curr == '\\' {
			c, isMeta, consumed, err := parseEscapeSequence(tk.Repr[i:])
			if err != nil {
				return CharacterSet{}, ErrInvalidEscapeSequence
			}
			if isMeta {
				metaClass, err := getMetaClass(c)
				if err != nil {
					return CharacterSet{}, err
				}
				ranges = append(ranges, metaClass.Ranges...)
			} else {
				sing := CharacterRange{Start: c, End: c}
				ranges = append(ranges, sing)
			}
			i += consumed - 1
			continue
		}
		if next == '-' {
			// Check for end of range
			if i+2 >= len(tk.Repr) {
				return CharacterSet{}, ErrInvalidCharClass
			}
			// Process the end of range
			end := tk.Repr[i+2]
			if end < curr {
				return CharacterSet{}, ErrInvalidCharClass
			}
			// Append range
			cr := CharacterRange{curr, end}
			ranges = append(ranges, cr)
			i += 2 // Skip 2 to land after char range
			continue
		}
		sing := CharacterRange{Start: curr, End: curr}
		ranges = append(ranges, sing)
	}
	charSet := BuildCharacterSet(ranges)
	if isNegated {
		charSet = NegateCharacterSet(charSet)
	}
	return charSet, nil
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

func getMetaClass(char rune) (CharacterSet, error) {
	isNegated := unicode.IsUpper(char)
	switch char {
	case 'w', 'W':
		ranges := []CharacterRange{
			{Start: '0', End: '9'},
			{Start: 'A', End: 'Z'},
			{Start: '_', End: '_'},
			{Start: 'a', End: 'z'},
		}
		regularSet := CharacterSet{Ranges: ranges}
		if isNegated {
			return NegateCharacterSet(regularSet), nil
		}
		return regularSet, nil
	case 's', 'S':
		ranges := []CharacterRange{
			{Start: '\t', End: '\t'},
			{Start: '\n', End: '\n'},
			{Start: '\v', End: '\v'},
			{Start: '\f', End: '\f'},
			{Start: '\r', End: '\r'},
			{Start: ' ', End: ' '},
		}
		regularSet := CharacterSet{Ranges: ranges}
		if isNegated {
			return NegateCharacterSet(regularSet), nil
		}
		return regularSet, nil
	case 'd', 'D':
		ranges := []CharacterRange{
			{Start: '0', End: '9'},
		}
		regularSet := CharacterSet{Ranges: ranges}
		if isNegated {
			return regularSet, nil
		}
		return regularSet, nil
	case '.':
		ranges := []CharacterRange{
			{Start: '\n', End: '\n'},
			{Start: '\r', End: '\r'},
		}
		regularSet := CharacterSet{Ranges: ranges}
		return NegateCharacterSet(regularSet), nil
	}
	return CharacterSet{}, ErrUnknownMetaClass
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
