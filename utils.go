package gopiler

import "errors"

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
