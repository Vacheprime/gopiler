package main

import (
	"fmt"

	re "github.com/Vacheprime/gopiler/lexer/regex"
	gl "github.com/Vacheprime/gopiler/lexer/regex/glushkov"
)

func main() {
	expr, _ := re.RegexToParseTree("(ab|ac)c*[0-9A-Z]")
	symInfo := gl.BuildSymbolInformation(expr)
	nfa, err := gl.BuildNFA(symInfo)
	if err != nil {
		panic(err)
	}
	match, _, ok := gl.Match("accccY", nfa)
	if !ok {
		fmt.Println("No match!")
	} else {
		fmt.Println(match.Match)
	}
}
