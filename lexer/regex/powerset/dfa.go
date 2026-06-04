package powerset

import (
	"slices"

	gp "github.com/Vacheprime/gopiler"
	gl "github.com/Vacheprime/gopiler/lexer/regex/glushkov"
)

type DFA struct {
	Classifier  gl.SymbolClassifier
	Transitions [][]int // A transition only has one next state
	FinalStates gp.BitSet
}

type dfaState struct {
	nfaStates gp.BitSet
	Pos       int
}

func BuildDFA(nfa gl.NFA) DFA {
	TOTAL_STATES := uint(len(nfa.Transitions))
	TOTAL_SYM_OCCURENCES := nfa.Classifier.Total()
	dfa := DFA{Classifier: nfa.Classifier}
	dfaTransitions := make([][]int, 0)

	annotatedTransitions := []dfaState{}
	dfaStateStack := gp.NewStack[dfaState]()
	dfaStateCounter := 0

	startNfaStates := gp.NewBitSet(TOTAL_STATES)
	startNfaStates.Set(0)
	dfaStateStack.Push(dfaState{startNfaStates, dfaStateCounter})

	for dfaStateStack.Len() > 0 {
		dfaStateCounter++

		currState, _ := dfaStateStack.Pop()
		nfaStateIndexes := currState.nfaStates.GetActiveBitPositions()
		dfaTransitions := make([]dfaState, TOTAL_SYM_OCCURENCES)
		for _, pos := range nfaStateIndexes {
			nfaState := nfa.Transitions[pos]

			for idx, nfaTransition := range nfaState {

			}
		}
	}

	return dfa
}

func findDfaState(state dfaState, transitions []dfaState) int {
	idx := slices.IndexFunc(transitions, func(s dfaState) bool {
		if state.nfaStates.Equals(s.nfaStates) {
			return true
		}
		return false
	})
	return idx
}
