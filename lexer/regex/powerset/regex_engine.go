package powerset

type ReMatch struct {
	StartIndex int
	EndIndex   int
	Match      string
}

func Match(search string, dfa DFA) (ReMatch, bool) {
	chars := []rune(search)
	state := 0
	latestMatch := ReMatch{0, -1, ""}
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
		if dfa.FinalStates.IsSet(uint(state)) {
			latestMatch.EndIndex = i
		}
	}
	if latestMatch.EndIndex != -1 {
		latestMatch.Match = string(chars[latestMatch.StartIndex : latestMatch.EndIndex+1])
	}
	return latestMatch, latestMatch.EndIndex != -1
}
