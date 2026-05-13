package parser

import (
	"errors"
	"fmt"
	"mycompiler/lexer"
	"mycompiler/parser/ast"
	"strconv"
)

// Parser represents a recursive descent parser for the language.
type Parser struct {
	tokens   []lexer.Token
	position int
	errors   []error
}

// NewParser creates a new parser instance with the given tokens.
func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens:   tokens,
		position: 0,
		errors:   make([]error, 0),
	}
}

// Parse parses the tokens and returns a slice of statements.
func (p *Parser) Parse() []ast.Statement {
	var statements []ast.Statement
	for !p.isAtEnd() {
		stmt := p.parseDeclaration()
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}
	return statements
}

// Errors returns all parsing errors encountered.
func (p *Parser) Errors() []error {
	return p.errors
}

// HasErrors returns true if any parsing errors were encountered.
func (p *Parser) HasErrors() bool {
	return len(p.errors) > 0
}

// --- Statements ---

func (p *Parser) parseDeclaration() ast.Statement {
	if p.match(lexer.FUN) {
		return p.parseFunctionDeclaration()
	}
	if p.match(lexer.VAR) {
		return p.parseVarDeclaration()
	}
	return p.parseStatement()
}

func (p *Parser) parseStatement() ast.Statement {
	if p.match(lexer.RETURN) {
		return p.parseReturnStatement()
	}
	if p.match(lexer.IF) {
		return p.parseIfStatement()
	}
	if p.match(lexer.WHILE) {
		return p.parseWhileStatement()
	}
	if p.match(lexer.PRINT) {
		return p.parsePrintStatement()
	}
	if p.match(lexer.LBRACE) {
		return &ast.BlockStatement{Statements: p.parseBlock()}
	}

	return p.parseExpressionStatement()
}

func (p *Parser) parseFunctionDeclaration() ast.Statement {
	name := p.consume(lexer.ID, "expected function name")
	p.consume(lexer.LPAREN, "expected '(' after function name")

	params := make([]string, 0)
	if !p.check(lexer.RPAREN) {
		for {
			param := p.consume(lexer.ID, "expected parameter name")
			params = append(params, param.Value)

			if !p.match(lexer.COMMA) {
				break
			}
		}
	}

	p.consume(lexer.RPAREN, "expected ')' after function parameters")
	p.consume(lexer.LBRACE, "expected '{' before function body")
	body := p.parseBlock()

	return &ast.FunctionStatement{Name: name.Value, Params: params, Body: body}
}

func (p *Parser) parseReturnStatement() ast.Statement {
	var value ast.Expression
	if !p.check(lexer.SEMICOLON) {
		value = p.parseExpression()
	}
	p.consume(lexer.SEMICOLON, "expected ';' after return value")
	return &ast.ReturnStatement{Value: value}
}

func (p *Parser) parseVarDeclaration() ast.Statement {
	name := p.consume(lexer.ID, "expected variable name")

	var initializer ast.Expression
	if p.match(lexer.EQ) {
		initializer = p.parseExpression()
	}

	p.consume(lexer.SEMICOLON, "expected ';' after variable declaration")
	return &ast.VarStatement{Name: name.Value, Initializer: initializer}
}

func (p *Parser) parseIfStatement() ast.Statement {
	p.consume(lexer.LPAREN, "expected '(' after 'if'")
	condition := p.parseExpression()
	p.consume(lexer.RPAREN, "expected ')' after if condition")

	thenBranch := p.parseStatement()
	var elseBranch ast.Statement

	if p.match(lexer.ELSE) {
		elseBranch = p.parseStatement()
	}

	return &ast.IfStatement{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}
}

func (p *Parser) parseWhileStatement() ast.Statement {
	p.consume(lexer.LPAREN, "expected '(' after 'while'")
	condition := p.parseExpression()
	p.consume(lexer.RPAREN, "expected ')' after while condition")

	body := p.parseStatement()
	return &ast.WhileStatement{Condition: condition, Body: body}
}

func (p *Parser) parsePrintStatement() ast.Statement {
	value := p.parseExpression()
	p.consume(lexer.SEMICOLON, "expected ';' after print expression")
	return &ast.PrintStatement{Expression: value}
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	expr := p.parseExpression()
	p.consume(lexer.SEMICOLON, "expected ';' after expression")
	return &ast.ExpressionStatement{Expression: expr}
}

func (p *Parser) parseBlock() []ast.Statement {
	var statements []ast.Statement

	for !p.check(lexer.RBRACE) && !p.isAtEnd() {
		stmt := p.parseDeclaration()
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	p.consume(lexer.RBRACE, "expected '}' after block")
	return statements
}

func (p *Parser) parseExpression() ast.Expression {
	return p.parseAssignment()
}

// Assign
func (p *Parser) parseAssignment() ast.Expression {
	expr := p.parseLogicalOr()

	if p.match(lexer.EQ) {
		value := p.parseAssignment()

		if variable, ok := expr.(*ast.VariableExpression); ok {
			return &ast.AssignExpression{Name: variable.Name, Value: value}
		}

		p.error(fmt.Sprintf("invalid assignment target at position %d", p.previous().Position))
		return nil
	}

	return expr
}

// ||
func (p *Parser) parseLogicalOr() ast.Expression {
	expr := p.parseLogicalAnd()

	for p.match(lexer.OR) {
		op := p.previous().Type
		right := p.parseLogicalAnd()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// &&
func (p *Parser) parseLogicalAnd() ast.Expression {
	expr := p.parseEquality()

	for p.match(lexer.AND) {
		op := p.previous().Type
		right := p.parseEquality()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// ==, !=
func (p *Parser) parseEquality() ast.Expression {
	expr := p.parseComparison()

	for p.match(lexer.EQEQ, lexer.NEQ) {
		op := p.previous().Type
		right := p.parseComparison()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// <, >, <=, >=
func (p *Parser) parseComparison() ast.Expression {
	expr := p.parseTerm()

	for p.match(lexer.LT, lexer.LTEQ, lexer.GT, lexer.GTEQ) {
		op := p.previous().Type
		right := p.parseTerm()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// +, -
func (p *Parser) parseTerm() ast.Expression {
	expr := p.parseFactor()

	for p.match(lexer.PLUS, lexer.MINUS) {
		op := p.previous().Type
		right := p.parseFactor()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// *, /
func (p *Parser) parseFactor() ast.Expression {
	expr := p.parseUnary()

	for p.match(lexer.STAR, lexer.SLASH) {
		op := p.previous().Type
		right := p.parseUnary()
		expr = &ast.BinaryExpression{Left: expr, Operator: op, Right: right}
	}

	return expr
}

// !, -
func (p *Parser) parseUnary() ast.Expression {
	if p.match(lexer.EXCL, lexer.MINUS) {
		op := p.previous().Type
		right := p.parseUnary()
		return &ast.UnaryExpression{Operator: op, Right: right}
	}
	return p.parseCall()
}

func (p *Parser) parseCall() ast.Expression {
	expr := p.parsePrimary()

	for {
		if p.match(lexer.LPAREN) {
			expr = p.finishCall(expr)
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) finishCall(callee ast.Expression) ast.Expression {
	arguments := make([]ast.Expression, 0)
	if !p.check(lexer.RPAREN) {
		for {
			arguments = append(arguments, p.parseExpression())
			if !p.match(lexer.COMMA) {
				break
			}
		}
	}

	p.consume(lexer.RPAREN, "expected ')' after arguments")
	return &ast.CallExpression{Callee: callee, Arguments: arguments}
}

// 9. Primary
func (p *Parser) parsePrimary() ast.Expression {
	if p.match(lexer.NUMBER) {
		valStr := p.previous().Value
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			p.error(fmt.Sprintf("failed to parse number: %s", valStr))
			return nil
		}
		return &ast.NumberExpression{Value: val}
	}

	if p.match(lexer.STRING) {
		return &ast.StringExpression{Value: p.previous().Value}
	}

	if p.match(lexer.TRUE) {
		return &ast.BoolExpression{Value: true}
	}

	if p.match(lexer.FALSE) {
		return &ast.BoolExpression{Value: false}
	}

	if p.match(lexer.ID) {
		return &ast.VariableExpression{Name: p.previous().Value}
	}

	if p.match(lexer.LPAREN) {
		expr := p.parseExpression()
		p.consume(lexer.RPAREN, "expected ')' after expression")
		return expr
	}

	token := p.peek()
	p.error(fmt.Sprintf("expected expression at position %d, found: %s", token.Position, token.Type))
	return nil
}

func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

func (p *Parser) advance() lexer.Token {
	if !p.isAtEnd() {
		p.position++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == lexer.EOF
}

func (p *Parser) peek() lexer.Token {
	if p.position >= len(p.tokens) {
		return lexer.NewToken(lexer.EOF, "", -1)
	}
	return p.tokens[p.position]
}

func (p *Parser) previous() lexer.Token {
	return p.tokens[p.position-1]
}

func (p *Parser) consume(t lexer.TokenType, message string) lexer.Token {
	if p.check(t) {
		return p.advance()
	}
	token := p.peek()
	p.error(fmt.Sprintf("at position %d: %s, found: %s", token.Position, message, token.Type))
	return token
}

func (p *Parser) error(message string) {
	p.errors = append(p.errors, errors.New(message))
}
