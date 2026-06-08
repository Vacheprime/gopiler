package main

import (
	"fmt"

	"github.com/Vacheprime/gopiler/lexer"
)

func main() {
	// expr, _ := re.RegexToParseTree(``)
	// symInfo := gl.BuildSymbolInformation(expr)
	// nfa, err := gl.BuildNFA(symInfo)
	// if err != nil {
	// 	panic(err)
	// }
	// dfa := pw.BuildDFA(nfa)
	// match, ok := pw.Match("01000000000000", dfa)
	// if ok {
	// 	fmt.Println(match.Match)
	// } else {
	// 	fmt.Println("No match!")
	// }
	// fmt.Println(len(dfa.Transitions))
	// fmt.Println(len(nfa.Transitions))
	// fmt.Println(dfa.FinalStates.ToBinaryString())
	defs, err := lexer.ParseDefinitions("../token_regexp.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println(defs)
}
