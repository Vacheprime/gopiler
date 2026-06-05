package gopiler

import (
	"errors"
	"math"
	"math/bits"
	"slices"
)

var (
	ErrEmptyStack = errors.New("empty stack")
)

type Stack[T any] struct {
	items []T
}

func NewStack[T any]() Stack[T] {
	return Stack[T]{items: make([]T, 0)}
}

func (s *Stack[T]) Peak() (T, error) {
	if len(s.items) == 0 {
		var zero T
		return zero, ErrEmptyStack
	}
	return s.items[len(s.items)-1], nil
}

func (s *Stack[T]) Push(items ...T) {
	s.items = append(s.items, items...)
}

func (s *Stack[T]) Pop() (T, error) {
	// Ignore
	if len(s.items) == 0 {
		var zero T
		return zero, ErrEmptyStack
	}
	n := len(s.items)
	value := s.items[n-1]
	s.items = s.items[:n-1]
	return value, nil
}

func (s *Stack[T]) Len() int {
	return len(s.items)
}

type Set[T comparable] struct {
	items []T
}

func NewSet[T comparable](items ...T) Set[T] {
	s := Set[T]{make([]T, 0, len(items)*2)}
	s.Add(items...)
	return s
}

func (s *Set[T]) Add(items ...T) {
	for i := range items {
		c := items[i]
		if !s.Contains(c) {
			s.items = append(s.items, c)
		}
	}
}

func (s *Set[T]) Remove(item T) {
	for i, v := range s.items {
		if v == item {
			s.items = append(s.items[:i], s.items[i+1:]...)
			break
		}
	}
}

func (s *Set[T]) Len() int {
	return len(s.items)
}

func (s *Set[T]) Clear() {
	clear(s.items)
}

func (s *Set[T]) Contains(item T) bool {
	return slices.Contains(s.items, item)
}

func (s *Set[T]) Items() []T {
	return s.items
}

type BitSet struct {
	Bits []uint64
}

func NewBitSet(minBitNbr uint) BitSet {
	length := int(math.Ceil(float64(minBitNbr) / 64))
	return BitSet{make([]uint64, length)}
}

func (bs *BitSet) Set(bitPos uint) {
	intPos := uint(math.Floor(float64(bitPos) / 64))
	realBitPos := bitPos % 64
	bs.Bits[intPos] |= 1 << realBitPos
}

func (bs *BitSet) Unset(bitPos ...uint) {
	for _, num := range bitPos {
		intPos := uint(math.Floor(float64(num) / 64))
		realBitPos := num % 64
		bs.Bits[intPos] &^= 1 << realBitPos
	}

}

func (bs *BitSet) GetActiveBitPositions() []uint {
	var positions []uint
	for idx, integer := range bs.Bits {
		for integer != 0 {
			bit := bits.TrailingZeros64(integer)
			pos := uint(idx*64 + bit)
			integer &= integer - 1
			positions = append(positions, pos)
		}
	}
	return positions
}

func (bs *BitSet) Or(other BitSet) {
	for i := range bs.Bits {
		bs.Bits[i] |= other.Bits[i]
	}
}

func (bs *BitSet) And(other BitSet) {
	for i := range bs.Bits {
		bs.Bits[i] &= other.Bits[i]
	}
}

func (bs *BitSet) CreateCopy() BitSet {
	size := uint(len(bs.Bits) * 64)
	newBs := NewBitSet(size)
	for i := range newBs.Bits {
		newBs.Bits[i] = bs.Bits[i]
	}
	return newBs
}

func (bs *BitSet) Equals(other BitSet) bool {
	if len(bs.Bits) != len(other.Bits) {
		return false
	}
	for i := range bs.Bits {
		if bs.Bits[i] != other.Bits[i] {
			return false
		}
	}
	return true
}

func (bs *BitSet) Overlaps(other BitSet) bool {
	for i := range bs.Bits {
		if bs.Bits[i]&other.Bits[i] != 0 {
			return true
		}
	}
	return false
}

func (bs *BitSet) IsZero() bool {
	for _, num := range bs.Bits {
		if num != 0 {
			return false
		}
	}
	return true
}

func (bs *BitSet) IsNil() bool {
	return bs.Bits == nil
}
