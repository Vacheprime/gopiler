package main

import (
	"fmt"

	"github.com/Vacheprime/gopiler/lexer"
)

func main() {
	// tokens := gopiler.GetTokens("(2 * 8))")
	// expr, err := gopiler.ParseTokens(&tokens)
	// if err != nil {
	// 	panic(err)
	// }
	// // Generate code
	// code := gopiler.GenerateCode(expr)
	// gopiler.InterpretCode(code)
	dfa := lexer.NewDFA("regeax")
	match := dfa.Matches("reeaxaa")
	if match != nil {
		fmt.Println(*match)
	} else {
		fmt.Println("No results!")
	}
}
