package interpreter

import (
	"fmt"

	"mycompiler/lexer"
	"mycompiler/parser/ast"
)

type Interpreter struct {
	globals  *Environment
	env      *Environment
	warnings []error
}

type returnValue struct {
	value Object
}

func (r returnValue) Error() string {
	return "return"
}

type Function struct {
	name    string
	params  []string
	body    []ast.Statement
	closure *Environment
}

func (f *Function) call(i *Interpreter, args []Object) (Object, error) {
	if len(args) != len(f.params) {
		return Object{}, fmt.Errorf("runtime error: function '%s' expects %d arguments, got %d", f.name, len(f.params), len(args))
	}

	env := NewEnvironment(f.closure)
	for idx, name := range f.params {
		env.DefineInitialized(name, args[idx])
	}

	err := i.executeBlock(f.body, env)
	if err != nil {
		if ret, ok := err.(returnValue); ok {
			return ret.value, nil
		}
		return Object{}, err
	}

	return NilObject(), nil
}

func NewInterpreter() *Interpreter {
	globals := NewEnvironment(nil)
	return &Interpreter{
		globals:  globals,
		env:      globals,
		warnings: make([]error, 0),
	}
}

func (i *Interpreter) Warnings() []error {
	return i.warnings
}

func (i *Interpreter) Interpret(statements []ast.Statement) error {
	for _, stmt := range statements {
		if err := i.Execute(stmt); err != nil {
			if _, ok := err.(returnValue); ok {
				return fmt.Errorf("runtime error: 'return' outside function")
			}
			return err
		}
	}

	i.warnings = append(i.warnings, i.globals.CollectUnusedWarnings()...)
	return nil
}

func (i *Interpreter) Execute(stmt ast.Statement) error {
	if stmt == nil {
		return fmt.Errorf("runtime error: nil statement")
	}

	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		_, err := i.evaluate(s.Expression)
		return err

	case *ast.PrintStatement:
		value, err := i.evaluate(s.Expression)
		if err != nil {
			return err
		}
		fmt.Println(value.String())
		return nil

	case *ast.VarStatement:
		if s.Initializer == nil {
			i.env.Define(s.Name)
			return nil
		}
		value, err := i.evaluate(s.Initializer)
		if err != nil {
			return err
		}
		i.env.DefineInitialized(s.Name, value)
		return nil

	case *ast.BlockStatement:
		return i.executeBlock(s.Statements, NewEnvironment(i.env))

	case *ast.IfStatement:
		cond, err := i.evaluate(s.Condition)
		if err != nil {
			return err
		}
		b, ok := asBool(cond)
		if !ok {
			return fmt.Errorf("runtime error: if condition must be bool")
		}
		if b {
			return i.Execute(s.ThenBranch)
		}
		if s.ElseBranch != nil {
			return i.Execute(s.ElseBranch)
		}
		return nil

	case *ast.WhileStatement:
		for {
			cond, err := i.evaluate(s.Condition)
			if err != nil {
				return err
			}
			b, ok := asBool(cond)
			if !ok {
				return fmt.Errorf("runtime error: while condition must be bool")
			}
			if !b {
				break
			}
			if err := i.Execute(s.Body); err != nil {
				return err
			}
		}
		return nil

	case *ast.FunctionStatement:
		fn := &Function{name: s.Name, params: s.Params, body: s.Body, closure: i.env}
		i.env.DefineInitialized(s.Name, FunctionObject(fn))
		return nil

	case *ast.ReturnStatement:
		var value Object
		if s.Value != nil {
			v, err := i.evaluate(s.Value)
			if err != nil {
				return err
			}
			value = v
		} else {
			value = NilObject()
		}
		return returnValue{value: value}

	default:
		return fmt.Errorf("runtime error: unknown statement type %T", stmt)
	}
}

func (i *Interpreter) executeBlock(statements []ast.Statement, env *Environment) error {
	previous := i.env
	i.env = env
	defer func() {
		i.warnings = append(i.warnings, env.CollectUnusedWarnings()...)
		i.env = previous
	}()

	for _, stmt := range statements {
		if err := i.Execute(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) evaluate(expr ast.Expression) (Object, error) {
	if expr == nil {
		return Object{}, fmt.Errorf("runtime error: nil expression")
	}

	switch e := expr.(type) {
	case *ast.NumberExpression:
		return NumberObject(e.Value), nil
	case *ast.StringExpression:
		return StringObject(e.Value), nil
	case *ast.BoolExpression:
		return BoolObject(e.Value), nil

	case *ast.VariableExpression:
		return i.env.Get(e.Name)

	case *ast.AssignExpression:
		value, err := i.evaluate(e.Value)
		if err != nil {
			return Object{}, err
		}
		if err := i.env.Assign(e.Name, value); err != nil {
			return Object{}, err
		}
		return value, nil

	case *ast.CallExpression:
		callee, err := i.evaluate(e.Callee)
		if err != nil {
			return Object{}, err
		}
		fn, ok := asFunction(callee)
		if !ok {
			return Object{}, fmt.Errorf("runtime error: can only call functions")
		}
		args := make([]Object, 0, len(e.Arguments))
		for _, argExpr := range e.Arguments {
			arg, err := i.evaluate(argExpr)
			if err != nil {
				return Object{}, err
			}
			args = append(args, arg)
		}
		return fn.call(i, args)

	case *ast.UnaryExpression:
		right, err := i.evaluate(e.Right)
		if err != nil {
			return Object{}, err
		}
		switch e.Operator {
		case lexer.MINUS:
			v, ok := asNumber(right)
			if !ok {
				return Object{}, fmt.Errorf("runtime error: unary '-' works only with numbers")
			}
			return NumberObject(-v), nil
		case lexer.EXCL:
			v, ok := asBool(right)
			if !ok {
				return Object{}, fmt.Errorf("runtime error: unary '!' works only with bool")
			}
			return BoolObject(!v), nil
		default:
			return Object{}, fmt.Errorf("runtime error: unknown unary operator %s", e.Operator)
		}

	case *ast.BinaryExpression:
		if e.Operator == lexer.OR {
			left, err := i.evaluate(e.Left)
			if err != nil {
				return Object{}, err
			}
			lb, ok := asBool(left)
			if !ok {
				return Object{}, fmt.Errorf("runtime error: operator '||' works only with bool")
			}
			if lb {
				return BoolObject(true), nil
			}
			right, err := i.evaluate(e.Right)
			if err != nil {
				return Object{}, err
			}
			rb, ok := asBool(right)
			if !ok {
				return Object{}, fmt.Errorf("runtime error: operator '||' works only with bool")
			}
			return BoolObject(rb), nil
		}

		if e.Operator == lexer.AND {
			left, err := i.evaluate(e.Left)
			if err != nil {
				return Object{}, err
			}
			lb, ok := asBool(left)
			if !ok {
				return Object{}, fmt.Errorf("runtime error: operator '&&' works only with bool")
			}
			if !lb {
				return BoolObject(false), nil
			}
			right, err := i.evaluate(e.Right)
			if err != nil {
				return Object{}, err
			}
			rb, ok := asBool(right)
			if !ok {
				return Object{}, fmt.Errorf("runtime error: operator '&&' works only with bool")
			}
			return BoolObject(rb), nil
		}

		left, err := i.evaluate(e.Left)
		if err != nil {
			return Object{}, err
		}
		right, err := i.evaluate(e.Right)
		if err != nil {
			return Object{}, err
		}

		switch e.Operator {
		case lexer.PLUS:
			if l, lok := asNumber(left); lok {
				if r, rok := asNumber(right); rok {
					return NumberObject(l + r), nil
				}
			}
			if l, lok := asString(left); lok {
				if r, rok := asString(right); rok {
					return StringObject(l + r), nil
				}
			}
			return Object{}, fmt.Errorf("runtime error: '+' supports number+number or string+string")

		case lexer.MINUS:
			l, lok := asNumber(left)
			r, rok := asNumber(right)
			if !lok || !rok {
				return Object{}, fmt.Errorf("runtime error: '-' supports only numbers")
			}
			return NumberObject(l - r), nil

		case lexer.STAR:
			l, lok := asNumber(left)
			r, rok := asNumber(right)
			if !lok || !rok {
				return Object{}, fmt.Errorf("runtime error: '*' supports only numbers")
			}
			return NumberObject(l * r), nil

		case lexer.SLASH:
			l, lok := asNumber(left)
			r, rok := asNumber(right)
			if !lok || !rok {
				return Object{}, fmt.Errorf("runtime error: '/' supports only numbers")
			}
			if r == 0 {
				return Object{}, fmt.Errorf("runtime error: division by zero")
			}
			return NumberObject(l / r), nil

		case lexer.LT:
			l, lok := asNumber(left)
			r, rok := asNumber(right)
			if !lok || !rok {
				return Object{}, fmt.Errorf("runtime error: '<' supports only numbers")
			}
			return BoolObject(l < r), nil

		case lexer.LTEQ:
			l, lok := asNumber(left)
			r, rok := asNumber(right)
			if !lok || !rok {
				return Object{}, fmt.Errorf("runtime error: '<=' supports only numbers")
			}
			return BoolObject(l <= r), nil

		case lexer.GT:
			l, lok := asNumber(left)
			r, rok := asNumber(right)
			if !lok || !rok {
				return Object{}, fmt.Errorf("runtime error: '>' supports only numbers")
			}
			return BoolObject(l > r), nil

		case lexer.GTEQ:
			l, lok := asNumber(left)
			r, rok := asNumber(right)
			if !lok || !rok {
				return Object{}, fmt.Errorf("runtime error: '>=' supports only numbers")
			}
			return BoolObject(l >= r), nil

		case lexer.EQEQ:
			return BoolObject(equals(left, right)), nil

		case lexer.NEQ:
			return BoolObject(!equals(left, right)), nil

		default:
			return Object{}, fmt.Errorf("runtime error: unknown binary operator %s", e.Operator)
		}

	default:
		return Object{}, fmt.Errorf("runtime error: unknown expression type %T", expr)
	}
}

func asNumber(o Object) (float64, bool) {
	if o.Type != ObjectNumber {
		return 0, false
	}
	v, ok := o.Value.(float64)
	return v, ok
}

func asString(o Object) (string, bool) {
	if o.Type != ObjectString {
		return "", false
	}
	v, ok := o.Value.(string)
	return v, ok
}

func asBool(o Object) (bool, bool) {
	if o.Type != ObjectBool {
		return false, false
	}
	v, ok := o.Value.(bool)
	return v, ok
}

func asFunction(o Object) (*Function, bool) {
	if o.Type != ObjectFunction {
		return nil, false
	}
	fn, ok := o.Value.(*Function)
	return fn, ok
}

func equals(left, right Object) bool {
	if left.Type != right.Type {
		return false
	}

	switch left.Type {
	case ObjectNil:
		return true
	case ObjectNumber:
		l, _ := asNumber(left)
		r, _ := asNumber(right)
		return l == r
	case ObjectString:
		l, _ := asString(left)
		r, _ := asString(right)
		return l == r
	case ObjectBool:
		l, _ := asBool(left)
		r, _ := asBool(right)
		return l == r
	case ObjectFunction:
		return left.Value == right.Value
	default:
		return false
	}
}
