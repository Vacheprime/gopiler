package main

import (
	"fmt"

	// "github.com/Vacheprime/gopiler/lexer"
	"github.com/Vacheprime/gopiler/lexer/regex"
	"github.com/Vacheprime/gopiler/lexer/regex/glushkov"
	"github.com/Vacheprime/gopiler/lexer/regex/powerset"
)

func main() {
	expr, err := regex.RegexToParseTree(`[a-zA-Z]\w*`)
	expr2, err := regex.RegexToParseTree(`if|else|for|while`)
	if err != nil {
		panic(err)
	}
	lblTree1 := glushkov.LabelledRegexTree{Root: expr, Label: "FIRST"}
	lblTree2 := glushkov.LabelledRegexTree{Root: expr2, Label: "SECOND"}
	combinedInfo := glushkov.BuildSymbolInformation([]glushkov.LabelledRegexTree{lblTree1, lblTree2})
	// symInfo := glushkov.BuildSymbolInformation(expr)
	// symInfo2 := glushkov.BuildSymbolInformation(expr2)

	nfa, err := glushkov.BuildNFA(combinedInfo)
	if err != nil {
		panic(err)
	}
	dfa := powerset.BuildDFA(nfa)
	search := "my_var8"
	match, ok := powerset.Match(search, dfa)
	if ok {
		fmt.Println(match.Match)
		fmt.Println(match.Labels)
		fmt.Println(len(match.Match))
	} else {
		fmt.Println("No match!")
	}

	// defs, err := lexer.ParseDefinitions("../token_regexp.txt")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(defs)
}
