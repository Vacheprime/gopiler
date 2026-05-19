package main

import (
	"fmt"

	"github.com/Vacheprime/gopiler/lexer"
)

func main() {
	post, err := lexer.RegexToPostfix("[0-9][0-9][0-9] [0-9][0-9][0-9]-[0-9][0-9][0-9][0-9]")
	if err != nil {
		panic(err)
	}
	nfa, err := lexer.PostfixToNFA(post)
	if err != nil {
		panic(err)
	}

	match, ok := lexer.Match(nfa, "514 872-8373")
	if ok {
		fmt.Println(match)
	} else {
		fmt.Println("No match!")
	}
}
