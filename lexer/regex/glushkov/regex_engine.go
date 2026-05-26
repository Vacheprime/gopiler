package glushkov

import gp "github.com/Vacheprime/gopiler"

type ReMatch struct {
	StartIndex int
	EndIndex   int
	Match      string
}

func Match(search string, nfa NFA) (ReMatch, bool) {
	chars := []rune(search)
	latestMatch := ReMatch{-1, -1, ""}
	// Initialize current states with state 1
	currentStates := gp.NewBitSet(uint(len(nfa.Transitions)))
	currentStates.Set(0)
	for i, c := range chars {
		// Get rune class ID
		cIds := nfa.Classifier.Classify(c)

		// Loop over every state
		activeStates := currentStates.GetActiveBitPositions()
		newStates := gp.NewBitSet(uint(len(nfa.Transitions)))
		for _, state := range activeStates {
			// get next states
			var nextStates gp.BitSet
			for _, cId := range cIds {
				s := nfa.Transitions[state][cId]
				if s.Bits != nil {
					nextStates = s
					break
				}
			}

			// Check if a transition exists
			if nextStates.Bits == nil {
				continue
			}

		}

	}
}
