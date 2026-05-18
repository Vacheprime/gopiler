package lexer

import (
	"errors"

	"github.com/Vacheprime/gopiler"
)

var (
	ErrUnmatchedParenthesis = errors.New("right parenthesis unmatched.")
)

type CharacterRange struct {
	start rune
	end   rune
}

type KeyDerivator interface {
	DeriveTransitionKey(c rune) int32
}

type DirectDerivator struct{}

func (dd *DirectDerivator) DeriveTransitionKey(c rune) int32 {
	return c
}

type GroupDerivator struct {
	acceptedRanges    []CharacterRange
	acceptedSingulars []rune
}

func (gd *GroupDerivator) DeriveTransitionKey(c rune) int32 {
	// Check match in acceptedRanges
	for i := range gd.acceptedRanges {
		charRange := gd.acceptedRanges[i]
		if c >= charRange.start && c <= charRange.end {
			return -1 // Correct transition key
		}
	}
	// Check match in acceptedSingulars
	for i := range gd.acceptedSingulars {
		if gd.acceptedSingulars[i] == c {
			return -1
		}
	}
	return -2 // No transitions
}

type AnyDerivator struct{}

func (ad *AnyDerivator) DeriveTransitionKey(c rune) int32 {
	return -1 // Accepts any
}

type NFANode struct {
	transitions        map[rune]gopiler.Set[*NFANode]
	epsilonTransitions gopiler.Set[*NFANode]
	derivator          KeyDerivator
}

func NewNode(d KeyDerivator) NFANode {
	return NFANode{make(map[rune]gopiler.Set[*NFANode]), gopiler.NewSet[*NFANode](), d}
}

func (node *NFANode) AddTransition(c rune, n *NFANode) {
	s, ok := node.transitions[c]
	if ok {
		s.Add(n)
	} else {
		transitionSet := gopiler.NewSet[*NFANode]()
		transitionSet.Add(n)
		node.transitions[c] = transitionSet
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

func createCharNFA(c rune) *NFA {
	s := NewNode()
	e := NewNode()
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
	s := NewNode()
	e := NewNode()
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
	s := NewNode()
	e := NewNode()
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
	s := NewNode()
	e := NewNode()
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

func isOp(r rune) bool {
	return isQuantifier(r) || r == '|' || r == '&'
}

func isQuantifier(r rune) bool {
	return r == '*' || r == '+' || r == '?'
}

func isRegexChar(r rune) bool {
	return !isOp(r) && !isParenthesis(r)
}

func isParenthesis(r rune) bool {
	return r == '(' || r == ')'
}
