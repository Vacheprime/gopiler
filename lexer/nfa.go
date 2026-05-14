package lexer

import (
	"errors"

	"github.com/Vacheprime/gopiler"
)

var (
	ErrUnmatchedParenthesis = errors.New("right parenthesis unmatched.")
)

/*
NFA as array of arrays?
[

	0 (State): [{1, 2}, {3}, {2}] // Input needs to be converted into index. Node metadata not available
	1 (State): [...]

]

NFA as array of maps?
[

	0 (State): map[]{}

]

NFA as connected Nodes?
*/
type NFANode struct {
	isAccepting bool
	position    int
	transitions map[rune]*[]NFANode
	matchChar   rune
}

func PostfixToNFA(postfixChars []rune) {
	stack := gopiler.NewStack[*NFANode]()
	nodeCount := 0
	for i := range postfixChars {
		c := postfixChars[i]

		// Push C if is a char to match
		if isRegexChar(c) {
			newNode := NFANode{position: nodeCount, isAccepting: false, transitions: make(map[rune]*[]NFANode), matchChar: c}
			stack.Push(&newNode)
			nodeCount++
		}

		// Handle concat
		if c == '&' {
			// Pop 2 previous nodes
			prev2, _ := stack.Pop()
			prev1, _ := stack.Pop()
			// Connect together via transition
			prev1.transitions[prev2.matchChar] = prev2
			// Push back prev2
			stack.Push(prev2)
		}

		// Handle Union
		if c == '|' {
			// Pop 2 previous
			prev2, _ := stack.Pop()
			prev1, _ := stack.Pop()
			// Create start
			startNode := NFANode{position: -1, isAccepting: false, transitions: make(map[rune]*[]NFANode)}
			//startNode.transitions[''] =
		}
	}
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
	outputQueue := make([]rune, len(chars))
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
Prepares a regex for postfix convertion by adding
'&' to denote concatenation.
*/
func prepareRegexString(regex string) []rune {
	chars := []rune(regex)
	preparedChars := make([]rune, 0)
	for i := 0; i < len(chars)-1; i++ {
		curr := chars[i]
		next := chars[i+1]
		if (isRegexChar(next) || next == '(') && (curr == ')' || isRegexChar(curr)) || isQuantifier(curr) {
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
