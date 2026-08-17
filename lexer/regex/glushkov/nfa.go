package glushkov

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	gp "github.com/Vacheprime/gopiler"
	re "github.com/Vacheprime/gopiler/lexer/regex"
)

type SymbolOccurrences map[*re.RegexToken]re.Symbol

type LabelledRegexTree struct {
	Root  *re.Expression
	Label string
}

// Attempt to implement a table-based NFA.
// The NFA would store a table array where the
// row is the state, column the input char class, and the value is the set of states.
type NFA struct {
	Classifier       SymbolClassifier
	Transitions      [][]gp.BitSet
	FinalStates      gp.BitSet
	FinalStateLabels map[int]string
}

// SymbolPair represents a pair of symbols.
//
// Used in Glushkov's construction for keeping track
// of which symbols can reach which other symbols.
type SymbolPair struct {
	S1 re.Symbol
	S2 re.Symbol
}

type SymbolInformation struct {
	StartSymbols gp.Set[re.Symbol]
	FinalSymbols gp.Set[re.Symbol]
	SymbolPairs  *gp.Set[SymbolPair]
	Occurrences  SymbolOccurrences
	AcceptsEmpty bool
}

// ContainsFinalSymbol checks if a given symbol is in the FinalSymbols list.
// This method ignores the label field.
func (si *SymbolInformation) ContainsFinalSymbol(sym re.Symbol) bool {
	return slices.ContainsFunc(si.FinalSymbols.Items(), func(s re.Symbol) bool {
		return s.Token == sym.Token
	})
}

// GetFinalSymbolLabel returns the label of the final symbol given a symbol.
// The symbol given MUST be inside the final symbol list.
func (si *SymbolInformation) GetFinalSymbolLabel(sym re.Symbol) string {
	finalSymbols := si.FinalSymbols.Items()
	idx := slices.IndexFunc(finalSymbols, func(s re.Symbol) bool {
		return s.Token == sym.Token
	})
	return finalSymbols[idx].Label
}

func AddSymbolToOccurrences(tk *re.RegexToken, occurrences SymbolOccurrences) re.Symbol {
	id := len(occurrences)
	s, _ := re.BuildSymbol(tk, id, "")
	occurrences[tk] = s
	return s
}

func BuildSymbolInformation(rootExprs []LabelledRegexTree) SymbolInformation {
	symPairs := gp.NewSet[SymbolPair]()
	symOccurrences := map[*re.RegexToken]re.Symbol{}
	symInfo := SymbolInformation{SymbolPairs: &symPairs, Occurrences: symOccurrences}
	for _, labelledTree := range rootExprs {
		// Create new symbol information to have a clean list of start and final symbol for this tree
		newInfo := SymbolInformation{SymbolPairs: &symPairs, Occurrences: symInfo.Occurrences}
		newInfo = determineSymbolPairs(labelledTree.Root, newInfo)

		// Attach label to final states
		finalSymbols := newInfo.FinalSymbols.Items()
		for i := range finalSymbols {
			finalSymbols[i].Label = labelledTree.Label
		}

		// Handle empty
		if symInfo.AcceptsEmpty {
			newInfo.AcceptsEmpty = true
		}

		// Add back initial start and final symbols as well as
		newInfo.StartSymbols.Add(symInfo.StartSymbols.Items()...)
		newInfo.FinalSymbols.Add(symInfo.FinalSymbols.Items()...)

		symInfo = newInfo
	}
	return symInfo
}

func determineSymbolPairs(expr *re.Expression, symInfo SymbolInformation) SymbolInformation {
	switch expr.Type {
	case re.ATOMIC:
		// For atomic expressions, right and left reachables correspond to
		// the symbol of the atomic expression
		s, ok := symInfo.Occurrences[expr.Atom]
		if !ok {
			s = AddSymbolToOccurrences(expr.Atom, symInfo.Occurrences)
		}
		symInfo.StartSymbols.Add(s)
		symInfo.FinalSymbols.Add(s)
	case re.UNARY_EXPR:
		// Get reachables from sub expr
		subExprReachables := determineSymbolPairs(expr.LExpr, symInfo)

		switch expr.Operator {
		case '*', '+':
			// For kleene star, the sympairs that can be determined are the cartesian product of
			// right reachable and left reachable sets (Rr X Lr). Those are the repetition combinations.
			for _, vL := range subExprReachables.FinalSymbols.Items() {
				for _, vR := range subExprReachables.StartSymbols.Items() {
					sp := SymbolPair{vL, vR}
					symInfo.SymbolPairs.Add(sp)
				}
			}
			// * and ? accept empty string
			if expr.Operator == '*' {
				symInfo.AcceptsEmpty = true
			}
		case '?':
			symInfo.AcceptsEmpty = true
		default:
		}
		// For Unary expressions, left and right reachables correspond to
		// the left and right reachables of the sub expression
		symInfo.StartSymbols.Add(subExprReachables.StartSymbols.Items()...)
		symInfo.FinalSymbols.Add(subExprReachables.FinalSymbols.Items()...)
	case re.BINARY_EXPR:
		// Get reachables from left and right sub expressions
		leftExprRs := determineSymbolPairs(expr.LExpr, symInfo)
		rightExprRs := determineSymbolPairs(expr.RExpr, symInfo)

		switch expr.Operator {
		case '&':
			// Concat binary expr accept empty strings if both the left and right expr accept
			// the empty string
			if leftExprRs.AcceptsEmpty && rightExprRs.AcceptsEmpty {
				symInfo.AcceptsEmpty = true
			}

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
			// Alternate binary expr accept empty strings if either the left or right expr accept
			// the empty string
			if leftExprRs.AcceptsEmpty || rightExprRs.AcceptsEmpty {
				symInfo.AcceptsEmpty = true
			}
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

func BuildNFA(symInfo SymbolInformation) (NFA, error) {
	classifier, err := BuildClassifier(symInfo.Occurrences)
	if err != nil {
		return NFA{}, err
	}
	totalStates := len(symInfo.Occurrences) + 1
	totalClassIds := classifier.Total()
	symToState := map[re.Symbol]int{}
	finalStates := gp.NewBitSet(uint(totalStates))
	finalStateLabels := map[int]string{}
	// Create all nodes
	transitionIdx := 1
	transitions := make([][]gp.BitSet, totalStates)
	transitions[0] = make([]gp.BitSet, totalClassIds)
	for _, sym := range symInfo.Occurrences {
		symToState[sym] = transitionIdx
		transitions[transitionIdx] = make([]gp.BitSet, totalClassIds)
		// Set final states and attach label
		if symInfo.ContainsFinalSymbol(sym) {
			finalStateLabels[transitionIdx] = symInfo.GetFinalSymbolLabel(sym)
			finalStates.Set(uint(transitionIdx))
		}
		transitionIdx++
	}
	// Create transitions for initial state
	state := transitions[0]
	for _, sym := range symInfo.StartSymbols.Items() {
		stateIdx := symToState[sym]

		classIds := classifier.GetClassIdsFromSymbol(sym)
		for _, classId := range classIds {
			bitSet := state[classId]
			if bitSet.Bits == nil {
				bitSet = gp.NewBitSet(uint(totalStates))
				state[classId] = bitSet
			}
			bitSet.Set(uint(stateIdx))
		}

	}
	// Add initial state to final states if the expr accepts empty strings
	if symInfo.AcceptsEmpty {
		finalStates.Set(0)
	}
	// Create all transitions
	for _, sym := range symInfo.Occurrences {
		// Current state
		state := transitions[symToState[sym]]
		// Get related pairs
		relatedPairs := filterPairsByStart(sym, symInfo.SymbolPairs.Items())
		for _, p := range relatedPairs {
			// Get state ID of next state
			stateIdx := symToState[p.S2]

			// Get class IDs associated with next state
			classIds := classifier.GetClassIdsFromSymbol(p.S2)
			for _, classId := range classIds {
				bitSet := state[classId]
				if bitSet.Bits == nil {
					bitSet = gp.NewBitSet(uint(totalStates))
					state[classId] = bitSet
				}
				// Add transition
				bitSet.Set(uint(stateIdx))
			}
		}
	}
	return NFA{classifier, transitions, finalStates, finalStateLabels}, nil
}

func filterPairsByStart(sym re.Symbol, pairs []SymbolPair) []SymbolPair {
	filteredPairs := []SymbolPair{}
	for _, p := range pairs {
		if p.S1 == sym {
			filteredPairs = append(filteredPairs, p)
		}
	}
	return filteredPairs
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
