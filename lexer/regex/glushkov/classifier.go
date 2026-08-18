package glushkov

import (
	"container/heap"
	"maps"
	"slices"

	gp "github.com/Vacheprime/gopiler"
	re "github.com/Vacheprime/gopiler/lexer/regex"
)

// rangeQueue is a slice of character ranges that implements heap.Interface.
//
// Used solely for building the equivalence ranges.
type rangeQueue []re.CharacterRange

func (rq rangeQueue) Len() int           { return len(rq) }
func (rq rangeQueue) Less(i, h int) bool { return rq[i].Start < rq[h].Start }
func (rq rangeQueue) Swap(i, j int)      { rq[i], rq[j] = rq[j], rq[i] }

func (rq *rangeQueue) Push(cr any) {
	*rq = append(*rq, cr.(re.CharacterRange))
}

func (rq *rangeQueue) Pop() any {
	length := len(*rq)
	last := (*rq)[length-1]
	*rq = (*rq)[0 : length-1]
	return last
}

// ownershipClass is a helper type used to associate an ownership bitset with a class ID.
type ownershipClass struct {
	Bs      gp.BitSet
	ClassId int
}

// IdCharacterRange associates a character range with a class ID.
type IdCharacterRange struct {
	re.CharacterRange
	ClassId int
}

// SymbolClassifier is used to organize the unicode range into distinct classes
// for creating an NFA / DFA table.
type SymbolClassifier struct {
	EquivalenceClasses []IdCharacterRange
	SymToIds           map[re.Symbol][]int
}

/* Total returns the total number of character classes.*/
func (sc SymbolClassifier) Total() int {
	maxId := -1
	for _, eq := range sc.EquivalenceClasses {
		if eq.ClassId > maxId {
			maxId = eq.ClassId
		}
	}
	return maxId + 1
}

// GetClassIdsFromSymbol returns the class IDs that fully match the given symbol.
func (sc SymbolClassifier) GetClassIdsFromSymbol(sym re.Symbol) []int {
	return sc.SymToIds[sym]
}

// Classify takes a rune and returns its class ID.
func (sc SymbolClassifier) Classify(r rune) (classId int) {
	for _, class := range sc.EquivalenceClasses {
		if class.CharacterRange.Matches(r) {
			return class.ClassId
		}
	}
	return -1
}

// GetEquiRangesFromClassId returns the ranges of characters that map to the given class ID.
func (sc SymbolClassifier) GetEquiRangesFromClassId(classId int) (equiRanges []re.CharacterRange) {
	for _, idRange := range sc.EquivalenceClasses {
		if idRange.ClassId == classId {
			equiRanges = append(equiRanges, idRange.CharacterRange)
		}
	}
	return equiRanges
}

// BuildClassifier creates a classifier based on symbol occurrences.
func BuildClassifier(occurrences SymbolOccurrences) (SymbolClassifier, error) {
	classifier := SymbolClassifier{
		SymToIds: map[re.Symbol][]int{},
	}
	allRanges := []re.CharacterRange{}
	for _, sym := range occurrences {
		allRanges = append(allRanges, sym.CharSet.Ranges...)
	}
	re.SortCharacterRanges(&allRanges)
	equivalenceRanges := getEquivalenceRanges(allRanges)
	ownerships := determineOwnership(equivalenceRanges, occurrences)
	equivalenceClasses, symToIds := determineEquivalenceClasses(ownerships, occurrences)

	classifier.EquivalenceClasses = equivalenceClasses
	classifier.SymToIds = symToIds
	return classifier, nil
}

// getEquivalenceRanges computes the ranges that form the equivalence classes.
//
// Assumes that allRanges is sorted by start value of the ranges.
func getEquivalenceRanges(allRanges []re.CharacterRange) (equivalenceRanges []re.CharacterRange) {
	if len(allRanges) <= 1 {
		return allRanges
	}
	equivalenceRanges = []re.CharacterRange{allRanges[0]}
	var rangeQueue rangeQueue = allRanges[1:]
	heap.Init(&rangeQueue)

	for rangeQueue.Len() != 0 {
		charRange := heap.Pop(&rangeQueue).(re.CharacterRange)
		// Get latest equivalence
		latestEquivalence := &equivalenceRanges[len(equivalenceRanges)-1]

		// Handle Same range
		if *latestEquivalence == charRange {
			continue
		}

		// Handle total overlap at start
		if charRange.Start == latestEquivalence.Start {
			if charRange.End < latestEquivalence.End {
				charRange.Start = charRange.End + 1
				charRange.End = latestEquivalence.End
				latestEquivalence.End = charRange.Start - 1
			} else {
				charRange.Start = latestEquivalence.End + 1
			}
			heap.Push(&rangeQueue, charRange)
			continue
		}

		// Handle total overlap at end
		if charRange.Start <= latestEquivalence.End && charRange.End == latestEquivalence.End {
			latestEquivalence.End = charRange.Start - 1
		}

		// Handle total overlap middle
		if charRange.Start < latestEquivalence.End && charRange.End < latestEquivalence.End {
			endRange := re.CharacterRange{Start: charRange.End + 1, End: latestEquivalence.End}
			latestEquivalence.End = charRange.Start - 1
			heap.Push(&rangeQueue, endRange)
		}

		// Handle regular overlap
		if charRange.Start <= latestEquivalence.End && charRange.End > latestEquivalence.End {
			overlapRange := re.CharacterRange{Start: charRange.Start, End: latestEquivalence.End}
			latestEquivalence.End = overlapRange.Start - 1
			charRange.Start = overlapRange.End + 1
			equivalenceRanges = append(equivalenceRanges, overlapRange)
			heap.Push(&rangeQueue, charRange)
			continue
		}

		equivalenceRanges = append(equivalenceRanges, charRange)
	}
	return equivalenceRanges
}

// determineOwnership maps equivalence ranges to the symbol they match.
func determineOwnership(equiRanges []re.CharacterRange, occurrences SymbolOccurrences) map[re.CharacterRange]gp.BitSet {
	ownershipSets := map[re.CharacterRange]gp.BitSet{}
	MAX_BITS := uint(len(occurrences))
	for _, equiRange := range equiRanges {
		bs := gp.NewBitSet(MAX_BITS)
		for _, sym := range occurrences {
			if !sym.CharSet.RangeIsFullyContained(equiRange) {
				continue
			}
			bs.Set(uint(sym.SymbolId))
		}
		ownershipSets[equiRange] = bs
	}
	return ownershipSets
}

func determineEquivalenceClasses(ownerships map[re.CharacterRange]gp.BitSet, occurrences SymbolOccurrences) (equivalenceClasses []IdCharacterRange, symToIds map[re.Symbol][]int) {
	equivalenceClasses = []IdCharacterRange{}
	symToIds = map[re.Symbol][]int{}
	classIds := []ownershipClass{}
	symSlice := slices.Collect(maps.Values(occurrences))
	if len(ownerships) == 0 {
		return equivalenceClasses, symToIds
	}
	for cr, bs := range ownerships {
		equiClass := IdCharacterRange{CharacterRange: cr}
		existingClassIdx := slices.IndexFunc(classIds, func(oc ownershipClass) bool {
			return oc.Bs.Equals(bs)
		})
		if existingClassIdx != -1 {
			equiClass.ClassId = classIds[existingClassIdx].ClassId
			equivalenceClasses = append(equivalenceClasses, equiClass)
			continue
		}
		oc := ownershipClass{Bs: bs, ClassId: len(classIds)}
		equiClass.ClassId = oc.ClassId
		equivalenceClasses = append(equivalenceClasses, equiClass)
		classIds = append(classIds, oc)

		for _, symId := range bs.GetActiveBitPositions() {
			symIdx := slices.IndexFunc(symSlice, func(sym re.Symbol) bool {
				return sym.SymbolId == int(symId)
			})
			sym := symSlice[symIdx]
			symToIds[sym] = append(symToIds[sym], oc.ClassId)
		}
	}
	return equivalenceClasses, symToIds
}
