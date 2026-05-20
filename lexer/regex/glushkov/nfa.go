package glushkov

import (
	gp "github.com/Vacheprime/gopiler"
	re "github.com/Vacheprime/gopiler/lexer/regex"
)

type Symbol struct {
	SymbolId int
	Repr     *[]rune
}

type SymbolPair struct {
	S1 Symbol
	S2 Symbol
}

type SymbolInformation struct {
	StartSymbols gp.Set[Symbol]
	FinalSymbols gp.Set[Symbol]
	SymbolPairs  gp.Set[SymbolPair]
}

func NewSymbolInformation() SymbolInformation {
	return SymbolInformation{gp.NewSet[Symbol](), gp.NewSet[Symbol](), gp.NewSet[SymbolPair]()}
}

func BuildSymbolInformation(postfixTokens []re.RegexToken) SymbolInformation {
	info := NewSymbolInformation()
	stack := gp.NewStack[re.RegexToken]()
	for i := range postfixTokens {
		tk := postfixTokens[i]

		switch tk.Class {
		case re.ANY_CHAR, re.SINGLE_CHAR, re.CHAR_CLASS:
			// Add to start symbols if first
			if stack.Len() == 0 {

			}
		}
	}
	return info
}

func determineStartSymbols(reRootExpr *re.Expression) []Symbol {
	symbols := []Symbol{}
	exprStack := gp.NewStack[*re.Expression]()
	exprStack.Push(reRootExpr)
	for exprStack.Len() > 0 {
		// Get next expr to process
		e, _ := exprStack.Pop()
		switch e.Type {
		case re.BINARY_EXPR:
			switch e.Operator {
			case '|':
				exprStack.Push(e.RExpr, e.LExpr)
			case '&':
				left := e.LExpr
				if left.Type == re.UNARY_EXPR && (left.Operator == '*' || left.Operator == '?') {
					exprStack.Push(e.RExpr)
				}
				exprStack.Push(e.LExpr)
			}
		case re.UNARY_EXPR:
			exprStack.Push(e.LExpr)
		case re.ATOMIC:
			s := Symbol{len(symbols), &e.Atom.Repr}
			symbols = append(symbols, s)
		}
	}
	return symbols
}
