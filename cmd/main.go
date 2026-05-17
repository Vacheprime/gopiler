package main

import (
	"fmt"

	"github.com/Vacheprime/gopiler/lexer"
)

func main() {
	postfix, err := lexer.RegexToPostfix("a+b")
	if err != nil {
		panic(err)
	}
	post := []rune{'a', '?', 'b'}
	nfa, err := lexer.PostfixToNFA(post)
	fmt.Println(string(*postfix))
	match, ok := lexer.Match(nfa, "ab")
	if ok {
		fmt.Println(match)
	} else {
		fmt.Println("No match!")
	}
}
