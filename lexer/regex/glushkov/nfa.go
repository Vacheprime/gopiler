package glushkov

import (
	gp "github.com/Vacheprime/gopiler"
	re "github.com/Vacheprime/gopiler/lexer/regex"
)

type Symbol struct {
	SymbolId int
	Token    *re.RegexToken
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

func addSymbolToAlphabet(tk *re.RegexToken, alphabet map[*re.RegexToken]Symbol) Symbol {
	s := Symbol{len(alphabet), tk}
	alphabet[tk] = s
	return s
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

func determineStartSymbols(reRootExpr *re.Expression, alphabet map[*re.RegexToken]Symbol) gp.Set[Symbol] {
	symbols := gp.NewSet[Symbol]()
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
			s, ok := alphabet[e.Atom]
			if !ok {
				s = addSymbolToAlphabet(e.Atom, alphabet)
			}
			symbols.Add(s)
		}
	}
	return symbols
}

func determineFinalSymbols(reRootExpr *re.Expression, alphabet map[*re.RegexToken]Symbol) gp.Set[Symbol] {
	symbols := gp.NewSet[Symbol]()
	exprStack := gp.NewStack[*re.Expression]()
	exprStack.Push(reRootExpr)
	for exprStack.Len() > 0 {
		e, _ := exprStack.Pop()
		switch e.Type {
		case re.BINARY_EXPR:
			switch e.Operator {
			case '|':
				exprStack.Push(e.LExpr, e.RExpr)
			case '&':
				right := e.RExpr
				if right.Type == re.UNARY_EXPR && (right.Operator == '*' || right.Operator == '?') {
					exprStack.Push(e.LExpr)
				}
				exprStack.Push(right)
			}
		case re.UNARY_EXPR:
			exprStack.Push(e.LExpr)
		case re.ATOMIC:
			s, ok := alphabet[e.Atom]
			if !ok {
				s = addSymbolToAlphabet(e.Atom, alphabet)
			}
			symbols.Add(s)
		}
	}
	return symbols
}
