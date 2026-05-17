package lexer

import (
	"errors"

	"github.com/Vacheprime/gopiler"
)

var (
	ErrUnmatchedParenthesis = errors.New("right parenthesis unmatched.")
)

type NFANode struct {
	transitions        map[rune]gopiler.Set[*NFANode]
	epsilonTransitions gopiler.Set[*NFANode]
}

func NewNode() NFANode {
	return NFANode{make(map[rune]gopiler.Set[*NFANode]), gopiler.NewSet[*NFANode]()}
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
		}

		// Concat Op
		if c == '&' {
			n2, _ := nfaFrags.Pop()
			n1, _ := nfaFrags.Pop()
			nfaFrags.Push(concatNFAs(n1, n2))
		}

		// Kleene Star Op
		if c == '*' {
			n, _ := nfaFrags.Pop()
			nfaFrags.Push(quantify0orMore(n))
		}

		// One or more quantifier
		if c == '+' {
			n, _ := nfaFrags.Pop()
			nfaFrags.Push(quantify1OrMore(n))
		}

		// Zero or one quantifier
		if c == '?' {
			n, _ := nfaFrags.Pop()
			nfaFrags.Push(quantify0or1(n))
		}

		// Union op
		if c == '|' {
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
Quantifiers *,+,? - Second
Concatenation ab - Third
Alternation | - Fourth

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
			for {
				topOp, err := operatorStack.Peak()
				if err != nil {
					operatorStack.Push(currentChar)
					break
				}
				if topOp == '*' || topOp == '+' || topOp == '?' {
					operatorStack.Pop()
					outputQueue = append(outputQueue, topOp)
					continue
				}
				operatorStack.Push(currentChar)
				break
			}
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
		if (isRegexChar(next) || next == '(') && (curr == ')' || isRegexChar(curr)) || isQuantifier(curr) {
			if !isQuantifier(curr) && !isQuantifier(next) {
				preparedChars = append(preparedChars, curr, '&')
			} else {
				preparedChars = append(preparedChars, curr)
			}
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
