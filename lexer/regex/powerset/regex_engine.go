package powerset

type ReMatch struct {
	StartIndex int
	EndIndex   int
	Match      string
	Labels     []string
}

func Match(search string, dfa DFA) (match ReMatch, hasMatch bool) {
	chars := []rune(search)
	state := 0
	latestMatch := ReMatch{0, -1, "", nil}
	if dfa.FinalStates.Bits[0]&1<<0 != 0 {
		latestMatch.EndIndex = 0
	}
	for i, c := range chars {
		charClass := dfa.Classifier.Classify(c)
		if charClass == -1 {
			break
		}
		nextState := dfa.Transitions[state][charClass]
		if nextState == -1 {
			break
		}
		state = nextState
		// Found result
		if dfa.FinalStates.IsSet(uint(state)) {
			latestMatch.EndIndex = i
			latestMatch.Labels = dfa.FinalStateLabels[state]
		}
	}
	if latestMatch.EndIndex != -1 {
		latestMatch.Match = string(chars[latestMatch.StartIndex : latestMatch.EndIndex+1])
	}
	return latestMatch, latestMatch.EndIndex != -1
}
