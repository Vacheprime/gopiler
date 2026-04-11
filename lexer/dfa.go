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
	Transition func(rune, *State) *State
}

type DFA struct {
	Regex      string
	FirstState State
}

func (dfa *DFA) Matches() bool {
	dfa.buildStates()
	chars := []rune(SEARCH_STR)
	currentState := &dfa.FirstState
	for i := 0; i < len(chars); i++ {
		currentState = currentState.Transition(chars[i], &dfa.FirstState)
		if currentState == nil {
			return true
		}
	}
	return false
}

func (dfa *DFA) buildStates() {
	// Start with last state and go backwards
	// State 4
	state4Trans := func(char rune, firstState *State) *State {
		if char == 'x' {
			return nil
		}
		return firstState
	}
	state4 := State{Transition: state4Trans}
	// State 3
	state3 := State{Transition: func(char rune, firstState *State) *State {
		if char >= 48 && char <= 57 {
			return &state4
		}
		return firstState
	}}
	// State 2
	state2 := State{Transition: func(r rune, s *State) *State {
		if r == 'c' {
			return &state3
		}
		return s
	}}
	// State 1
	state1 := State{Transition: func(r rune, s *State) *State {
		if r == 'b' {
			return &state2
		}
		return s
	}}
	state0trans := func(r rune, s *State) *State {
		if r == 'a' {
			return &state1
		}
		return s
	}
	dfa.FirstState = State{Transition: state0trans}
}
