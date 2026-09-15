package parser

import lx "github.com/Vacheprime/gopiler/lexer"

type DataType int

const (
	INTEGER_TYPE = DataType(lx.DTYPE_INT)
	FLOAT_TYPE   = DataType(lx.DTYPE_FLOAT)
)

type ComparisonType int

const (
	CMP_EQUAL           = ComparisonType(lx.EQUAL)
	CMP_GREATER_THAN    = ComparisonType(lx.GREATER_THAN)
	CMP_LESS_THAN       = ComparisonType(lx.LESS_THAN)
	CMP_GREATER_EQ_THAN = ComparisonType(lx.GREATER_EQ_THAN)
	CMP_LESS_EQ_THAN    = ComparisonType(lx.LESS_EQ_THAN)
)

type AdditionOperator int

const (
	OP_SUM      = AdditionOperator(lx.SUM)
	OP_SUBTRACT = AdditionOperator(lx.SUBTRACTION)
)

type MultiplicationOperator int

const (
	OP_MULTI  = MultiplicationOperator(lx.MULTIPLICATION)
	OP_DIVIDE = MultiplicationOperator(lx.DIVISION)
)
