package semantic

import (
	"mycompiler/lexer"
	"mycompiler/parser/ast"
)

type constValue struct {
	typ valueType
	num float64
	str string
	b   bool
	ok  bool
}

func Optimize(statements []ast.Statement) []ast.Statement {
	out := make([]ast.Statement, 0, len(statements))
	for _, stmt := range statements {
		optimized := optimizeStmt(stmt)
		if optimized != nil {
			out = append(out, optimized)
		}
	}
	return out
}

func optimizeStmt(stmt ast.Statement) ast.Statement {
	if stmt == nil {
		return nil
	}

	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		s.Expression = optimizeExpr(s.Expression)
		return s

	case *ast.PrintStatement:
		s.Expression = optimizeExpr(s.Expression)
		return s

	case *ast.VarStatement:
		if s.Initializer != nil {
			s.Initializer = optimizeExpr(s.Initializer)
		}
		return s

	case *ast.BlockStatement:
		optimized := make([]ast.Statement, 0, len(s.Statements))
		for _, inner := range s.Statements {
			stmt := optimizeStmt(inner)
			if stmt != nil {
				optimized = append(optimized, stmt)
			}
		}
		s.Statements = optimized
		return s

	case *ast.IfStatement:
		s.Condition = optimizeExpr(s.Condition)
		cond := evalConst(s.Condition)
		if cond.ok && cond.typ == typeBool {
			if cond.b {
				return optimizeStmt(s.ThenBranch)
			}
			if s.ElseBranch != nil {
				return optimizeStmt(s.ElseBranch)
			}
			return nil
		}
		s.ThenBranch = optimizeStmt(s.ThenBranch)
		if s.ElseBranch != nil {
			s.ElseBranch = optimizeStmt(s.ElseBranch)
		}
		return s

	case *ast.WhileStatement:
		s.Condition = optimizeExpr(s.Condition)
		cond := evalConst(s.Condition)
		if cond.ok && cond.typ == typeBool && !cond.b {
			return nil
		}
		s.Body = optimizeStmt(s.Body)
		return s

	case *ast.FunctionStatement:
		optimized := make([]ast.Statement, 0, len(s.Body))
		for _, inner := range s.Body {
			stmt := optimizeStmt(inner)
			if stmt != nil {
				optimized = append(optimized, stmt)
			}
		}
		s.Body = optimized
		return s

	case *ast.ReturnStatement:
		if s.Value != nil {
			s.Value = optimizeExpr(s.Value)
		}
		return s

	default:
		return stmt
	}
}

func optimizeExpr(expr ast.Expression) ast.Expression {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.UnaryExpression:
		e.Right = optimizeExpr(e.Right)
		if c := evalConst(e); c.ok {
			return constToExpr(c)
		}
		return e

	case *ast.BinaryExpression:
		e.Left = optimizeExpr(e.Left)
		e.Right = optimizeExpr(e.Right)
		if c := evalConst(e); c.ok {
			return constToExpr(c)
		}
		return e

	case *ast.AssignExpression:
		e.Value = optimizeExpr(e.Value)
		return e

	case *ast.CallExpression:
		e.Callee = optimizeExpr(e.Callee)
		for i, arg := range e.Arguments {
			e.Arguments[i] = optimizeExpr(arg)
		}
		return e

	default:
		return expr
	}
}

func evalConst(expr ast.Expression) constValue {
	switch e := expr.(type) {
	case *ast.NumberExpression:
		return constValue{typ: typeNumber, num: e.Value, ok: true}
	case *ast.StringExpression:
		return constValue{typ: typeString, str: e.Value, ok: true}
	case *ast.BoolExpression:
		return constValue{typ: typeBool, b: e.Value, ok: true}

	case *ast.UnaryExpression:
		right := evalConst(e.Right)
		if !right.ok {
			return constValue{}
		}
		switch e.Operator {
		case lexer.MINUS:
			if right.typ == typeNumber {
				return constValue{typ: typeNumber, num: -right.num, ok: true}
			}
		case lexer.EXCL:
			if right.typ == typeBool {
				return constValue{typ: typeBool, b: !right.b, ok: true}
			}
		}
		return constValue{}

	case *ast.BinaryExpression:
		left := evalConst(e.Left)
		right := evalConst(e.Right)
		if !left.ok || !right.ok {
			return constValue{}
		}

		switch e.Operator {
		case lexer.PLUS:
			if left.typ == typeNumber && right.typ == typeNumber {
				return constValue{typ: typeNumber, num: left.num + right.num, ok: true}
			}
			if left.typ == typeString && right.typ == typeString {
				return constValue{typ: typeString, str: left.str + right.str, ok: true}
			}
		case lexer.MINUS:
			if left.typ == typeNumber && right.typ == typeNumber {
				return constValue{typ: typeNumber, num: left.num - right.num, ok: true}
			}
		case lexer.STAR:
			if left.typ == typeNumber && right.typ == typeNumber {
				return constValue{typ: typeNumber, num: left.num * right.num, ok: true}
			}
		case lexer.SLASH:
			if left.typ == typeNumber && right.typ == typeNumber && right.num != 0 {
				return constValue{typ: typeNumber, num: left.num / right.num, ok: true}
			}
		case lexer.LT:
			if left.typ == typeNumber && right.typ == typeNumber {
				return constValue{typ: typeBool, b: left.num < right.num, ok: true}
			}
		case lexer.LTEQ:
			if left.typ == typeNumber && right.typ == typeNumber {
				return constValue{typ: typeBool, b: left.num <= right.num, ok: true}
			}
		case lexer.GT:
			if left.typ == typeNumber && right.typ == typeNumber {
				return constValue{typ: typeBool, b: left.num > right.num, ok: true}
			}
		case lexer.GTEQ:
			if left.typ == typeNumber && right.typ == typeNumber {
				return constValue{typ: typeBool, b: left.num >= right.num, ok: true}
			}
		case lexer.EQEQ:
			if left.typ == right.typ {
				switch left.typ {
				case typeNumber:
					return constValue{typ: typeBool, b: left.num == right.num, ok: true}
				case typeString:
					return constValue{typ: typeBool, b: left.str == right.str, ok: true}
				case typeBool:
					return constValue{typ: typeBool, b: left.b == right.b, ok: true}
				}
			}
		case lexer.NEQ:
			if left.typ == right.typ {
				switch left.typ {
				case typeNumber:
					return constValue{typ: typeBool, b: left.num != right.num, ok: true}
				case typeString:
					return constValue{typ: typeBool, b: left.str != right.str, ok: true}
				case typeBool:
					return constValue{typ: typeBool, b: left.b != right.b, ok: true}
				}
			}
		case lexer.AND:
			if left.typ == typeBool && right.typ == typeBool {
				return constValue{typ: typeBool, b: left.b && right.b, ok: true}
			}
		case lexer.OR:
			if left.typ == typeBool && right.typ == typeBool {
				return constValue{typ: typeBool, b: left.b || right.b, ok: true}
			}
		}
		return constValue{}

	default:
		return constValue{}
	}
}

func constToExpr(c constValue) ast.Expression {
	switch c.typ {
	case typeNumber:
		return &ast.NumberExpression{Value: c.num}
	case typeString:
		return &ast.StringExpression{Value: c.str}
	case typeBool:
		return &ast.BoolExpression{Value: c.b}
	default:
		return nil
	}
}
