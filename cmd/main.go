package main

import (
	"fmt"
	"io"
	"strings"

	// "github.com/Vacheprime/gopiler/lexer"
	"github.com/Vacheprime/gopiler/lexer"
	"github.com/Vacheprime/gopiler/lexer/regex"
	"github.com/Vacheprime/gopiler/lexer/regex/glushkov"
	"github.com/Vacheprime/gopiler/lexer/regex/powerset"
)

func main() {
	defs, err := lexer.ParseDefinitions("../token_regexp.txt")
	if err != nil {
		panic(err)
	}
	lblTrees := []glushkov.LabelledRegexTree{}
	for _, def := range defs {
		expr, err := regex.RegexToParseTree(def.Regex)
		if err != nil {
			panic(err)
		}
		lblTree := glushkov.LabelledRegexTree{
			Root:  expr,
			Label: def.Identifier,
		}
		lblTrees = append(lblTrees, lblTree)
	}
	symInfo := glushkov.BuildSymbolInformation(lblTrees)
	nfa, err := glushkov.BuildNFA(symInfo)
	if err != nil {
		panic(err)
	}
	dfa := powerset.BuildTableDFA(nfa)

	input := `
		int a = 6
	`
	reader := io.NopCloser(strings.NewReader(input))

	matcher := powerset.NewDFAMatcher(reader, dfa)
	fmt.Println(defs)
	lx := lexer.NewLexer(matcher, defs)
	for {
		tk, err := lx.NextToken()
		if err != nil {
			panic(err)
		}
		fmt.Printf("TOKEN: %+v\n", tk)
		if tk.TkType == lexer.EOF {
			break
		}
	}
}
