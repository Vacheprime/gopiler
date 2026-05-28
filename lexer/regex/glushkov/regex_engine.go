package glushkov

import (
	gp "github.com/Vacheprime/gopiler"
)

type ReMatch struct {
	StartIndex int
	EndIndex   int
	Match      string
}

func Match(search string, nfa NFA) (ReMatch, uint, bool) {
	chars := []rune(search)
	totalStates := uint(len(nfa.Transitions))
	latestMatch := ReMatch{0, -1, ""}
	// Initialize current states with state 1
	currentStates := gp.NewBitSet(totalStates)
	currentStates.Set(0)
	if nfa.FinalStates.Bits[0]&1<<0 != 0 {
		latestMatch.EndIndex = 0
	}
	var consumedChars uint = 0
	for i, c := range chars {
		cIds := nfa.Classifier.Classify(c)
		if len(cIds) == 0 {
			break // No transitions possible
		}

		// Loop over every state
		activeStates := currentStates.GetActiveBitPositions()
		newStates := gp.NewBitSet(totalStates)
		for _, state := range activeStates {
			// get next states
			var nextStates gp.BitSet
			for _, cId := range cIds {
				s := nfa.Transitions[state][cId]
				if s.Bits != nil {
					if nextStates.Bits == nil {
						nextStates = gp.NewBitSet(totalStates)
					}
					nextStates.Or(s)
				}
			}

			// Check if a transition exists
			if nextStates.Bits == nil {
				continue
			}

			// Add next states
			newStates.Or(nextStates)
		}

		if newStates.IsZero() {
			break // No transitions possible
		}
		// Char counts as consumed if it advances the automaton
		consumedChars++

		// Set the new states
		currentStates = newStates

		// Check for a match
		if currentStates.Overlaps(nfa.FinalStates) {
			latestMatch.EndIndex = i
			latestMatch.Match = string(chars[latestMatch.StartIndex : latestMatch.EndIndex+1])
		}
	}
	return latestMatch, consumedChars, latestMatch.EndIndex != -1
}
