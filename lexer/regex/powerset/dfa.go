package powerset

import (
	gp "github.com/Vacheprime/gopiler"
	gl "github.com/Vacheprime/gopiler/lexer/regex/glushkov"
)

type DFA struct {
	Classifier  gl.SymbolClassifier
	Transitions [][]int // A transition only has one next state
	FinalStates gp.BitSet
}

func BuildDFA(nfa gl.NFA) DFA {
	// Reusing the NFA classifier as it is identical
	dfa := DFA{Classifier: nfa.Classifier}
	dfaTransitions := make([][]int, 0)

	return dfa
}
