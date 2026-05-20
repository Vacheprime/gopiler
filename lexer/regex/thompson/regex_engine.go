package thompson

import "github.com/Vacheprime/gopiler"

/*
Searches the search string for a match given an NFA.
*/
func Match(nfa *NFA, search string) (string, bool) {
	chars := []rune(search)
	epsilonSet := gopiler.NewSet[*NFANode]()
	nbrChars, ok := findMatch(nfa.startNode, nfa.acceptStates[0], chars, &epsilonSet)
	if ok {
		return string(chars[:nbrChars]), true
	}
	return "", false
}

func findMatch(node *NFANode, accept *NFANode, chars []rune, epsilonNodes *gopiler.Set[*NFANode]) (int, bool) {
	runeCount := 0
	biggestMatch := 0
	matches := false
	if len(chars) > 0 && node.derivator != nil {
		c := chars[0]
		key := node.derivator(c)
		if nextNodes, ok := node.transitions[key]; ok {
			// Clear upon char consumption
			epsilonNodes.Clear()
			runeCount++
			for _, v := range nextNodes.Items() {
				nbrChars, ok := findMatch(v, accept, chars[1:], epsilonNodes)
				if !ok {
					continue
				}
				matches = true
				if nbrChars > biggestMatch {
					biggestMatch = nbrChars
				}
			}
		}
	}

	if node.epsilonTransitions.Len() > 0 {
		for _, next := range node.epsilonTransitions.Items() {
			if next == node || epsilonNodes.Contains(next) {
				continue
			}
			epsilonNodes.Add(next)
			nbrChars, ok := findMatch(next, accept, chars, epsilonNodes)
			if !ok {
				continue
			}
			matches = true
			if nbrChars > biggestMatch {
				biggestMatch = nbrChars
			}
		}
	}
	// Also clear upon dead-end
	if !matches {
		epsilonNodes.Clear()
	}
	runeCount += biggestMatch
	return runeCount, matches || node == accept
}
