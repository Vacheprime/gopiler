package glushkov

import (
	"fmt"
	"strconv"
	"strings"

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
	SymbolPairs  *gp.Set[SymbolPair]
	Occurences   SymbolOccurences
}

func AddSymbolToOccurences(tk *re.RegexToken, occurences SymbolOccurences) Symbol {
	s := Symbol{len(occurences) + 1, tk}
	occurences[tk] = s
	return s
}

func BuildSymbolInformation(reRootExpr *re.Expression) SymbolInformation {
	symPairs := gp.NewSet[SymbolPair]()
	return determineSymbolPairs(reRootExpr, SymbolInformation{SymbolPairs: &symPairs, Occurences: map[*re.RegexToken]Symbol{}})
}

func determineSymbolPairs(expr *re.Expression, symInfo SymbolInformation) SymbolInformation {
	switch expr.Type {
	case re.ATOMIC:
		// For atomic expressions, right and left reachables correspond to
		// the symbol of the atomic expression
		s, ok := symInfo.Occurences[expr.Atom]
		if !ok {
			s = AddSymbolToOccurences(expr.Atom, symInfo.Occurences)
		}
		symInfo.StartSymbols.Add(s)
		symInfo.FinalSymbols.Add(s)
	case re.UNARY_EXPR:
		// Get reachables from sub expr
		subExprReachables := determineSymbolPairs(expr.LExpr, symInfo)

		switch expr.Operator {
		case '*':
			// For kleene star, the sympairs that can be determined are the cartesian product of
			// right reachable and left reachable sets (Rr X Lr). Those are the repetition combinations.
			for _, vL := range subExprReachables.FinalSymbols.Items() {
				for _, vR := range subExprReachables.StartSymbols.Items() {
					sp := SymbolPair{vL, vR}
					symInfo.SymbolPairs.Add(sp)
				}
			}
			// For kleene star, left and right reachables correspond to
			// the left and right reachables of the sub expression
			symInfo.StartSymbols.Add(subExprReachables.StartSymbols.Items()...)
			symInfo.FinalSymbols.Add(subExprReachables.FinalSymbols.Items()...)
		case '?':
			// For 0 or 1 quantifier, no sympairs can be computed.
			// Forward reachable states up.
		case '+':
			panic("+ NOT IMPLEMENTED YET")
		default:
			panic("unimplemented unary operator")
		}
	case re.BINARY_EXPR:
		// Get reachables from left and right sub expressions
		leftExprRs := determineSymbolPairs(expr.LExpr, symInfo)
		rightExprRs := determineSymbolPairs(expr.RExpr, symInfo)

		switch expr.Operator {
		case '&':
			// Compute sympairs. They correspond to cartesian product of left with right
			// reachables. (Right reachables of left expr, Left reachables of right expr).
			for _, vLR := range leftExprRs.FinalSymbols.Items() {
				for _, vRL := range rightExprRs.StartSymbols.Items() {
					sp := SymbolPair{vLR, vRL}
					symInfo.SymbolPairs.Add(sp)
				}
			}

			// For concatenation, left and right reachables depend on whether the sub exprs
			// are optional or not.
			symInfo.StartSymbols.Add(leftExprRs.StartSymbols.Items()...)
			symInfo.FinalSymbols.Add(rightExprRs.FinalSymbols.Items()...)

			if re.IsOptionalExpr(*expr.LExpr) {
				// Add left reachables of right expression to left reachables of this expr
				symInfo.StartSymbols.Add(rightExprRs.StartSymbols.Items()...)
			}
			if re.IsOptionalExpr(*expr.RExpr) {
				// Add right reachables of left expression to right reachables of this expr
				symInfo.FinalSymbols.Add(leftExprRs.FinalSymbols.Items()...)
			}
		case '|':
			// No symbol pairs can be computed for alternation since it represents Lexpr or Rexpr
			// Reachables correspond to the addition of left and right reachables from both expressions
			symInfo.StartSymbols.Add(leftExprRs.StartSymbols.Items()...)
			symInfo.StartSymbols.Add(rightExprRs.StartSymbols.Items()...)
			symInfo.FinalSymbols.Add(leftExprRs.FinalSymbols.Items()...)
			symInfo.FinalSymbols.Add(rightExprRs.FinalSymbols.Items()...)
		default:
			panic("unsupported binary operator")
		}
	}
	return symInfo
}

func PrintSymbolPairs(pairs []SymbolPair) {
	var b strings.Builder
	// 0[(0a)]
	for i, v := range pairs {
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
