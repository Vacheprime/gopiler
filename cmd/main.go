package main

import (
	"fmt"

	re "github.com/Vacheprime/gopiler/lexer/regex"
	th "github.com/Vacheprime/gopiler/lexer/regex/thompson"
)

func main() {
	post, err := re.RegexToPostfix("([+][(][0-9]+[)])? [0-9][0-9][0-9] [0-9][0-9][0-9]-[0-9][0-9][0-9][0-9]")
	if err != nil {
		panic(err)
	}
	nfa, err := th.PostfixToNFA(post)
	if err != nil {
		panic(err)
	}

	match, ok := th.Match(nfa, "+(135) 514 872-8373")
	if ok {
		fmt.Println(match)
	} else {
		fmt.Println("No match!")
	}
}
