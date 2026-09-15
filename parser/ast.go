package parser

import lx "github.com/Vacheprime/gopiler/lexer"

type Node interface {
	Pos() lx.Position
}

type astNode struct {
	Position lx.Position
}

func (a *astNode) Pos() lx.Position { return a.Position }

type Primary interface {
	Node
	primary()
}

type ValueLiteral interface {
	Node
	Primary
	valueLiteral()
}

type Identifier struct {
	astNode
	Name string
}

func (i *Identifier) primary() {}

type Integer struct {
	astNode
	Val int64
}

func (i *Integer) valueLiteral() {}

type Float struct {
	astNode
	Val float64
}

func (f *Float) valueLiteral() {}

type TypeDeclaration struct {
	astNode
	DataType DataType
}

type VarDeclaration struct {
	astNode
	TypeDecl   *TypeDeclaration
	Identifier *Identifier
	Value      ValueLiteral
}

type Comparison struct {
	astNode
	Cmp ComparisonType
}

type Expr struct {
	astNode
	LeftOperand  *Factor
	Operator     AdditionOperator
	RightOperand *Factor
}

func (e *Expr) primary() {}

type Factor struct {
	astNode
	LeftOperand  *Power
	Operator     MultiplicationOperator
	RightOperand *Power
}

type Power struct {
	LeftOperand  Primary
	RightOperand *Power
}

type ParenExpr struct {
	astNode
	Expression *Expr
}

func (p *ParenExpr) primary() {}

type FuncCall struct {
	astNode
	Identifier *Identifier
	Args       *ArgList
}

func (f *FuncCall) primary() {}

type ArgList struct {
	astNode
	Arguments []*Expr
}
