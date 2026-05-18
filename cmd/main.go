package main

import (
	"fmt"

	"github.com/Vacheprime/gopiler/lexer"
)

func main() {
	post, err := lexer.RegexToPostfix("(a*)*")
	if err != nil {
		panic(err)
	}
	nfa, err := lexer.PostfixToNFA(*post)
	fmt.Println(string(*post))
	match, ok := lexer.Match(nfa, "abe")
	if ok {
		fmt.Println(match)
	} else {
		fmt.Println("No match!")
	}
}
