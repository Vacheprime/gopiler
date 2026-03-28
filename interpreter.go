package gopiler

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

type Stack struct {
	numbers []int64
}

func NewStack() Stack {
	return Stack{numbers: make([]int64, 0)}
}

func (s *Stack) Push(x int64) {
	s.numbers = append(s.numbers, x)
}

func (s *Stack) Pop() int64 {
	// Ignore
	if len(s.numbers) == 0 {
		return -1 // Indicate error, demo purposes
	}
	n := len(s.numbers)
	value := s.numbers[n-1]
	s.numbers = s.numbers[:n-1]
	return value
}

func InterpretCode(code string) {
	reader := strings.NewReader(code)
	scanner := bufio.NewScanner(reader)
	stack := NewStack()
	for {
		// Advance to next line
		if !scanner.Scan() {
			break
		}
		// Read line
		line := scanner.Text()

		// Print instruction
		if line == "PRINT" {
			fmt.Println(stack.Pop())
			continue
		}

		if strings.HasPrefix(line, "PUSH") {
			instructionParts := strings.Split(line, " ")
			number, _ := strconv.ParseInt(instructionParts[1], 10, 64)
			stack.Push(number)
			continue
		}

		num1 := stack.Pop()
		num2 := stack.Pop()
		// Check if MULT or ADD
		switch line {
		case "MULT":
			stack.Push(num1 * num2)
		case "ADD":
			stack.Push(num1 + num2)
		}
	}
}
