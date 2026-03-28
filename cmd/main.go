package main

import (
	"github.com/Vacheprime/gopiler"
)

func main() {
	tokens := gopiler.GetTokens("(2 * 8))")
	expr, err := gopiler.ParseTokens(&tokens)
	if err != nil {
		panic(err)
	}
	// Generate code
	code := gopiler.GenerateCode(expr)
	gopiler.InterpretCode(code)
}
