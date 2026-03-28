package gopiler

import (
	"fmt"
	"strings"
)

func generateExpressionCode(top_expr *Expression) string {
	var code strings.Builder
	switch top_expr.Type {
	case EXP_TYPE_DIGIT:
		code.WriteString(fmt.Sprintf("PUSH %v\n", top_expr.DigitValue))
	case EXP_TYPE_EXP:
		// Process left and right expressions
		code.WriteString(generateExpressionCode(top_expr.LeftExpr))
		code.WriteString(generateExpressionCode(top_expr.RightExpr))
		// Process operator
		switch top_expr.Operator {
		case '*':
			code.WriteString("MULT\n")
		case '+':
			code.WriteString("ADD\n")
		}
	}
	return code.String()
}

// Generates the target code from the AST
func GenerateCode(top_expr *Expression) string {
	code := generateExpressionCode(top_expr)
	return code + "PRINT\n"
}
