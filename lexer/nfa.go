package lexer

import (
	"errors"

	"github.com/Vacheprime/gopiler"
)

var (
	ErrUnmatchedParenthesis = errors.New("right parenthesis unmatched.")
)

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
func RegexToPostfix(regex string) (*gopiler.Stack[rune], error) {
	chars := prepareRegexString(regex)
	outputQueue := gopiler.NewStack[rune]()
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
				outputQueue.Push(topOp)
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
					outputQueue.Push(topOp)
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
					outputQueue.Push(topOp)
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
					outputQueue.Push(topOp)
					continue
				}
				operatorStack.Push(currentChar)
				break
			}
		default:
			outputQueue.Push(currentChar)
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
		outputQueue.Push(topOp)
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
