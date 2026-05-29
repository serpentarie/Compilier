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
	typeFunction
	typeArray
)

func (t valueType) String() string {
	switch t {
	case typeNumber:
		return "number"
	case typeString:
		return "string"
	case typeBool:
		return "bool"
	case typeFunction:
		return "function"
	case typeArray:
		return "array"
	default:
		return "unknown"
	}
}

type typeInfo struct {
	kind valueType
	elem valueType
}

func simpleType(kind valueType) typeInfo {
	return typeInfo{kind: kind, elem: typeUnknown}
}

func arrayType(elem valueType) typeInfo {
	return typeInfo{kind: typeArray, elem: elem}
}

func (t typeInfo) String() string {
	if t.kind == typeArray {
		if t.elem == typeUnknown {
			return "array<unknown>"
		}
		return fmt.Sprintf("array<%s>", t.elem)
	}
	return t.kind.String()
}

type varInfo struct {
	defined bool
	used    bool
	typ     typeInfo
}

type Analyzer struct {
	scopes        []map[string]*varInfo
	errors        []error
	warnings      []error
	functionDepth int
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
	scope[name] = &varInfo{defined: false, used: false, typ: simpleType(typeUnknown)}
}

func (a *Analyzer) define(name string, typ typeInfo) {
	info, ok := a.resolve(name)
	if !ok {
		return
	}
	if info.typ.kind == typeUnknown {
		info.typ = typ
		info.defined = true
		return
	}
	if info.typ.kind == typeArray && info.typ.elem == typeUnknown && typ.kind == typeArray && typ.elem != typeUnknown {
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
			if conditionType.kind != typeBool && conditionType.kind != typeUnknown {
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
			if conditionType.kind != typeBool && conditionType.kind != typeUnknown {
				a.errorf("while condition must be bool, got %s", conditionType)
			}
		}
		if s.Body == nil {
			a.errorf("while statement with nil body")
			return
		}
		a.analyzeStmt(s.Body)

	case *ast.FunctionStatement:
		if s.Name == "" {
			a.errorf("function declaration with empty name")
			return
		}

		a.declare(s.Name)
		a.define(s.Name, simpleType(typeFunction))

		a.beginScope()
		seenParams := map[string]struct{}{}
		for _, p := range s.Params {
			if p == "" {
				a.errorf("function '%s' has empty parameter name", s.Name)
				continue
			}
			if _, exists := seenParams[p]; exists {
				a.errorf("function '%s' has duplicate parameter '%s'", s.Name, p)
				continue
			}
			seenParams[p] = struct{}{}
			a.declare(p)
			a.define(p, simpleType(typeUnknown))
		}

		a.functionDepth++
		for _, bodyStmt := range s.Body {
			a.analyzeStmt(bodyStmt)
		}
		a.functionDepth--
		a.endScope()

	case *ast.ReturnStatement:
		if a.functionDepth == 0 {
			a.errorf("'return' is only allowed inside function body")
			return
		}
		if s.Value != nil {
			a.analyzeExpr(s.Value)
		}

	default:
		a.errorf("unknown statement node type: %T", stmt)
	}
}

func (a *Analyzer) analyzeExpr(expr ast.Expression) typeInfo {
	if expr == nil {
		a.errorf("nil expression")
		return simpleType(typeUnknown)
	}

	switch e := expr.(type) {
	case *ast.NumberExpression:
		return simpleType(typeNumber)
	case *ast.StringExpression:
		return simpleType(typeString)
	case *ast.BoolExpression:
		return simpleType(typeBool)
	case *ast.VariableExpression:
		if e.Name == "" {
			a.errorf("variable expression with empty name")
			return simpleType(typeUnknown)
		}
		info, ok := a.resolve(e.Name)
		if !ok {
			a.errorf("variable '%s' is not declared", e.Name)
			return simpleType(typeUnknown)
		}
		a.markUsed(e.Name)
		if !info.defined {
			a.errorf("variable '%s' is used before it is initialized", e.Name)
			return simpleType(typeUnknown)
		}
		return info.typ

	case *ast.AssignExpression:
		if e.Name == "" {
			a.errorf("assignment with empty target name")
			return simpleType(typeUnknown)
		}
		if e.Value == nil {
			a.errorf("assignment to '%s' with nil value", e.Name)
			return simpleType(typeUnknown)
		}

		valueType := a.analyzeExpr(e.Value)

		info, ok := a.resolve(e.Name)
		if !ok {
			a.errorf("cannot assign to undeclared variable '%s'", e.Name)
		} else {
			if info.typ.kind == typeUnknown {
				info.typ = valueType
			} else if valueType.kind != typeUnknown && !sameType(info.typ, valueType) {
				a.errorf("type mismatch in assignment to '%s': expected %s, got %s", e.Name, info.typ, valueType)
			}
			a.define(e.Name, info.typ)
		}
		return valueType

	case *ast.CallExpression:
		if e.Callee == nil {
			a.errorf("call expression with nil callee")
			return simpleType(typeUnknown)
		}
		calleeType := a.analyzeExpr(e.Callee)
		for _, arg := range e.Arguments {
			a.analyzeExpr(arg)
		}
		if calleeType.kind != typeFunction && calleeType.kind != typeUnknown {
			a.errorf("attempt to call non-function value of type %s", calleeType)
		}
		return simpleType(typeUnknown)

	case *ast.UnaryExpression:
		if e.Right == nil {
			a.errorf("unary expression with nil operand")
			return simpleType(typeUnknown)
		}
		rightType := a.analyzeExpr(e.Right)
		switch e.Operator {
		case lexer.MINUS:
			if rightType.kind != typeNumber && rightType.kind != typeUnknown {
				a.errorf("operator '-' requires number operand, got %s", rightType)
				return simpleType(typeUnknown)
			}
			return simpleType(typeNumber)
		case lexer.EXCL:
			if rightType.kind != typeBool && rightType.kind != typeUnknown {
				a.errorf("operator '!' requires bool operand, got %s", rightType)
				return simpleType(typeUnknown)
			}
			return simpleType(typeBool)
		default:
			a.errorf("unknown unary operator: %s", e.Operator)
			return simpleType(typeUnknown)
		}

	case *ast.BinaryExpression:
		if e.Left == nil || e.Right == nil {
			a.errorf("binary expression with nil operand")
			return simpleType(typeUnknown)
		}
		leftType := a.analyzeExpr(e.Left)
		rightType := a.analyzeExpr(e.Right)

		switch e.Operator {
		case lexer.PLUS:
			if leftType.kind == typeNumber && rightType.kind == typeNumber {
				return simpleType(typeNumber)
			}
			if leftType.kind == typeString && rightType.kind == typeString {
				return simpleType(typeString)
			}
			a.errorf("operator '+' supports only number+number or string+string, got %s+%s", leftType, rightType)
			return simpleType(typeUnknown)

		case lexer.MINUS, lexer.STAR, lexer.SLASH:
			if leftType.kind == typeNumber && rightType.kind == typeNumber {
				return simpleType(typeNumber)
			}
			a.errorf("operator '%s' supports only number operands, got %s and %s", e.Operator, leftType, rightType)
			return simpleType(typeUnknown)

		case lexer.LT, lexer.LTEQ, lexer.GT, lexer.GTEQ:
			if leftType.kind == typeNumber && rightType.kind == typeNumber {
				return simpleType(typeBool)
			}
			a.errorf("comparison operator '%s' supports only numbers, got %s and %s", e.Operator, leftType, rightType)
			return simpleType(typeUnknown)

		case lexer.EQEQ, lexer.NEQ:
			if leftType.kind == typeUnknown || rightType.kind == typeUnknown {
				return simpleType(typeBool)
			}
			if sameType(leftType, rightType) {
				return simpleType(typeBool)
			}
			a.errorf("equality operator '%s' requires same operand types, got %s and %s", e.Operator, leftType, rightType)
			return simpleType(typeUnknown)

		case lexer.AND, lexer.OR:
			if leftType.kind == typeBool && rightType.kind == typeBool {
				return simpleType(typeBool)
			}
			a.errorf("logical operator '%s' supports only bool operands, got %s and %s", e.Operator, leftType, rightType)
			return simpleType(typeUnknown)

		default:
			a.errorf("unknown binary operator: %s", e.Operator)
			return simpleType(typeUnknown)
		}

	case *ast.ArrayExpression:
		elemType := simpleType(typeUnknown)
		for _, el := range e.Elements {
			if el == nil {
				a.errorf("array literal contains nil element")
				continue
			}
			curr := a.analyzeExpr(el)
			if elemType.kind == typeUnknown && curr.kind != typeUnknown {
				elemType = curr
				continue
			}
			if curr.kind != typeUnknown && !sameType(elemType, curr) {
				a.errorf("array literal elements must be of same type, got %s and %s", elemType, curr)
			}
		}
		return arrayType(elemType.kind)

	case *ast.IndexExpression:
		if e.Target == nil || e.Index == nil {
			a.errorf("index expression with nil target or index")
			return simpleType(typeUnknown)
		}
		arrType := a.analyzeExpr(e.Target)
		idxType := a.analyzeExpr(e.Index)
		if idxType.kind != typeNumber && idxType.kind != typeUnknown {
			a.errorf("array index must be number, got %s", idxType)
		}
		if arrType.kind == typeArray {
			return simpleType(arrType.elem)
		}
		if arrType.kind != typeUnknown {
			a.errorf("indexing requires array, got %s", arrType)
		}
		return simpleType(typeUnknown)

	case *ast.IndexAssignExpression:
		if e.Target == nil || e.Index == nil || e.Value == nil {
			a.errorf("index assignment with nil target, index, or value")
			return simpleType(typeUnknown)
		}
		arrType := a.analyzeExpr(e.Target)
		idxType := a.analyzeExpr(e.Index)
		valueType := a.analyzeExpr(e.Value)
		if idxType.kind != typeNumber && idxType.kind != typeUnknown {
			a.errorf("array index must be number, got %s", idxType)
		}
		if arrType.kind == typeArray {
			if arrType.elem != typeUnknown && valueType.kind != typeUnknown && arrType.elem != valueType.kind {
				a.errorf("type mismatch in index assignment: expected %s, got %s", simpleType(arrType.elem), valueType)
			}
			return simpleType(arrType.elem)
		}
		if arrType.kind != typeUnknown {
			a.errorf("index assignment requires array, got %s", arrType)
		}
		return simpleType(typeUnknown)

	default:
		a.errorf("unknown expression node type: %T", expr)
		return simpleType(typeUnknown)
	}
}

func sameType(a, b typeInfo) bool {
	if a.kind != b.kind {
		return false
	}
	if a.kind == typeArray {
		if a.elem == typeUnknown || b.elem == typeUnknown {
			return true
		}
		return a.elem == b.elem
	}
	return true
}
