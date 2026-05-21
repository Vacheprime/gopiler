package glushkov

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Vacheprime/gopiler"
	gp "github.com/Vacheprime/gopiler"
	re "github.com/Vacheprime/gopiler/lexer/regex"
)

type SymbolOccurences map[*re.RegexToken]Symbol

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

func AddSymbolToOccurences(tk *re.RegexToken, occurences SymbolOccurences) Symbol {
	s := Symbol{len(occurences) + 1, tk}
	occurences[tk] = s
	return s
}

func NewSymbolInformation() SymbolInformation {
	return SymbolInformation{gp.NewSet[Symbol](), gp.NewSet[Symbol](), gp.NewSet[SymbolPair]()}
}

func BuildSymbolInformation(reRootExpr *re.Expression) SymbolInformation {
	occurences := DetermineSymbolOccurences(reRootExpr)
	startSymbols := DetermineStartSymbols(reRootExpr, occurences)
	finalSymbols := DetermineFinalSymbols(reRootExpr, occurences)
	symbolPairs := gp.NewSet[SymbolPair]()
	DetermineSymbolPairs(&symbolPairs, reRootExpr, occurences)
	return SymbolInformation{startSymbols, finalSymbols, symbolPairs}
}

func DetermineSymbolOccurences(reRootExpr *re.Expression) SymbolOccurences {
	occurences := make(map[*re.RegexToken]Symbol)
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
			_, ok := occurences[e.Atom]
			if !ok {
				AddSymbolToOccurences(e.Atom, occurences)
			}
		}
	}
	return occurences
}

func DetermineStartSymbols(reRootExpr *re.Expression, occurences SymbolOccurences) gp.Set[Symbol] {
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
			s, _ := occurences[e.Atom]
			symbols.Add(s)
		}
	}
	return symbols
}

func DetermineFinalSymbols(reRootExpr *re.Expression, occurences SymbolOccurences) gp.Set[Symbol] {
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
			s, _ := occurences[e.Atom]
			symbols.Add(s)
		}
	}
	return symbols
}

/*
Struct used to represent symbols that are reachable from the left and right of an expression.

Used to determine possible symbol pairs for expressions.
*/
type reachableSymbols struct {
	LeftReachable  gp.Set[Symbol]
	RightReachable gp.Set[Symbol]
}

func DetermineSymbolPairs(symPairs *gp.Set[SymbolPair], expr *re.Expression, occurences SymbolOccurences) reachableSymbols {
	reachables := reachableSymbols{}
	switch expr.Type {
	case re.ATOMIC:
		// For atomic expressions, right and left reachables correspond to
		// the symbol of the atomic expression
		s, _ := occurences[expr.Atom]
		reachables.LeftReachable.Add(s)
		reachables.RightReachable.Add(s)
	case re.UNARY_EXPR:
		// Get reachables from sub expr
		subExprReachables := DetermineSymbolPairs(symPairs, expr.LExpr, occurences)

		switch expr.Operator {
		case '*':
			// For kleene star, the sympairs that can be determined are the cartesian product of
			// right reachable and left reachable sets (Rr X Lr). Those are the repetition combinations.
			for _, vL := range subExprReachables.RightReachable.Items() {
				for _, vR := range subExprReachables.LeftReachable.Items() {
					sp := SymbolPair{vL, vR}
					symPairs.Add(sp)
				}
			}
			// For kleene star, left and right reachables correspond to
			// the left and right reachables of the sub expression
			reachables.LeftReachable.Add(subExprReachables.LeftReachable.Items()...)
			reachables.RightReachable.Add(subExprReachables.RightReachable.Items()...)
		case '?':
			// For 0 or 1 quantifier, no sympairs can be computed.
			// Forward reachable states up.
			reachables = subExprReachables
		case '+':
			panic("+ NOT IMPLEMENTED YET")
		default:
			panic("unimplemented unary operator")
		}
	case re.BINARY_EXPR:
		// Get reachables from left and right sub expressions
		leftExprRs := DetermineSymbolPairs(symPairs, expr.LExpr, occurences)
		rightExprRs := DetermineSymbolPairs(symPairs, expr.RExpr, occurences)

		switch expr.Operator {
		case '&':
			// Compute sympairs. They correspond to cartesian product of left with right
			// reachables. (Right reachables of left expr, Left reachables of right expr).
			for _, vLR := range leftExprRs.RightReachable.Items() {
				for _, vRL := range rightExprRs.LeftReachable.Items() {
					sp := SymbolPair{vLR, vRL}
					symPairs.Add(sp)
				}
			}

			// For concatenation, left and right reachables depend on whether the sub exprs
			// are optional or not.
			reachables.LeftReachable.Add(leftExprRs.LeftReachable.Items()...)
			reachables.RightReachable.Add(rightExprRs.RightReachable.Items()...)

			if re.IsOptionalExpr(*expr.LExpr) {
				// Add left reachables of right expression to left reachables of this expr
				reachables.LeftReachable.Add(rightExprRs.LeftReachable.Items()...)
			}
			if re.IsOptionalExpr(*expr.RExpr) {
				// Add right reachables of left expression to right reachables of this expr
				reachables.RightReachable.Add(leftExprRs.RightReachable.Items()...)
			}
		case '|':
			// No symbol pairs can be computed for alternation since it represents Lexpr or Rexpr
			// Reachables correspond to the addition of left and right reachables from both expressions
			reachables.LeftReachable.Add(leftExprRs.LeftReachable.Items()...)
			reachables.LeftReachable.Add(rightExprRs.LeftReachable.Items()...)
			reachables.RightReachable.Add(leftExprRs.RightReachable.Items()...)
			reachables.RightReachable.Add(rightExprRs.RightReachable.Items()...)
		default:
			panic("unsupported binary operator")
		}
	}
	return reachables
}

func PrintSymbolPairs(pairs gopiler.Set[SymbolPair]) {
	var b strings.Builder
	// 0[(0a)]
	for i, v := range pairs.Items() {
		b.WriteString(strconv.Itoa(i))
		b.WriteString("[(")
		b.WriteString(strconv.Itoa(v.S1.SymbolId))
		b.WriteString(string(v.S1.Token.Repr))
		b.WriteRune(')')
		b.WriteString(",(")
		b.WriteString(strconv.Itoa(v.S2.SymbolId))
		b.WriteString(string(v.S2.Token.Repr))
		b.WriteString(")] ")
	}
	fmt.Println(b.String())
}
