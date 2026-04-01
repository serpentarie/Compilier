package semantic

import (
	"fmt"

	"mycompiler/lexer"
	"mycompiler/parser/ast"
)

type valueType int

const (
	typeUnknown valueType = iota
	typeNumber
	typeString
	typeBool
)

func (t valueType) String() string {
	switch t {
	case typeNumber:
		return "number"
	case typeString:
		return "string"
	case typeBool:
		return "bool"
	default:
		return "unknown"
	}
}

type varInfo struct {
	defined bool
	used    bool
	typ     valueType
}

type Analyzer struct {
	scopes   []map[string]*varInfo
	errors   []error
	warnings []error
}

func NewAnalyzer() *Analyzer {
	a := &Analyzer{}
	a.beginScope()
	return a
}

func (a *Analyzer) Analyze(statements []ast.Statement) []error {
	for _, stmt := range statements {
		a.analyzeStmt(stmt)
	}
	a.endScope()
	return append(a.errors, a.warnings...)
}

func (a *Analyzer) HasErrors() bool { return len(a.errors) > 0 }

func (a *Analyzer) Errors() []error { return append(a.errors, a.warnings...) }

func (a *Analyzer) errorf(format string, args ...any) {
	a.errors = append(a.errors, fmt.Errorf(format, args...))
}

func (a *Analyzer) warnf(format string, args ...any) {
	a.warnings = append(a.warnings, fmt.Errorf("[WARNING] "+format, args...))
}

func (a *Analyzer) beginScope() {
	a.scopes = append(a.scopes, map[string]*varInfo{})
}

func (a *Analyzer) endScope() {
	if len(a.scopes) == 0 {
		return
	}
	scope := a.scopes[len(a.scopes)-1]
	for name, info := range scope {
		if !info.used {
			a.warnf("variable '%s' is declared but never used", name)
		}
	}
	a.scopes = a.scopes[:len(a.scopes)-1]
}

func (a *Analyzer) currentScope() map[string]*varInfo {
	if len(a.scopes) == 0 {
		a.beginScope()
	}
	return a.scopes[len(a.scopes)-1]
}

func (a *Analyzer) resolve(name string) (*varInfo, bool) {
	for i := len(a.scopes) - 1; i >= 0; i-- {
		if info, ok := a.scopes[i][name]; ok {
			return info, true
		}
	}
	return nil, false
}

func (a *Analyzer) declare(name string) {
	scope := a.currentScope()
	if _, exists := scope[name]; exists {
		a.errorf("redeclared variable '%s' in the same scope", name)
		return
	}
	scope[name] = &varInfo{defined: false, used: false, typ: typeUnknown}
}

func (a *Analyzer) define(name string, typ valueType) {
	info, ok := a.resolve(name)
	if !ok {
		return
	}
	if info.typ == typeUnknown {
		info.typ = typ
	}
	info.defined = true
}

func (a *Analyzer) markUsed(name string) {
	info, ok := a.resolve(name)
	if !ok {
		return
	}
	info.used = true
}

func (a *Analyzer) analyzeStmt(stmt ast.Statement) {
	if stmt == nil {
		a.errorf("nil statement")
		return
	}

	switch s := stmt.(type) {
	case *ast.VarStatement:
		if s.Name == "" {
			a.errorf("variable declaration with empty name")
			return
		}
		a.declare(s.Name)
		if s.Initializer != nil {
			typ := a.analyzeExpr(s.Initializer)
			a.define(s.Name, typ)
		}

	case *ast.ExpressionStatement:
		if s.Expression == nil {
			a.errorf("expression statement with nil expression")
			return
		}
		a.analyzeExpr(s.Expression)

	case *ast.PrintStatement:
		if s.Expression == nil {
			a.errorf("print statement with nil expression")
			return
		}
		a.analyzeExpr(s.Expression)

	case *ast.BlockStatement:
		a.beginScope()
		for _, inner := range s.Statements {
			a.analyzeStmt(inner)
		}
		a.endScope()

	case *ast.IfStatement:
		if s.Condition == nil {
			a.errorf("if statement with nil condition")
		} else {
			conditionType := a.analyzeExpr(s.Condition)
			if conditionType != typeBool && conditionType != typeUnknown {
				a.errorf("if condition must be bool, got %s", conditionType)
			}
		}
		if s.ThenBranch == nil {
			a.errorf("if statement with nil then-branch")
		} else {
			a.analyzeStmt(s.ThenBranch)
		}
		if s.ElseBranch != nil {
			a.analyzeStmt(s.ElseBranch)
		}

	case *ast.WhileStatement:
		if s.Condition == nil {
			a.errorf("while statement with nil condition")
		} else {
			conditionType := a.analyzeExpr(s.Condition)
			if conditionType != typeBool && conditionType != typeUnknown {
				a.errorf("while condition must be bool, got %s", conditionType)
			}
		}
		if s.Body == nil {
			a.errorf("while statement with nil body")
			return
		}
		a.analyzeStmt(s.Body)

	default:
		a.errorf("unknown statement node type: %T", stmt)
	}
}

func (a *Analyzer) analyzeExpr(expr ast.Expression) valueType {
	if expr == nil {
		a.errorf("nil expression")
		return typeUnknown
	}

	switch e := expr.(type) {
	case *ast.NumberExpression:
		return typeNumber
	case *ast.StringExpression:
		return typeString
	case *ast.BoolExpression:
		return typeBool
	case *ast.VariableExpression:
		if e.Name == "" {
			a.errorf("variable expression with empty name")
			return typeUnknown
		}
		info, ok := a.resolve(e.Name)
		if !ok {
			a.errorf("variable '%s' is not declared", e.Name)
			return typeUnknown
		}
		a.markUsed(e.Name)
		if !info.defined {
			a.errorf("variable '%s' is used before it is initialized", e.Name)
			return typeUnknown
		}
		return info.typ

	case *ast.AssignExpression:
		if e.Name == "" {
			a.errorf("assignment with empty target name")
			return typeUnknown
		}
		if e.Value == nil {
			a.errorf("assignment to '%s' with nil value", e.Name)
			return typeUnknown
		}

		valueType := a.analyzeExpr(e.Value)

		info, ok := a.resolve(e.Name)
		if !ok {
			a.errorf("cannot assign to undeclared variable '%s'", e.Name)
		} else {
			if info.typ == typeUnknown {
				info.typ = valueType
			} else if valueType != typeUnknown && info.typ != valueType {
				a.errorf("type mismatch in assignment to '%s': expected %s, got %s", e.Name, info.typ, valueType)
			}
			a.define(e.Name, info.typ)
		}
		return valueType

	case *ast.UnaryExpression:
		if e.Right == nil {
			a.errorf("unary expression with nil operand")
			return typeUnknown
		}
		rightType := a.analyzeExpr(e.Right)
		switch e.Operator {
		case lexer.MINUS:
			if rightType != typeNumber && rightType != typeUnknown {
				a.errorf("operator '-' requires number operand, got %s", rightType)
				return typeUnknown
			}
			return typeNumber
		case lexer.EXCL:
			if rightType != typeBool && rightType != typeUnknown {
				a.errorf("operator '!' requires bool operand, got %s", rightType)
				return typeUnknown
			}
			return typeBool
		default:
			a.errorf("unknown unary operator: %s", e.Operator)
			return typeUnknown
		}

	case *ast.BinaryExpression:
		if e.Left == nil || e.Right == nil {
			a.errorf("binary expression with nil operand")
			return typeUnknown
		}
		leftType := a.analyzeExpr(e.Left)
		rightType := a.analyzeExpr(e.Right)

		switch e.Operator {
		case lexer.PLUS:
			if leftType == typeNumber && rightType == typeNumber {
				return typeNumber
			}
			if leftType == typeString && rightType == typeString {
				return typeString
			}
			a.errorf("operator '+' supports only number+number or string+string, got %s+%s", leftType, rightType)
			return typeUnknown

		case lexer.MINUS, lexer.STAR, lexer.SLASH:
			if leftType == typeNumber && rightType == typeNumber {
				return typeNumber
			}
			a.errorf("operator '%s' supports only number operands, got %s and %s", e.Operator, leftType, rightType)
			return typeUnknown

		case lexer.LT, lexer.LTEQ, lexer.GT, lexer.GTEQ:
			if leftType == typeNumber && rightType == typeNumber {
				return typeBool
			}
			a.errorf("comparison operator '%s' supports only numbers, got %s and %s", e.Operator, leftType, rightType)
			return typeUnknown

		case lexer.EQEQ, lexer.NEQ:
			if leftType == typeUnknown || rightType == typeUnknown {
				return typeBool
			}
			if leftType == rightType {
				return typeBool
			}
			a.errorf("equality operator '%s' requires same operand types, got %s and %s", e.Operator, leftType, rightType)
			return typeUnknown

		case lexer.AND, lexer.OR:
			if leftType == typeBool && rightType == typeBool {
				return typeBool
			}
			a.errorf("logical operator '%s' supports only bool operands, got %s and %s", e.Operator, leftType, rightType)
			return typeUnknown

		default:
			a.errorf("unknown binary operator: %s", e.Operator)
			return typeUnknown
		}

	default:
		a.errorf("unknown expression node type: %T", expr)
		return typeUnknown
	}
}
