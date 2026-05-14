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
	// testStack := gopiler.NewStack[rune]()
	// val, err := testStack.Pop()
	// if err == gopiler.ErrEmptyStack {
	// 	fmt.Println("An error occurred")
	// }
	// fmt.Println(val)
	// dfa := lexer.NewDFA("regeax")
	// match := dfa.Matches("reeaxaa")
	// if match != nil {
	// 	fmt.Println(*match)
	// } else {
	// 	fmt.Println("No results!")
	// }
	postfix, err := lexer.RegexToPostfix("aabb")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(*postfix))
}
