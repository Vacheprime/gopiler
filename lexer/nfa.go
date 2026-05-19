package lexer

import (
	"errors"

	"github.com/Vacheprime/gopiler"
)

var (
	ErrUnmatchedParenthesis = errors.New("right parenthesis unmatched.")
	ErrUnmatchedBracket     = errors.New("left bracket of character class unmatched.")
	ErrInvalidCharClass     = errors.New("invalid character class")
)

const (
	VALID_KEY   = -1
	INVALID_KEY = -2
)

type DeriveTransitionKey func(c rune) int32

type CharacterRange struct {
	start rune
	end   rune
}

func DirectDerivator() DeriveTransitionKey {
	return func(c rune) int32 {
		return c
	}
}

func AnyDerivator() DeriveTransitionKey {
	return func(c rune) int32 {
		return VALID_KEY
	}
}

func GroupDerivator(acceptedRanges []CharacterRange, acceptedSingulars []rune) DeriveTransitionKey {
	return func(c rune) int32 {
		// Check match in acceptedRanges
		for i := range acceptedRanges {
			charRange := acceptedRanges[i]
			if c >= charRange.start && c <= charRange.end {
				return VALID_KEY
			}
		}
		// Check match in acceptedSingulars
		for i := range acceptedSingulars {
			if acceptedSingulars[i] == c {
				return VALID_KEY
			}
		}
		return INVALID_KEY // No transitions
	}
}

func createGroupDerivator(t RegexToken) (DeriveTransitionKey, error) {
	acceptedRanges := []CharacterRange{}
	acceptedSingulars := []rune{}
	for i := 1; i < len(t.repr)-1; i++ {
		curr := t.repr[i]
		next := t.repr[i+1]
		if next == '-' {
			// Check for end of range
			if i+2 >= len(t.repr) {
				return nil, ErrInvalidCharClass
			}
			// Process the end of range
			end := t.repr[i+2]
			if end < curr {
				return nil, ErrInvalidCharClass
			}
			// Append range
			acceptedRanges = append(acceptedRanges, CharacterRange{curr, end})
			i += 2 // Skip 2 to land after char range
			continue
		}
		acceptedSingulars = append(acceptedSingulars, curr)
	}
	return GroupDerivator(acceptedRanges, acceptedSingulars), nil
}

type NFANode struct {
	transitions        map[rune]gopiler.Set[*NFANode]
	epsilonTransitions gopiler.Set[*NFANode]
	derivator          DeriveTransitionKey
}

func NewNode(d DeriveTransitionKey) NFANode {
	return NFANode{make(map[rune]gopiler.Set[*NFANode]), gopiler.NewSet[*NFANode](), d}
}

func (node *NFANode) AddDirectTransition(c rune, n *NFANode) {
	s, ok := node.transitions[c]
	if ok {
		s.Add(n)
	} else {
		transitionSet := gopiler.NewSet[*NFANode]()
		transitionSet.Add(n)
		node.transitions[c] = transitionSet
	}
}

func (node *NFANode) AddGroupedTransition(n *NFANode) {
	s, ok := node.transitions[VALID_KEY]
	if ok {
		s.Add(n)
	} else {
		transitionSet := gopiler.NewSet[*NFANode]()
		transitionSet.Add(n)
		node.transitions[VALID_KEY] = transitionSet
	}
}

func (node *NFANode) AddEpsilonTransition(n *NFANode) {
	node.epsilonTransitions.Add(n)
}

type NFA struct {
	startNode    *NFANode
	endNode      *NFANode
	acceptStates []*NFANode
}

func createSingleCharNFA(c rune) *NFA {
	s := NewNode(DirectDerivator())
	e := NewNode(nil)
	// Add transition
	s.AddDirectTransition(c, &e)
	// Build NFA
	return &NFA{&s, &e, []*NFANode{&e}}
}

func createGroupedNFA(d DeriveTransitionKey) *NFA {
	s := NewNode(d)
	e := NewNode(nil)
	// Add grouped transition
	s.AddGroupedTransition(&e)
	return &NFA{&s, &e, []*NFANode{&e}}
}

func concatNFAs(n1 *NFA, n2 *NFA) *NFA {
	// Add epsilon transition to n2
	n1.endNode.AddEpsilonTransition(n2.startNode)
	return &NFA{n1.startNode, n2.endNode, []*NFANode{n2.endNode}}
}

func quantify0orMore(n *NFA) *NFA {
	s := NewNode(nil)
	e := NewNode(nil)
	// Add start transitions
	s.AddEpsilonTransition(&e)
	s.AddEpsilonTransition(n.startNode)
	// Add loop transition
	n.endNode.AddEpsilonTransition(n.startNode)
	// Connect to new end node
	n.endNode.AddEpsilonTransition(&e)
	return &NFA{&s, &e, []*NFANode{&e}}
}

func quantify1OrMore(n *NFA) *NFA {
	s := NewNode(nil)
	e := NewNode(nil)
	// Add start transition
	s.AddEpsilonTransition(n.startNode)
	// Add loop transition
	n.endNode.AddEpsilonTransition(n.startNode)
	// Connect to new end node
	n.endNode.AddEpsilonTransition(&e)
	return &NFA{&s, &e, []*NFANode{&e}}
}

func quantify0or1(n *NFA) *NFA {
	n.startNode.AddEpsilonTransition(n.endNode)
	return n
}

func alternateNFAs(n1 *NFA, n2 *NFA) *NFA {
	s := NewNode(nil)
	e := NewNode(nil)
	// Add start epsilon transitions
	s.AddEpsilonTransition(n1.startNode)
	s.AddEpsilonTransition(n2.startNode)
	// Add end epsilon transitions
	n1.endNode.AddEpsilonTransition(&e)
	n2.endNode.AddEpsilonTransition(&e)
	return &NFA{&s, &e, []*NFANode{&e}}
}

func PostfixToNFA(postfixTokens []RegexToken) (*NFA, error) {
	nfaFrags := gopiler.NewStack[*NFA]()
	for i := range postfixTokens {
		t := postfixTokens[i]

		// Simple State
		if t.class == SINGLE_CHAR {
			nfaFrags.Push(createSingleCharNFA(t.repr[0]))
			continue
		}

		// Char class
		if t.class == CHAR_CLASS {
			d, err := createGroupDerivator(t)
			if err != nil {
				return nil, err
			}
			nfaFrags.Push(createGroupedNFA(d))
		}

		switch t.repr[0] {
		// Concat
		case '&':
			n2, _ := nfaFrags.Pop()
			n1, _ := nfaFrags.Pop()
			nfaFrags.Push(concatNFAs(n1, n2))
		// Quantify 0 or more (Kleene)
		case '*':
			n, _ := nfaFrags.Pop()
			nfaFrags.Push(quantify0orMore(n))
		// Quantify 1 or more
		case '+':
			n, _ := nfaFrags.Pop()
			nfaFrags.Push(quantify1OrMore(n))
		// Quantify 0 or 1
		case '?':
			n, _ := nfaFrags.Peak()
			quantify0or1(n)
		// Alternate
		case '|':
			n2, _ := nfaFrags.Pop()
			n1, _ := nfaFrags.Pop()
			nfaFrags.Push(alternateNFAs(n1, n2))
		}
	}
	final, _ := nfaFrags.Pop()
	return final, nil
}

type RegexTokenType int

const (
	SINGLE_CHAR RegexTokenType = iota
	CHAR_CLASS
	QUANTIFIER
	OPERATOR
	LEFT_PARENTHESIS
	RIGHT_PARENTHESIS
)

type RegexToken struct {
	class RegexTokenType
	repr  []rune
}

func tokensToString(tks []RegexToken) string {
	f := []rune{}
	for i := range tks {
		f = append(f, tks[i].repr...)
	}
	return string(f)
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
	operatorStack := gopiler.NewStack[RegexToken]()

	for i := range tokens {
		currToken := tokens[i]
		switch {
		case currToken.class == LEFT_PARENTHESIS:
			operatorStack.Push(currToken)
		case currToken.class == RIGHT_PARENTHESIS:
			for {
				topOp, err := operatorStack.Pop()
				if err != nil {
					return nil, ErrUnmatchedParenthesis
				}
				if topOp.class == LEFT_PARENTHESIS {
					break
				}
				outputQueue = append(outputQueue, topOp)
			}
		case currToken.class == QUANTIFIER:
			outputQueue = append(outputQueue, currToken)
		case currToken.class == OPERATOR && currToken.repr[0] == '&':
			for {
				topOp, err := operatorStack.Peak()
				if err != nil {
					operatorStack.Push(currToken)
					break
				}
				if topOp.class == QUANTIFIER || topOp.repr[0] == '&' {
					operatorStack.Pop()
					outputQueue = append(outputQueue, topOp)
					continue
				}
				operatorStack.Push(currToken)
				break
			}
		case currToken.class == OPERATOR && currToken.repr[0] == '|':
			for {
				topOp, err := operatorStack.Peak()
				if err != nil {
					break
				}
				if topOp.class == QUANTIFIER || topOp.class == OPERATOR {
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
		if topOp.class == LEFT_PARENTHESIS {
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

		if (next.class == SINGLE_CHAR || next.class == CHAR_CLASS || next.class == LEFT_PARENTHESIS) &&
			(curr.class == RIGHT_PARENTHESIS || curr.class == SINGLE_CHAR || curr.class == CHAR_CLASS || curr.class == QUANTIFIER) {
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

func isRegexChar(r rune) bool {
	return !isOp(r) && !isQuantifier(r) && !isParenthesis(r)
}

func isParenthesis(r rune) bool {
	return r == '(' || r == ')'
}
