package powerset

import (
	"bufio"
	"errors"
	"io"
)

var (
	ErrUnexpectedIOError = errors.New("an unexpected IO error occurred while reading input")
)

// Matcher matches regular expressions from a source reader.
//
// Implementations must not skip any characters from the input
// source. If the next characters do not match anything, an error
// should be produced.
type SequentialMatcher interface {
	// MatchNext finds the next match in the string.
	//
	// ok should be false if the next characters could not be matched
	// or if a read error occurred.
	//
	// Following a failed call to MatchNext, subsequent calls to the function
	// should skip the unrecognizable characters and attempt to match the rest
	// of the input.
	MatchNext() (nextMatch ReMatch, ok bool)

	// Error returns the latest error recorded, usually following a failed
	// call to MatchNext.
	Error() (lastError error)

	// CurrentPosition returns the position as a 0-based index of the next
	// character that will be read.
	CurrentPosition() (position int)

	io.Closer
}

// ReMatch represents a regex match.
//
// The start and end index are the 0-based index of the
// first and last rune of the matched string.
//
// The match is the matched string itself and the labels
// are the labels of the regular expressions that produced
// the match.
type ReMatch struct {
	StartIndex int
	EndIndex   int
	Match      string
	Labels     []string
}

// ReplayRuneReader wraps a buffered reader and allows replaying or rereading
// multiple characters from the reader.
type ReplayRuneReader struct {
	r           io.ReadCloser
	br          *bufio.Reader
	replayRunes []rune
}

// NewReplayRuneReader creates a ReplayRuneReader from a ReadCloser.
func NewReplayRuneReader(r io.ReadCloser) *ReplayRuneReader {
	return &ReplayRuneReader{
		r:           r,
		br:          bufio.NewReader(r),
		replayRunes: []rune{},
	}
}

// NextRune reads the next rune from the reader or returns the earliest rune
// added as a replay if there are any runes the replay.
func (rr *ReplayRuneReader) NextRune() (rune, error) {
	if len(rr.replayRunes) > 0 {
		r := rr.replayRunes[0]
		rr.replayRunes = rr.replayRunes[1:]
		return r, nil
	}
	r, _, err := rr.br.ReadRune()
	return r, err
}

// ReplayRunes add runes to be reread or replayed.
//
// The first rune in the slice is the first to be reread.
func (rr *ReplayRuneReader) ReplayRunes(runes []rune) {
	rr.replayRunes = append(rr.replayRunes, runes...)
}

// Close closes the underlying ReadCloser.
func (rr *ReplayRuneReader) Close() error {
	return rr.r.Close()
}

// DFAMatcher implements the SequentialMatcher interface using a DFA.
//
// Although a DFA can match empty strings, this DFAMatcher does not consider
// an empty string as a valid match.
type DFAMatcher struct {
	dfa          DFA
	replayReader *ReplayRuneReader
	latestError  error
	runePosition int
}

// NewDFAMatcher returns a DFAMatcher from a DFA and ReadCloser.
func NewDFAMatcher(r io.ReadCloser, dfa DFA) DFAMatcher {
	return DFAMatcher{
		dfa:          dfa,
		replayReader: NewReplayRuneReader(r),
		latestError:  nil,
		runePosition: 0,
	}
}

func (m DFAMatcher) Error() error {
	return m.latestError
}

func (m DFAMatcher) CurrentPosition() (position int) {
	return m.runePosition
}

func (m *DFAMatcher) Close() error {
	return m.replayReader.Close()
}

func (m *DFAMatcher) MatchNext() (nextMatch ReMatch, ok bool) {
	nextMatch = ReMatch{
		StartIndex: m.runePosition,
		EndIndex:   -1,
		Match:      "",
		Labels:     nil,
	}
	matchedChars := []rune{}
	possibleReplays := []rune{}
	state := 0
	for {
		r, err := m.replayReader.NextRune()
		if err == io.EOF {
			m.latestError = io.EOF
			break
		} else if err != nil {
			m.latestError = ErrUnexpectedIOError
			break
		}
		m.advanceRunePosition()
		possibleReplays = append(possibleReplays, r)
		charClass := m.dfa.Classifier.Classify(r)
		if charClass == -1 {
			break
		}
		nextState := m.dfa.Transitions[state][charClass]
		if nextState == -1 {
			break
		}
		state = nextState
		matchedChars = append(matchedChars, r)
		if m.dfa.FinalStates.IsSet(uint(state)) {
			possibleReplays = possibleReplays[:0] // Clear
			nextMatch.EndIndex = m.runePosition
			nextMatch.Labels = m.dfa.FinalStateLabels[state]
		}
	}
	if nextMatch.EndIndex != -1 {
		nextMatch.Match = string(matchedChars[0 : nextMatch.EndIndex-nextMatch.StartIndex])
		m.rewindRunes(possibleReplays)
	}
	return nextMatch, nextMatch.EndIndex != -1
}

// advanceRunePosition moves the rune position forward by 1.
func (m *DFAMatcher) advanceRunePosition() {
	m.runePosition++
}

// rewindRunes adds runes to the replayReader and decrements the
// runePosition to match the rewind.
func (m *DFAMatcher) rewindRunes(runes []rune) {
	m.runePosition -= len(runes)
	m.replayReader.ReplayRunes(runes)
}
