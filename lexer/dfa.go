package lexer

// Implement the regex: /abc[0-9]x/

// DFA would look like s0 (a) -> s1(b) -> s2(c) -> s3(digit 0-9) -> s4(x)
const SEARCH_STR string = " tsabc8sxsbc7x "

// Manual Match would look something like this
func ManualMatches() bool {
	chars := []rune(SEARCH_STR)
	currentState := 0
	for i := 0; i < len(chars); i++ {
		currentChar := chars[i]
		if currentState == 0 && currentChar == 'a' {
			currentState = 1
		} else if currentState == 1 && currentChar == 'b' {
			currentState = 2
		} else if currentState == 2 && currentChar == 'c' {
			currentState = 3
		} else if currentState == 3 && (currentChar >= 48 && currentChar <= 57) {
			currentState = 4
		} else if currentState == 4 && currentChar == 'x' {
			return true
		} else {
			currentState = 0
		}
	}
	return false
}

type State struct {
	Position    int
	NextStates  map[string]*State
	IsAccepting bool
}

type DFA struct {
	FirstState State
}

func NewDFA(regex string) *DFA {
	regexChars := []rune(regex)
	return &DFA{FirstState: *createState(regexChars, 0)}
}

func createState(regex []rune, idx int) *State {
	// Determine next possible states based on current index.
	if idx == len(regex)-1 {
		return &State{Position: idx, NextStates: nil, IsAccepting: true}
	}
	nextChar := regex[idx]

	switch nextChar {
	case '[':
		return nil
	default:
		idx++
		var nextState *State
		if idx >= len(regex) {
			nextState = nil
		} else {
			nextState = createState(regex, idx)
		}
		return &State{Position: idx, NextStates: map[string]*State{string(nextChar): createState(regex, idx+1)}}
	}
}

func (dfa *DFA) Matches(search string) *string {
	chars := []rune(search)
	// currentState := &dfa.FirstState
	// startIndex := 0
	// for i := 0; i < len(chars); i++ {
	// 	currentState = currentState.Transition(chars[i], &dfa.FirstState)
	// 	if currentState == nil {
	// 		substr := string(chars[startIndex : i+1])
	// 		return &substr
	// 	}
	// 	if currentState.Position == 0 {
	// 		startIndex = i
	// 	}
	// }
	// return nil
	startIndex := 0
	currentState := dfa.FirstState
	for i := 0; i < len(chars); i++ {
		possibleNext, ok := currentState.NextStates[string(chars[i])]
		if !ok {
			currentState = dfa.FirstState
			startIndex = 0
		}
		currentState = *possibleNext
	}
}
