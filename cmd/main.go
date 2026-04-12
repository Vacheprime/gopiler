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
	dfa := lexer.NewDFA("regex")
	fmt.Println(*dfa.Matches("regex"))
}
