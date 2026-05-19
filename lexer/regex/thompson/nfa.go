package thompson

import (
	"errors"

	"github.com/Vacheprime/gopiler"
	r "github.com/Vacheprime/gopiler/lexer/regex"
)

const (
	VALID_KEY   = -1
	INVALID_KEY = -2
)

var (
	ErrUnmatchedParenthesis = errors.New("right parenthesis unmatched.")
	ErrUnmatchedBracket     = errors.New("left bracket of character class unmatched.")
	ErrInvalidCharClass     = errors.New("invalid character class")
)

type CharacterRange struct {
	start rune
	end   rune
}

type DeriveTransitionKey func(c rune) int32

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

func createGroupDerivator(t r.RegexToken) (DeriveTransitionKey, error) {
	acceptedRanges := []CharacterRange{}
	acceptedSingulars := []rune{}
	for i := 1; i < len(t.Repr)-1; i++ {
		curr := t.Repr[i]
		next := t.Repr[i+1]
		if next == '-' {
			// Check for end of range
			if i+2 >= len(t.Repr) {
				return nil, ErrInvalidCharClass
			}
			// Process the end of range
			end := t.Repr[i+2]
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

func PostfixToNFA(postfixTokens []r.RegexToken) (*NFA, error) {
	nfaFrags := gopiler.NewStack[*NFA]()
	for i := range postfixTokens {
		t := postfixTokens[i]

		// Simple State
		if t.Class == r.SINGLE_CHAR {
			nfaFrags.Push(createSingleCharNFA(t.Repr[0]))
			continue
		}

		// Char class
		if t.Class == r.CHAR_CLASS {
			d, err := createGroupDerivator(t)
			if err != nil {
				return nil, err
			}
			nfaFrags.Push(createGroupedNFA(d))
		}

		switch t.Repr[0] {
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
