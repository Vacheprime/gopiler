package gopiler

import (
	"errors"
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
