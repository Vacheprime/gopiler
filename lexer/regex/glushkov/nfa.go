package glushkov

import (
	gp "github.com/Vacheprime/gopiler"
	re "github.com/Vacheprime/gopiler/lexer/regex"
)

type Alphabet map[*re.RegexToken]Symbol

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

func addSymbolToAlphabet(tk *re.RegexToken, alphabet Alphabet) Symbol {
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

func determineAlphabet(reRootExpr *re.Expression) Alphabet {
	alphabet := make(map[*re.RegexToken]Symbol)
	exprStack := gp.NewStack[*re.Expression]()
	exprStack.Push(reRootExpr)
	for exprStack.Len() > 0 {
		e, _ := exprStack.Pop()
		switch e.Type {
		case re.BINARY_EXPR:
			exprStack.Push(e.RExpr, e.LExpr)
		case re.UNARY_EXPR:
			exprStack.Push(e.LExpr)
		case re.ATOMIC:
			_, ok := alphabet[e.Atom]
			if !ok {
				addSymbolToAlphabet(e.Atom, alphabet)
			}
		}
	}
	return alphabet
}

func determineStartSymbols(reRootExpr *re.Expression, alphabet Alphabet) gp.Set[Symbol] {
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

func determineFinalSymbols(reRootExpr *re.Expression, alphabet Alphabet) gp.Set[Symbol] {
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

// type frame struct {
// 	Expr    *re.Expression
// 	Visited bool
// }

// func determineSymbolPairs(reRootExpr *re.Expression, alphabet Alphabet) gp.Set[Symbol] {
// 	symPairs := gp.NewSet[SymbolPair]()
// 	frameStack := gp.NewStack[frame]()
// 	firstFrame := frame{reRootExpr, false}
// 	frameStack.Push(firstFrame)
// 	possiblePairs := []Symbol{}
// 	for frameStack.Len() > 0 {
// 		f, _ := frameStack.Pop()

// 		// Stop Condition
// 		if f.Expr.Type == re.ATOMIC {
// 			// Lookup symbol in occurences
// 			s, _ := alphabet[f.Expr.Atom]
// 			// Add to possiblePairs
// 			possiblePairs = append(possiblePairs, s)
// 		}

// 		if f.Visited {
// 			// Unwind
// 			// Build possible pairs based on type

// 		} else {
// 			// Push current
// 			frameStack.Push(frame{f.Expr, true})

// 			// Push next based on type
// 			switch f.Expr.Type {
// 			case re.UNARY_EXPR:
// 				frameStack.Push(frame{f.Expr.LExpr, false})
// 			case re.BINARY_EXPR:
// 				frameStack.Push(frame{f.Expr.RExpr, false}, frame{f.Expr.LExpr, false})
// 			}
// 		}
// 	}
// }

func determineSymbolPairs(symPairs gp.Set[SymbolPair], expr *re.Expression, alphabet Alphabet) (gp.Set[SymbolPair], []Symbol) {
	possibleSymbols := []Symbol{}
	switch expr.Type {
	case re.ATOMIC:
		// Add possible symbol
		s, _ := alphabet[expr.Atom]
		possibleSymbols = append(possiblePairs, s)
	case re.UNARY_EXPR:
		// Get possible pairs from sub expr
		_, exprPossibleSymbols := determineSymbolPairs(symPairs, expr.LExpr, alphabet)
		// Build pair last + first
		sp := SymbolPair{exprPossibleSymbols[0], exprPossibleSymbols[len(exprPossibleSymbols)-1]}
		symPairs.Add(sp)
	}
	return symPairs, possiblePairs
}
