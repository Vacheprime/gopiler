package main

import (
	"fmt"

	re "github.com/Vacheprime/gopiler/lexer/regex"
	gl "github.com/Vacheprime/gopiler/lexer/regex/glushkov"
)

func main() {
	expr, _ := re.RegexToParseTree("(a|b)*[a]")
	symInfo := gl.BuildSymbolInformation(expr)
	gl.PrintSymbolPairs(symInfo.SymbolPairs.Items())
	nfa, err := gl.BuildNFA(symInfo)
	if err != nil {
		panic(err)
	}
	match, ok := gl.Match("aa", nfa)
	if !ok {
		fmt.Println("No match!")
	} else {
		fmt.Println(match.Match)
	}
}
