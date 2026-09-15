package parser

import (
	"errors"
	"strconv"

	lx "github.com/Vacheprime/gopiler/lexer"
)

var (
	ErrUnexpectedToken error = errors.New("Unexpected token")
)

type Parser struct {
	lexer lx.TokenStream
}

func NewParser(ts lx.TokenStream) Parser {
	return Parser{lexer: ts}
}

func (p *Parser) matchToken(tkType lx.TokenType) (tk *lx.Token, err error) {
	next, err := p.lexer.NextToken()
	if err != nil {
		return nil, err
	}
	if next.TkType != tkType {
		return nil, ErrUnexpectedToken
	}
	return &next, nil
}

func (p *Parser) ParseVarDeclaration() (node *VarDeclaration, err error) {
	declNode, err := p.ParseTypeDeclaration()
	if err != nil {
		return nil, err
	}
	identifier, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	_, err = p.matchToken(lx.ASSIGNMENT)
	if err != nil {
		return nil, err
	}
	value, err := p.ParseValueLiteral()
	if err != nil {
		return nil, err
	}
	return &VarDeclaration{
		TypeDecl:   declNode,
		Identifier: identifier,
		Value:      value,
		astNode: astNode{
			Position: declNode.Position,
		},
	}, nil
}

func (p *Parser) ParseTypeDeclaration() (node *TypeDeclaration, err error) {
	next, err := p.lexer.NextToken()
	if err != nil {
		return nil, err
	}
	var isValid bool
	var dType DataType
	switch next.TkType {
	case lx.DTYPE_FLOAT, lx.DTYPE_INT:
		isValid = true
		dType = DataType(next.TkType)
	default:
		isValid = false
	}
	if !isValid {
		return nil, ErrUnexpectedToken
	}
	return &TypeDeclaration{DataType: dType, astNode: astNode{Position: next.Pos}}, nil
}

func (p *Parser) ParseIdentifier() (node *Identifier, err error) {
	idenTk, err := p.matchToken(lx.IDENTIFIER)
	if err != nil {
		return nil, err
	}
	return &Identifier{
		Name:    idenTk.Repr,
		astNode: astNode{Position: idenTk.Pos},
	}, nil
}

func (p *Parser) ParseValueLiteral() (node ValueLiteral, err error) {
	value, err := p.lexer.PeekToken(0)
	if err != nil {
		return nil, err
	}
	switch value.TkType {
	case lx.INTEGER:
		p.lexer.NextToken() // consume
		intVal, _ := strconv.ParseInt(value.Repr, 10, 64)
		return &Integer{Val: intVal, astNode: astNode{Position: value.Pos}}, nil
	case lx.FLOAT:
		p.lexer.NextToken() // consume
		floatVal, _ := strconv.ParseFloat(value.Repr, 64)
		return &Float{Val: floatVal, astNode: astNode{Position: value.Pos}}, nil
	}
	return nil, ErrUnexpectedToken
}

func (p *Parser) ParseComparison() (node *Comparison, err error) {
	cmpTk, err := p.lexer.PeekToken(0)
	if err != nil {
		return nil, err
	}
	switch cmpTk.TkType {
	case lx.EQUAL, lx.GREATER_THAN, lx.GREATER_EQ_THAN, lx.LESS_THAN, lx.LESS_EQ_THAN:
		p.lexer.NextToken()
		return &Comparison{Cmp: ComparisonType(cmpTk.TkType), astNode: astNode{
			Position: cmpTk.Pos,
		}}, nil
	}
	return nil, ErrUnexpectedToken
}

func (p *Parser) ParsePrimary() (node Primary, err error) {
	next, err := p.lexer.PeekToken(0)
	if err != nil {
		return nil, err
	}
	switch next.TkType {
	case lx.LEFT_PAREN:
		parenExpr, err := p.ParseParenExpr()
		if err != nil {
			return nil, err
		}
		return parenExpr, nil
	case lx.IDENTIFIER:
		after, err := p.lexer.PeekToken(1)
		if err != nil {
			return nil, err
		}
		if after.TkType == lx.LEFT_PAREN {
			funcCall, err := p.ParseFuncCall()
			if err != nil {
				return nil, err
			}
			return funcCall, nil
		} else {
			iden, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			return iden, nil
		}
	default:
		valueLiteral, err := p.ParseValueLiteral()
		if err != nil {
			return nil, err
		}
		return valueLiteral, nil
	}
}

func (p *Parser) ParseParenExpr() (node *ParenExpr, err error) {
	lParen, err := p.matchToken(lx.LEFT_PAREN)
	if err != nil {
		return nil, err
	}
	expr, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}
	_, err = p.matchToken(lx.RIGHT_PAREN)
	if err != nil {
		return nil, err
	}
	return &ParenExpr{
		astNode:    astNode{Position: lParen.Pos},
		Expression: expr,
	}, nil
}

func (p *Parser) ParseExpr() (node *Expr, err error) { return nil, nil }

func (p *Parser) ParseFuncCall() (node *FuncCall, err error) {
	idenTk, err := p.matchToken(lx.IDENTIFIER)
	if err != nil {

	}
}
