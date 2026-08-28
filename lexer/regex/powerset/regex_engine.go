package powerset

import (
	"bufio"
	"errors"
	"io"
	"slices"
	"unicode/utf8"

	gp "github.com/Vacheprime/gopiler"
)

var (
	ErrUnexpectedReaderError = errors.New("an unexpected IO error occurred while reading input")
	ErrInvalidUTF8Sequence   = errors.New("an invalid UTF-8 byte sequence was encountered")
)

// Matcher matches regular expressions from a source reader.
//
// Implementations must not skip any characters from the input
// source. If the next characters do not match anything, an error
// should be produced.
type SequentialMatcher interface {
	// MatchNext finds the next match in the string.
	//
	// err indicates whether any error was encountered during matching.
	//
	// io.EOF indicates end of input. The match object returned should be considered and can
	// be matching, non-matching or meaningless, indicated by an empty match string.
	// Subsequent calls to MatchNext should return EOF aswell.
	//
	// ErrInvalidUTF8Sequence indicates an encoding error. Since invalid bytes cannot possibly be matched,
	// the regex matches right before and right after can no longer be accurate. The behavior of MatchNext
	// following an invalid sequence is not defined. The match returned along with the error should reflect
	// the current state of the matching process. If a partial match was found, it should be returned. If no match
	// was found, the unmatched sequence should be returned.
	//
	// ErrUnexpectedReaderError indicates an error coming from the underlying reader that is unexpected (e.g. IO error)
	// and occurred while reading input. Since the matcher cannot consume characters to produce accurate matches, the behavior
	// of the MatchNext following this error is not defined. The match returned along with the error should reflect
	// the current state of the matching process. If a partial match was found, it should be returned. If no match
	// was found, the unmatched sequence should be returned.
	MatchNext() (nextMatch ReMatch, err error)

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
	IsMatching bool
}

// ReplayRuneReader wraps a buffered reader and allows replaying or rereading
// multiple characters from the reader.
type ReplayRuneReader struct {
	r           io.ReadCloser
	br          *bufio.Reader
	replayRunes gp.Stack[rune]
}

// NewReplayRuneReader creates a ReplayRuneReader from a ReadCloser.
func NewReplayRuneReader(r io.ReadCloser) *ReplayRuneReader {
	return &ReplayRuneReader{
		r:           r,
		br:          bufio.NewReader(r),
		replayRunes: gp.NewStack[rune](),
	}
}

// NextRune reads the next rune from the reader or returns the earliest rune
// added as a replay if there are any runes the replay.
func (rr *ReplayRuneReader) NextRune() (rune, error) {
	if rr.replayRunes.Len() > 0 {
		r, _ := rr.replayRunes.Pop()
		return r, nil
	}
	r, size, err := rr.br.ReadRune()
	if size == 1 && r == utf8.RuneError {
		err = ErrInvalidUTF8Sequence
	} else if err != nil && err != io.EOF {
		err = ErrUnexpectedReaderError
	}
	return r, err
}

// ReplayRunes add runes to be reread or replayed.
//
// The first rune in the slice is the first to be reread.
func (rr *ReplayRuneReader) ReplayRunes(runes []rune) {
	c := make([]rune, len(runes))
	copy(c, runes)
	slices.Reverse(c)
	rr.replayRunes.Push(c...)
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
	runePosition int
}

// NewDFAMatcher returns a DFAMatcher from a DFA and ReadCloser.
func NewDFAMatcher(r io.ReadCloser, dfa DFA) *DFAMatcher {
	return &DFAMatcher{
		dfa:          dfa,
		replayReader: NewReplayRuneReader(r),
		runePosition: 0,
	}
}

func (m DFAMatcher) CurrentPosition() (position int) {
	return m.runePosition
}

func (m *DFAMatcher) Close() error {
	return m.replayReader.Close()
}

func (m *DFAMatcher) MatchNext() (nextMatch ReMatch, err error) {
	nextMatch = ReMatch{
		StartIndex: m.runePosition,
		EndIndex:   m.runePosition,
		Match:      "",
		Labels:     nil,
		IsMatching: false,
	}
	matchedChars := []rune{}
	possibleReplays := []rune{}
	state := m.dfa.Start()
	for {
		r, rErr := m.replayReader.NextRune()
		if rErr != io.EOF && rErr != ErrInvalidUTF8Sequence {
			m.advanceRunePosition()
		}
		if rErr != nil {
			// Don't set error if reached EOF but there are still
			// characters that can be matched individually.
			if rErr == io.EOF && len(possibleReplays) > 0 && (!slices.Equal(matchedChars, possibleReplays) || !nextMatch.IsMatching) {
				break
			}
			err = rErr
			break
		}
		possibleReplays = append(possibleReplays, r)
		nextState := m.dfa.TransitionState(state, r)
		if nextState == -1 {
			break
		}
		state = nextState
		matchedChars = append(matchedChars, r)
		if m.dfa.IsAccepting(state) {
			nextMatch.IsMatching = true
			nextMatch.EndIndex = m.CurrentPosition() - 1 // Record position of last match
			possibleReplays = possibleReplays[:0]        // Clear
			nextMatch.Labels = m.dfa.AcceptingLabels(state)
		}
	}
	if nextMatch.IsMatching {
		nextMatch.Match = string(matchedChars[0 : nextMatch.EndIndex-nextMatch.StartIndex+1])
		m.rewindRunes(possibleReplays)
	} else if len(possibleReplays) > 0 {
		nextMatch.EndIndex = nextMatch.StartIndex + len(possibleReplays) - 1
		nextMatch.Match = string(possibleReplays)
		m.rewindRunes(possibleReplays[1:])
	}
	return nextMatch, err
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
