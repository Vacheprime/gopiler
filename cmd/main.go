package main

import (
	"fmt"

	// "github.com/Vacheprime/gopiler/lexer"
	"github.com/Vacheprime/gopiler/lexer/regex"
	"github.com/Vacheprime/gopiler/lexer/regex/glushkov"
	"github.com/Vacheprime/gopiler/lexer/regex/powerset"
)

func main() {
	expr, err := regex.RegexToParseTree(`\q`)
	if err != nil {
		panic(err)
	}
	symInfo := glushkov.BuildSymbolInformation(expr)
	nfa, err := glushkov.BuildNFA(symInfo)
	if err != nil {
		panic(err)
	}
	dfa := powerset.BuildDFA(nfa)
	match, ok := powerset.Match(`\n`, dfa)
	if ok {
		fmt.Println(match.Match)
	} else {
		fmt.Println("No match!")
	}

	// defs, err := lexer.ParseDefinitions("../token_regexp.txt")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(defs)
}
