package gopiler

import (
	"errors"
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

func (s *Stack[T]) Push(x T) {
	s.items = append(s.items, x)
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

type Set[T comparable] struct {
	items map[T]struct{}
}

func NewSet[T comparable]() Set[T] {
	return Set[T]{make(map[T]struct{})}
}

func (s *Set[T]) Add(items ...T) {
	for i := range items {
		s.items[items[i]] = struct{}{}
	}
}

func (s *Set[T]) Remove(item T) {
	delete(s.items, item)
}

func (s *Set[T]) Len() int {
	return len(s.items)
}

func (s *Set[T]) Clear() {
	clear(s.items)
}

func (s *Set[T]) Contains(item T) bool {
	_, ok := s.items[item]
	return ok
}

func (s *Set[T]) Items() map[T]struct{} {
	return s.items
}
