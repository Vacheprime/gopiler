package main

import (
	"fmt"

	re "github.com/Vacheprime/gopiler/lexer/regex"
	th "github.com/Vacheprime/gopiler/lexer/regex/thompson"
)

func main() {
	post, err := re.RegexToPostfix("{.}")
	if err != nil {
		panic(err)
	}
	nfa, err := th.PostfixToNFA(post)
	if err != nil {
		panic(err)
	}

	match, ok := th.Match(nfa, "{.}")
	if ok {
		fmt.Println(match)
	} else {
		fmt.Println("No match!")
	}
}
