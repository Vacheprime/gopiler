package powerset

type ReMatch struct {
	StartIndex int
	EndIndex   int
	Match      string
	Label      string
}

func Match(search string, dfa DFA) (match ReMatch, hasMatch bool) {
	chars := []rune(search)
	state := 0
	latestMatch := ReMatch{0, -1, "", ""}
	if dfa.FinalStates.Bits[0]&1<<0 != 0 {
		latestMatch.EndIndex = 0
	}
	for i, c := range chars {
		charClasses := dfa.Classifier.Classify(c)
		if len(charClasses) == 0 {
			break
		}
		nextState := -1
		for _, charClass := range charClasses {
			nextState = dfa.Transitions[state][charClass]
			if nextState != -1 {
				break
			}
		}
		if nextState == -1 {
			break
		}
		state = nextState
		// Found result
		if dfa.FinalStates.IsSet(uint(state)) {
			latestMatch.EndIndex = i
			latestMatch.Label = dfa.FinalStateLabels[state]
		}
	}
	if latestMatch.EndIndex != -1 {
		latestMatch.Match = string(chars[latestMatch.StartIndex : latestMatch.EndIndex+1])
	}
	return latestMatch, latestMatch.EndIndex != -1
}
