package main

import (
	"fmt"

	"github.com/Vacheprime/gopiler"
	re "github.com/Vacheprime/gopiler/lexer/regex"
	gl "github.com/Vacheprime/gopiler/lexer/regex/glushkov"
)

func main() {
	regex := "[a-b]c"
	parseTree, err := re.RegexToParseTree(regex)
	if err != nil {
		panic(err)
	}
	occurences := gl.DetermineSymbolOccurences(parseTree)
	symPairs := gopiler.NewSet[gl.SymbolPair]()
	gl.DetermineSymbolPairs(&symPairs, parseTree, occurences)
	fmt.Println(symPairs.Len())
	gl.PrintSymbolPairs(symPairs)
}
