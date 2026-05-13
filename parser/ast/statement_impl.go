package ast

type ExpressionStatement struct {
	Expression Expression
}

func (e *ExpressionStatement) statementNode() {}

type PrintStatement struct {
	Expression Expression
}

func (p *PrintStatement) statementNode() {}

type VarStatement struct {
	Name        string
	Initializer Expression
}

func (v *VarStatement) statementNode() {}

type BlockStatement struct {
	Statements []Statement
}

func (b *BlockStatement) statementNode() {}

type IfStatement struct {
	Condition  Expression
	ThenBranch Statement
	ElseBranch Statement
}

func (i *IfStatement) statementNode() {}

type WhileStatement struct {
	Condition Expression
	Body      Statement
}

func (w *WhileStatement) statementNode() {}

type FunctionStatement struct {
	Name   string
	Params []string
	Body   []Statement
}

func (f *FunctionStatement) statementNode() {}

type ReturnStatement struct {
	Value Expression
}

func (r *ReturnStatement) statementNode() {}
