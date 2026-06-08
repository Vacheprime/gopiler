package main

import (
	"fmt"

	"github.com/Vacheprime/gopiler/lexer"
	"github.com/Vacheprime/gopiler/lexer/regex"
	"github.com/Vacheprime/gopiler/lexer/regex/glushkov"
	"github.com/Vacheprime/gopiler/lexer/regex/powerset"
)

func main() {
	expr, _ := regex.RegexToParseTree(".")
	symInfo := glushkov.BuildSymbolInformation(expr)
	nfa, err := glushkov.BuildNFA(symInfo)
	if err != nil {
		panic(err)
	}
	dfa := powerset.BuildDFA(nfa)
	match, ok := powerset.Match("\r", dfa)
	if ok {
		fmt.Println(match.Match)
	} else {
		fmt.Println("No match!")
	}
	fmt.Println(len(dfa.Transitions))
	fmt.Println(len(nfa.Transitions))
	fmt.Println(dfa.FinalStates.ToBinaryString())

	defs, err := lexer.ParseDefinitions("../token_regexp.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println(defs)
}
