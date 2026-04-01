package ast

import "mycompiler/lexer"

type NumberExpression struct {
	Value float64
}

func (n *NumberExpression) expressionNode() {}

type StringExpression struct {
	Value string
}

func (s *StringExpression) expressionNode() {}

type BoolExpression struct {
	Value bool
}

func (b *BoolExpression) expressionNode() {}

type VariableExpression struct {
	Name string
}

func (v *VariableExpression) expressionNode() {}

type BinaryExpression struct {
	Left     Expression
	Operator lexer.TokenType
	Right    Expression
}

func (b *BinaryExpression) expressionNode() {}

type UnaryExpression struct {
	Operator lexer.TokenType
	Right    Expression
}

func (u *UnaryExpression) expressionNode() {}

type AssignExpression struct {
	Name  string
	Value Expression
}

func (a *AssignExpression) expressionNode() {}
