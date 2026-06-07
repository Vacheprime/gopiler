package main

import (
	"fmt"

	re "github.com/Vacheprime/gopiler/lexer/regex"
	gl "github.com/Vacheprime/gopiler/lexer/regex/glushkov"
	pw "github.com/Vacheprime/gopiler/lexer/regex/powerset"
)

func main() {
	expr, _ := re.RegexToParseTree(``)
	symInfo := gl.BuildSymbolInformation(expr)
	nfa, err := gl.BuildNFA(symInfo)
	if err != nil {
		panic(err)
	}
	dfa := pw.BuildDFA(nfa)
	match, ok := pw.Match("01000000000000", dfa)
	if ok {
		fmt.Println(match.Match)
	} else {
		fmt.Println("No match!")
	}
	fmt.Println(len(dfa.Transitions))
	fmt.Println(len(nfa.Transitions))
	fmt.Println(dfa.FinalStates.ToBinaryString())
}
