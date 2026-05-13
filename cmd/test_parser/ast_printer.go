package main

import (
	"fmt"
	"mycompiler/parser/ast"
)

type AstPrinter struct{}

func (ap *AstPrinter) Print(statements []ast.Statement) {
	fmt.Println("Root (Program)")
	for i, stmt := range statements {
		ap.printNode(stmt, "", i == len(statements)-1)
	}
}

func (ap *AstPrinter) printNode(node interface{}, indent string, isLast bool) {
	if node == nil {
		return
	}

	marker := "├── "
	if isLast {
		marker = "└── "
	}
	fmt.Print(indent + marker)

	childIndent := indent + "│   "
	if isLast {
		childIndent = indent + "    "
	}

	switch n := node.(type) {

	case *ast.VarStatement:
		fmt.Printf("VarStatement: %s\n", n.Name)
		if n.Initializer != nil {
			ap.printNode(n.Initializer, childIndent, true)
		}

	case *ast.PrintStatement:
		fmt.Println("PrintStatement")
		ap.printNode(n.Expression, childIndent, true)

	case *ast.IfStatement:
		fmt.Println("IfStatement")
		ap.printNode(n.Condition, childIndent, false)

		hasElse := n.ElseBranch != nil
		ap.printNode(n.ThenBranch, childIndent, !hasElse)

		if hasElse {
			ap.printNode(n.ElseBranch, childIndent, true)
		}

	case *ast.WhileStatement:
		fmt.Println("WhileStatement")
		ap.printNode(n.Condition, childIndent, false)
		ap.printNode(n.Body, childIndent, true)

	case *ast.BlockStatement:
		fmt.Println("BlockStatement")
		for i, stmt := range n.Statements {
			ap.printNode(stmt, childIndent, i == len(n.Statements)-1)
		}

	case *ast.FunctionStatement:
		fmt.Printf("FunctionStatement: %s\n", n.Name)
		if len(n.Params) > 0 {
			fmt.Print(childIndent + "├── Params")
			for i, p := range n.Params {
				sep := ", "
				if i == len(n.Params)-1 {
					sep = ""
				}
				fmt.Print(" " + p + sep)
			}
			fmt.Println()
		}
		for i, stmt := range n.Body {
			ap.printNode(stmt, childIndent, i == len(n.Body)-1)
		}

	case *ast.ReturnStatement:
		fmt.Println("ReturnStatement")
		if n.Value != nil {
			ap.printNode(n.Value, childIndent, true)
		}

	case *ast.ExpressionStatement:
		fmt.Println("ExpressionStatement")
		ap.printNode(n.Expression, childIndent, true)

	case *ast.BinaryExpression:
		fmt.Printf("BinaryExpression: %s\n", n.Operator)
		ap.printNode(n.Left, childIndent, false)
		ap.printNode(n.Right, childIndent, true)

	case *ast.UnaryExpression:
		fmt.Printf("UnaryExpression: %s\n", n.Operator)
		ap.printNode(n.Right, childIndent, true)

	case *ast.AssignExpression:
		fmt.Printf("AssignExpression: %s =\n", n.Name)
		ap.printNode(n.Value, childIndent, true)

	case *ast.CallExpression:
		fmt.Println("CallExpression")
		ap.printNode(n.Callee, childIndent, len(n.Arguments) == 0)
		for i, arg := range n.Arguments {
			ap.printNode(arg, childIndent, i == len(n.Arguments)-1)
		}

	case *ast.NumberExpression:
		fmt.Printf("Number: %v\n", n.Value)

	case *ast.StringExpression:
		fmt.Printf("String: %s\n", n.Value)

	case *ast.BoolExpression:
		fmt.Printf("Bool: %v\n", n.Value)

	case *ast.VariableExpression:
		fmt.Printf("Variable: %s\n", n.Name)

	default:
		fmt.Printf("Unknown Node: %T\n", n)
	}
}
