package semantic

import (
	"fmt"

	"mycompiler/parser/ast"
)

type varInfo struct {
	defined bool
	used    bool
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
	scope[name] = &varInfo{defined: false, used: false}
}

func (a *Analyzer) define(name string) {
	info, ok := a.resolve(name)
	if !ok {
		return
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
			a.analyzeExpr(s.Initializer)
			a.define(s.Name)
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
			a.analyzeExpr(s.Condition)
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
			a.analyzeExpr(s.Condition)
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

func (a *Analyzer) analyzeExpr(expr ast.Expression) {
	if expr == nil {
		a.errorf("nil expression")
		return
	}

	switch e := expr.(type) {
	case *ast.NumberExpression:
	case *ast.StringExpression:
	case *ast.VariableExpression:
		if e.Name == "" {
			a.errorf("variable expression with empty name")
			return
		}
		info, ok := a.resolve(e.Name)
		if !ok {
			a.errorf("variable '%s' is not declared", e.Name)
			return
		}
		a.markUsed(e.Name)
		if !info.defined {
			a.errorf("variable '%s' is used before it is initialized", e.Name)
		}

	case *ast.AssignExpression:
		if e.Name == "" {
			a.errorf("assignment with empty target name")
			return
		}
		if e.Value == nil {
			a.errorf("assignment to '%s' with nil value", e.Name)
			return
		}

		a.analyzeExpr(e.Value)

		_, ok := a.resolve(e.Name)
		if !ok {
			a.errorf("cannot assign to undeclared variable '%s'", e.Name)
		} else {
			a.define(e.Name)
		}

	case *ast.UnaryExpression:
		if e.Right == nil {
			a.errorf("unary expression with nil operand")
			return
		}
		a.analyzeExpr(e.Right)

	case *ast.BinaryExpression:
		if e.Left == nil || e.Right == nil {
			a.errorf("binary expression with nil operand")
			return
		}
		a.analyzeExpr(e.Left)
		a.analyzeExpr(e.Right)

	default:
		a.errorf("unknown expression node type: %T", expr)
	}
}
