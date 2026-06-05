package main

import (
	"fmt"

	re "github.com/Vacheprime/gopiler/lexer/regex"
	gl "github.com/Vacheprime/gopiler/lexer/regex/glushkov"
	pw "github.com/Vacheprime/gopiler/lexer/regex/powerset"
)

func main() {
	expr, _ := re.RegexToParseTree("(ab|ac)c")
	symInfo := gl.BuildSymbolInformation(expr)
	nfa, err := gl.BuildNFA(symInfo)
	if err != nil {
		panic(err)
	}
	dfa := pw.BuildDFA(nfa)
	fmt.Println(dfa.Transitions)
}
