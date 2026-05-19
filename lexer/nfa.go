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
		return -1
	}
}

func GroupDerivator(acceptedRanges []CharacterRange, acceptedSingulars []rune) DeriveTransitionKey {
	return func(c rune) int32 {
		// Check match in acceptedRanges
		for i := range acceptedRanges {
			charRange := acceptedRanges[i]
			if c >= charRange.start && c <= charRange.end {
				return -1 // Correct transition key
			}
		}
		// Check match in acceptedSingulars
		for i := range acceptedSingulars {
			if acceptedSingulars[i] == c {
				return -1
			}
		}
		return -2 // No transitions
	}
}

type NFANode struct {
	transitions        map[rune]gopiler.Set[*NFANode]
	epsilonTransitions gopiler.Set[*NFANode]
	derivator          DeriveTransitionKey
}

func NewNode(d DeriveTransitionKey) NFANode {
	return NFANode{make(map[rune]gopiler.Set[*NFANode]), gopiler.NewSet[*NFANode](), d}
}

func (node *NFANode) AddTransition(c rune, n *NFANode) {
	if node.derivator == nil {
		return
	}
	key := node.derivator(c)
	s, ok := node.transitions[key]
	if ok {
		s.Add(n)
	} else {
		transitionSet := gopiler.NewSet[*NFANode]()
		transitionSet.Add(n)
		node.transitions[key] = transitionSet
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

func createSimpleNFA(c rune, d DeriveTransitionKey) *NFA {
	s := NewNode(d)
	e := NewNode(nil)
	// Add transition
	s.AddTransition(c, &e)
	// Build NFA
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

func PostfixToNFA(postfixChars []rune) (*NFA, error) {
	nfaFrags := gopiler.NewStack[*NFA]()
	for i := range postfixChars {
		c := postfixChars[i]

		// Simple State
		if isRegexChar(c) {
			nfaFrags.Push(createCharNFA(c))
			continue
		}

		switch c {
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
	PARENTHESIS
)

type RegexToken struct {
	class RegexTokenType
	repr  []rune
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
func RegexToPostfix(regex string) (*[]rune, error) {
	chars := prepareRegexString(regex)
	outputQueue := make([]rune, 0)
	operatorStack := gopiler.NewStack[rune]()

	for i := range chars {
		currentChar := chars[i]
		switch currentChar {
		case '(':
			operatorStack.Push(currentChar)
		case ')':
			for {
				topOp, err := operatorStack.Pop()
				if err != nil {
					return nil, ErrUnmatchedParenthesis
				}
				if topOp == '(' {
					break
				}
				outputQueue = append(outputQueue, topOp)
			}
		case '*', '+', '?':
			outputQueue = append(outputQueue, currentChar)
		case '&':
			for {
				topOp, err := operatorStack.Peak()
				if err != nil {
					operatorStack.Push(currentChar)
					break
				}
				if topOp == '*' || topOp == '+' || topOp == '?' || topOp == '&' {
					operatorStack.Pop()
					outputQueue = append(outputQueue, topOp)
					continue
				}
				operatorStack.Push(currentChar)
				break
			}
		case '|':
			for {
				topOp, err := operatorStack.Peak()
				if err != nil {
					break
				}
				if topOp == '*' || topOp == '+' || topOp == '?' || topOp == '&' || topOp == '|' {
					operatorStack.Pop()
					outputQueue = append(outputQueue, topOp)
					continue
				}
				operatorStack.Push(currentChar)
				break
			}
		default:
			outputQueue = append(outputQueue, currentChar)
		}
	}
	// Append the rest of the operators
	for {
		topOp, err := operatorStack.Pop()
		if err != nil {
			break
		}
		if topOp == '(' {
			return nil, ErrUnmatchedParenthesis
		}
		outputQueue = append(outputQueue, topOp)
	}
	return &outputQueue, nil
}

/*
Prepares a regex for postfix conversion by adding
'&' to denote concatenation.
*/
func prepareRegexString(regex string) []rune {
	chars := []rune(regex)
	preparedChars := make([]rune, 0)
	for i := 0; i < len(chars)-1; i++ {
		curr := chars[i]
		next := chars[i+1]
		if (isRegexChar(next) || next == '(') && (curr == ')' || isRegexChar(curr) || isQuantifier(curr)) {
			preparedChars = append(preparedChars, curr, '&')
		} else {
			preparedChars = append(preparedChars, curr)
		}
	}
	preparedChars = append(preparedChars, chars[len(chars)-1])
	return preparedChars
}

func tokenizeRegex(regex string) ([]RegexToken, error) {
	chars := []rune(regex)
	tokens := make([]RegexToken, 0)
	for i := range chars {
		c := chars[i]
		switch {
		case isParenthesis(c):
			tokens = append(tokens, RegexToken{PARENTHESIS, []rune{c}})
		case isQuantifier(c):
			tokens = append(tokens, RegexToken{QUANTIFIER, []rune{c}})
		case isOp(c):
			tokens = append(tokens, RegexToken{OPERATOR, []rune{c}})
		case c == '[':
			tk, err := tokenizeCharClass(i, &chars)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, *tk)
		default:
			tokens = append(tokens, RegexToken{SINGLE_CHAR, []rune{c}})
		}
	}
	return tokens, nil
}

func tokenizeCharClass(i int, chars *[]rune) (*RegexToken, error) {
	c := (*chars)[i]
	repr := []rune{c}
	skipNext := false
	// Read all runes in char class
	for j := range (*chars)[i:] {
		if skipNext {
			skipNext = !skipNext
			continue
		}
		r := (*chars)[j]
		// Handle escape
		if r == '\\' {
			// No next char
			if j == len(*chars)-1 {
				return nil, ErrInvalidCharClass
			}
			next := (*chars)[j+1]
			repr = append(repr, next)
			skipNext = true
		}
		repr = append(repr, r)
		if r == ']' {
			break
		}
	}
	if repr[len(repr)-1] != ']' {
		return nil, ErrUnmatchedBracket
	}
	return &RegexToken{CHAR_CLASS, repr}, nil
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
