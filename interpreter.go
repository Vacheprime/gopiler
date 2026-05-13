package gopiler

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

func InterpretCode(code string) {
	reader := strings.NewReader(code)
	scanner := bufio.NewScanner(reader)
	stack := NewStack[int64]()
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

		num1, err := stack.Pop()
		if err == ErrEmptyStack {
			fmt.Println("Unexpected end of stack.")
			break
		}
		num2, err := stack.Pop()
		if err == ErrEmptyStack {
			fmt.Println("Unexpected end of stack.")
			break
		}
		// Check if MULT or ADD
		switch line {
		case "MULT":
			stack.Push(num1 * num2)
		case "ADD":
			stack.Push(num1 + num2)
		}
	}
}
