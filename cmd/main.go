package main

import (
	"fmt"

	re "github.com/Vacheprime/gopiler/lexer/regex"
	gl "github.com/Vacheprime/gopiler/lexer/regex/glushkov"
)

func main() {
	regex := "(a|b)*c"
	parseTree, err := re.RegexToParseTree(regex)
	if err != nil {
		panic(err)
	}
	symInfo := gl.BuildSymbolInformation(parseTree)
	fmt.Println(symInfo.SymbolPairs.Len())
	fmt.Println(symInfo.StartSymbols.Len())
	gl.PrintSymbolPairs(symInfo.SymbolPairs.Items())
}
