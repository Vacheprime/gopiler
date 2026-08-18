package regex

import (
	"math"
	"slices"
	"unicode"
)

// CharacterRange represents a range of accepted characters.
type CharacterRange struct {
	Start rune
	End   rune
}

func (cr CharacterRange) Matches(r rune) bool {
	return cr.Start <= r && r <= cr.End
}

func MergeCharacterRanges(ranges []CharacterRange) (mergedRanges []CharacterRange) {
	rangesCopy := slices.Clone(ranges)
	if len(rangesCopy) == 0 || len(rangesCopy) == 1 {
		return rangesCopy
	}
	SortCharacterRanges(&rangesCopy)

	mergedRanges = append(mergedRanges, rangesCopy[0])
	for _, currentRange := range rangesCopy[1:] {
		latestSimpRange := &mergedRanges[len(mergedRanges)-1]

		// Handle distinct ranges (no overlap and not adjacent)
		if currentRange.Start-1 > latestSimpRange.End {
			mergedRanges = append(mergedRanges, currentRange)
			continue
		}

		// Handle total overlap
		if currentRange.End <= latestSimpRange.End {
			continue
		}

		// Handle partial overlap by extending the range
		latestSimpRange.End = currentRange.End
	}
	return mergedRanges
}

func SortCharacterRanges(ranges *[]CharacterRange) {
	slices.SortFunc(*ranges, func(r1 CharacterRange, r2 CharacterRange) int {
		return int(r1.Start) - int(r2.Start)
	})
}

// CharacterSet represents a regex character set.
//
// It is comprised of a list of character ranges that constitute the
// set of characters that matches the set.
type CharacterSet struct {
	Ranges []CharacterRange
}

func (cs *CharacterSet) RangeIsFullyContained(cr CharacterRange) bool {
	for _, setRange := range cs.Ranges {
		if setRange.Start <= cr.Start && cr.End <= setRange.End {
			return true
		}
	}
	return false
}

// Matches determines whether a rune r matches the CharacterSet.
func (cc *CharacterSet) Matches(r rune) (matches bool) {
	for _, cs := range cc.Ranges {
		if cs.Start <= r && r <= cs.End {
			return true
		}
	}
	return false
}

// Equals checks if a CharacterClass is equal to this CharacterClass.
//
// Deprecated: Character sets are comparable by default.
func (cc1 *CharacterSet) Equals(cc2 CharacterSet) (isEqual bool) {
	if len(cc1.Ranges) != len(cc2.Ranges) {
		return false
	}
	// Need to loop over because ranges may be located
	// at different indexes in both arrays.
	for _, s := range cc1.Ranges {
		if !slices.Contains(cc2.Ranges, s) {
			return false
		}
	}
	return true
}

// BuildCharacterSet creates a new character set from a slice of character ranges.
//
// This method is the preferred method to use when building a set
// since it simplifies and merges ranges to obtain the simplest set
// possible and it also keeps all ranges sorted by start value.
func BuildCharacterSet(ranges []CharacterRange) CharacterSet {
	mergedRanges := MergeCharacterRanges(ranges)
	return CharacterSet{Ranges: mergedRanges}
}

// NegateCharacterSet computes the negated set of the set given.
//
// This method assumes that the character ranges of the set given are sorted.
func NegateCharacterSet(cs CharacterSet) (negatedSet CharacterSet) {
	var intervalStart int32 = 0
	negatedSet = CharacterSet{[]CharacterRange{}}
	for _, cr := range cs.Ranges {
		intervalEnd := cr.Start - 1
		if intervalEnd < intervalStart {
			intervalStart = int32(math.Min(unicode.MaxRune, float64(cr.End+1)))
			continue // Skip, no interval possible
		}
		negatedRange := CharacterRange{Start: intervalStart, End: intervalEnd}
		negatedSet.Ranges = append(negatedSet.Ranges, negatedRange)

		intervalStart = int32(math.Min(unicode.MaxRune, float64(cr.End+1)))
	}
	// Add last range which includes the rest of the characters
	if intervalStart != unicode.MaxRune {
		lastRange := CharacterRange{Start: intervalStart, End: unicode.MaxRune}
		negatedSet.Ranges = append(negatedSet.Ranges, lastRange)
	}
	return negatedSet
}

// NewCharacterSet creates a new character set from a CHAR_CLASS or META_CLASS regex token.
func NewCharacterSet(tk RegexToken) (CharacterSet, error) {
	// Handle meta characters
	if tk.Class == META_CHAR {
		return getMetaClass(tk.Repr[0])
	}
	ranges := []CharacterRange{}
	isNegated := false

	for i := 1; i < len(tk.Repr)-1; i++ {
		curr := tk.Repr[i]
		next := tk.Repr[i+1]
		// Handle negated classes
		if i == 1 && curr == '^' {
			isNegated = true
			continue
		}
		if curr == '\\' {
			c, isMeta, consumed, err := parseEscapeSequence(tk.Repr[i:])
			if err != nil {
				return CharacterSet{}, ErrInvalidEscapeSequence
			}
			if isMeta {
				metaClass, err := getMetaClass(c)
				if err != nil {
					return CharacterSet{}, err
				}
				ranges = append(ranges, metaClass.Ranges...)
			} else {
				sing := CharacterRange{Start: c, End: c}
				ranges = append(ranges, sing)
			}
			i += consumed - 1
			continue
		}
		if next == '-' {
			// Check for end of range
			if i+2 >= len(tk.Repr) {
				return CharacterSet{}, ErrInvalidCharClass
			}
			// Process the end of range
			end := tk.Repr[i+2]
			if end < curr {
				return CharacterSet{}, ErrInvalidCharClass
			}
			// Append range
			cr := CharacterRange{curr, end}
			ranges = append(ranges, cr)
			i += 2 // Skip 2 to land after char range
			continue
		}
		sing := CharacterRange{Start: curr, End: curr}
		ranges = append(ranges, sing)
	}
	charSet := BuildCharacterSet(ranges)
	if isNegated {
		charSet = NegateCharacterSet(charSet)
	}
	return charSet, nil
}

func getMetaClass(char rune) (CharacterSet, error) {
	isNegated := unicode.IsUpper(char)
	switch char {
	case 'w', 'W':
		ranges := []CharacterRange{
			{Start: '0', End: '9'},
			{Start: 'A', End: 'Z'},
			{Start: '_', End: '_'},
			{Start: 'a', End: 'z'},
		}
		regularSet := CharacterSet{Ranges: ranges}
		if isNegated {
			return NegateCharacterSet(regularSet), nil
		}
		return regularSet, nil
	case 's', 'S':
		ranges := []CharacterRange{
			{Start: '\t', End: '\t'},
			{Start: '\n', End: '\n'},
			{Start: '\v', End: '\v'},
			{Start: '\f', End: '\f'},
			{Start: '\r', End: '\r'},
			{Start: ' ', End: ' '},
		}
		regularSet := CharacterSet{Ranges: ranges}
		if isNegated {
			return NegateCharacterSet(regularSet), nil
		}
		return regularSet, nil
	case 'd', 'D':
		ranges := []CharacterRange{
			{Start: '0', End: '9'},
		}
		regularSet := CharacterSet{Ranges: ranges}
		if isNegated {
			return regularSet, nil
		}
		return regularSet, nil
	case '.':
		ranges := []CharacterRange{
			{Start: '\n', End: '\n'},
			{Start: '\r', End: '\r'},
		}
		regularSet := CharacterSet{Ranges: ranges}
		return NegateCharacterSet(regularSet), nil
	}
	return CharacterSet{}, ErrUnknownMetaClass
}
