package powerset

import (
	"slices"

	gp "github.com/Vacheprime/gopiler"
	gl "github.com/Vacheprime/gopiler/lexer/regex/glushkov"
)

type DFA struct {
	Classifier       gl.SymbolClassifier
	Transitions      [][]int // A transition only has one next state
	FinalStates      gp.BitSet
	FinalStateLabels map[int][]string
	// Imagine a map that associates a final state bit position to a label given during construction
	// Upon match, the label is retrieved and returned along with the match.
}

type DFAState struct {
	nfaStates gp.BitSet
	Pos       int
}

func BuildDFA(nfa gl.NFA) DFA {
	TOTAL_STATES := uint(len(nfa.Transitions))
	TOTAL_CHAR_CLASSES := nfa.Classifier.Total()
	dfa := DFA{Classifier: nfa.Classifier}
	dfaTransitions := make([][]int, 0)
	dfaTransitions = append(dfaTransitions, make([]int, TOTAL_CHAR_CLASSES)) // Initial State

	annotatedTransitions := []DFAState{}
	dfaStateStack := gp.NewStack[DFAState]()
	dfaStateCounter := 0

	startNfaStates := gp.NewBitSet(TOTAL_STATES)
	startNfaStates.Set(0)
	dfaStateStack.Push(DFAState{startNfaStates, dfaStateCounter})
	dfaStateCounter++
	for dfaStateStack.Len() > 0 {
		currState, _ := dfaStateStack.Pop()
		nfaStateIndexes := currState.nfaStates.GetActiveBitPositions()

		currDfaTransitions := make([]DFAState, TOTAL_CHAR_CLASSES)

		// Build transitions for the current dfa state.
		// Map character class from classifier to sets of NFA states.
		for _, pos := range nfaStateIndexes {
			nfaState := nfa.Transitions[pos]

			for idx, nfaTransition := range nfaState {
				if nfaTransition.Bits == nil {
					continue
				}
				dfaState := currDfaTransitions[idx]
				if dfaState.nfaStates.IsNil() {
					dfaState.nfaStates = gp.NewBitSet(TOTAL_STATES)
				}
				dfaState.nfaStates.Or(nfaTransition)
				currDfaTransitions[idx] = dfaState
			}
		}

		// Build the final DFA transition table
		for idx, dfaState := range currDfaTransitions {
			if dfaState.nfaStates.IsNil() {
				dfaTransitions[currState.Pos][idx] = -1
				continue
			}
			foundIdx := findDfaState(dfaState, annotatedTransitions)
			if foundIdx == -1 {
				dfaState.Pos = dfaStateCounter
				dfaStateStack.Push(dfaState)
				annotatedTransitions = append(annotatedTransitions, dfaState)
				dfaTransitions = append(dfaTransitions, make([]int, TOTAL_CHAR_CLASSES))
				dfaStateCounter++
			} else {
				dfaState = annotatedTransitions[foundIdx]
			}
			dfaTransitions[currState.Pos][idx] = dfaState.Pos
		}
	}

	// While building the dfa final states, the label of the regex that produced the match can be
	// found by finding the label in the nfa
	dfaFinalStates := gp.NewBitSet(uint(len(dfaTransitions)))
	dfaFinalStateLabels := map[int][]string{}
	for _, dfaState := range annotatedTransitions {
		overlapPositions := nfa.FinalStates.OverlapsPos(dfaState.nfaStates)
		if len(overlapPositions) != 0 {
			labels := []string{}
			for _, pos := range overlapPositions {
				labels = append(labels, nfa.FinalStateLabels[int(pos)])
			}
			dfaFinalStateLabels[dfaState.Pos] = labels
			dfaFinalStates.Set(uint(dfaState.Pos))
		}
	}
	if nfa.FinalStates.Bits[0]&1<<0 != 0 {
		dfaFinalStates.Set(0)
	}
	dfa.Transitions = dfaTransitions
	dfa.FinalStates = dfaFinalStates
	dfa.FinalStateLabels = dfaFinalStateLabels
	return dfa
}

func findDfaState(state DFAState, transitions []DFAState) int {
	idx := slices.IndexFunc(transitions, func(s DFAState) bool {
		if state.nfaStates.Equals(s.nfaStates) {
			return true
		}
		return false
	})
	return idx
}
