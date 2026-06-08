package glushkov

import (
	"fmt"
	"strconv"
	"strings"

	gp "github.com/Vacheprime/gopiler"
	re "github.com/Vacheprime/gopiler/lexer/regex"
)

type SymbolOccurrences map[*re.RegexToken]Symbol

type IdCharacterRange struct {
	re.CharacterClass
	id int
}

type SymbolClassifier struct {
	Singulars map[rune]int
	Classes   []IdCharacterRange
	SymToId   map[Symbol]int
}

/* Returns the total number of character classes. */
func (sc *SymbolClassifier) Total() int {
	return len(sc.Singulars) + len(sc.Classes)
}

func BuildClassifier(occurrences SymbolOccurrences) (SymbolClassifier, error) {
	classifier := SymbolClassifier{make(map[rune]int), []IdCharacterRange{}, map[Symbol]int{}}
	idx := 0
	var anyCharClass *IdCharacterRange
	for _, sym := range occurrences {
		tk := sym.Token
		switch tk.Class {
		case re.SINGLE_CHAR:
			id, ok := classifier.Singulars[tk.Repr[0]]
			if !ok {
				classifier.Singulars[tk.Repr[0]] = idx
				id = idx
				idx++
			}
			classifier.SymToId[sym] = id
		case re.CHAR_CLASS:
			charClass, err := re.NewCharacterClass(*tk)
			if err != nil {
				return SymbolClassifier{}, re.ErrInvalidCharClass
			}
			classifier.SymToId[sym] = idx
			classifier.Classes = append(classifier.Classes, IdCharacterRange{charClass, idx})
			idx++
		case re.ANY_CHAR:
			if anyCharClass != nil {
				classifier.SymToId[sym] = anyCharClass.id
				continue
			}
			excludedRunes := []rune{'\r', '\n'}
			charClass := re.CharacterClass{Singulars: excludedRunes, Ranges: []re.CharacterRange{}, Excludes: true}
			anyCharClass = &IdCharacterRange{charClass, idx}
			classifier.SymToId[sym] = idx
			idx++
		}
	}
	// Append any character class at the end for least priority
	if anyCharClass != nil {
		classifier.Classes = append(classifier.Classes, *anyCharClass)
	}
	return classifier, nil
}

func (sc *SymbolClassifier) Classify(r rune) []int {
	classes := []int{}
	// Check if rune is a part of singular (highest priority)
	id, ok := sc.Singulars[r]
	if ok {
		classes = append(classes, id)
	}
	// Check if rune is a part of any one or more character classes (second priority)
outer:
	for _, cri := range sc.Classes {
		// Check for match in singulars
		for _, v := range cri.Singulars {
			if v == r {
				if cri.Excludes {
					continue outer
				}
				classes = append(classes, cri.id)
				continue outer
			}
		}

		// Check for match in ranges
		for _, cr := range cri.Ranges {
			if r >= cr.Start && r <= cr.End {
				if cri.Excludes {
					continue outer
				}
				classes = append(classes, cri.id)
				continue outer
			}
		}
		classes = append(classes, cri.id) // Not in excludes
	}
	return classes
}

// Attempt to implement a table-based NFA.
// The NFA would store a table array where the
// row is the state, column the input char class, and the value is the set of states.
type NFA struct {
	Classifier  SymbolClassifier
	Transitions [][]gp.BitSet
	FinalStates gp.BitSet
}

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
	Occurrences  SymbolOccurrences
	AcceptsEmpty bool
}

func AddSymbolToOccurrences(tk *re.RegexToken, occurrences SymbolOccurrences) Symbol {
	s := Symbol{len(occurrences) + 1, tk}
	occurrences[tk] = s
	return s
}

func BuildSymbolInformation(reRootExpr *re.Expression) SymbolInformation {
	symPairs := gp.NewSet[SymbolPair]()
	return determineSymbolPairs(reRootExpr, SymbolInformation{SymbolPairs: &symPairs, Occurrences: map[*re.RegexToken]Symbol{}})
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
	symToState := map[Symbol]int{}
	finalStates := gp.NewBitSet(uint(totalStates))
	// Create all nodes
	transitionIdx := 1
	transitions := make([][]gp.BitSet, totalStates)
	transitions[0] = make([]gp.BitSet, totalClassIds)
	for _, sym := range symInfo.Occurrences {
		symToState[sym] = transitionIdx
		transitions[transitionIdx] = make([]gp.BitSet, totalClassIds)
		// Set final states
		if symInfo.FinalSymbols.Contains(sym) {
			finalStates.Set(uint(transitionIdx))
		}
		transitionIdx++
	}
	// Create transitions for initial state
	state := transitions[0]
	for _, sym := range symInfo.StartSymbols.Items() {
		stateIdx := symToState[sym]
		classId := classifier.SymToId[sym]
		bitSet := state[classId]
		if bitSet.Bits == nil {
			bitSet = gp.NewBitSet(uint(totalStates))
			state[classId] = bitSet
		}
		bitSet.Set(uint(stateIdx))
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

			// Get class ID associated with next state
			classId := classifier.SymToId[p.S2]
			bitSet := state[classId]
			if bitSet.Bits == nil {
				bitSet = gp.NewBitSet(uint(totalStates))
				state[classId] = bitSet
			}

			// Add transition
			bitSet.Set(uint(stateIdx))
		}
	}
	return NFA{classifier, transitions, finalStates}, nil
}

func filterPairsByStart(sym Symbol, pairs []SymbolPair) []SymbolPair {
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
