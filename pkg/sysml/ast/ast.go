package ast

// SourceSpan represents the origin of the node in the exact 000 witness bytes
type SourceSpan struct {
	Line   int
	Column int
	Length int
}

// Node represents a generic AST node for the 010 Validator
type Node interface {
	Span() SourceSpan
	TokenLiteral() string
}

// Expr represents an evaluable mathematical expression
type Expr interface {
	Node
	expressionNode()
}

// Package represents the parsed sysml package
type Package struct {
	Name        string
	Imports     []Import
	Calcs       []CalcDef
	Constraints []ConstraintDef
	
	nodeSpan SourceSpan
}

func (p *Package) Span() SourceSpan { return p.nodeSpan }
func (p *Package) TokenLiteral() string { return "package" }

// Import represents a namespace inclusion
type Import struct {
	Path string
	
	nodeSpan SourceSpan
}

func (i *Import) Span() SourceSpan { return i.nodeSpan }
func (i *Import) TokenLiteral() string { return "import" }

// Parameter represents a directed attribute input/return with physical dimensions
type Parameter struct {
	Name      string
	Quantity  string // e.g. ISQ::length, ISQ::pressure
	
	nodeSpan SourceSpan
}

func (p *Parameter) Span() SourceSpan { return p.nodeSpan }
func (p *Parameter) TokenLiteral() string { return p.Name }

// CalcDef represents a candidate Set morphism (deterministic calculation)
type CalcDef struct {
	Name   string
	Inputs []Parameter
	Return Parameter
	Body   Expr
	
	nodeSpan SourceSpan
}

func (c *CalcDef) Span() SourceSpan { return c.nodeSpan }
func (c *CalcDef) TokenLiteral() string { return "calc def" }

// ConstraintDef represents a formal predicate in Rel
type ConstraintDef struct {
	Name   string
	Inputs []Parameter
	Body   Expr
	
	nodeSpan SourceSpan
}

func (c *ConstraintDef) Span() SourceSpan { return c.nodeSpan }
func (c *ConstraintDef) TokenLiteral() string { return "constraint def" }

// --- Expression Nodes ---

type NumberLiteral struct {
	Value float64
	
	nodeSpan SourceSpan
}
func (nl *NumberLiteral) Span() SourceSpan { return nl.nodeSpan }
func (nl *NumberLiteral) TokenLiteral() string { return "number" }
func (nl *NumberLiteral) expressionNode() {}

type Identifier struct {
	Value string
	
	nodeSpan SourceSpan
}
func (i *Identifier) Span() SourceSpan { return i.nodeSpan }
func (i *Identifier) TokenLiteral() string { return i.Value }
func (i *Identifier) expressionNode() {}

type InfixExpression struct {
	Left     Expr
	Operator string // +, -, *, /, ==, >=, <=, and
	Right    Expr
	
	nodeSpan SourceSpan
}
func (ie *InfixExpression) Span() SourceSpan { return ie.nodeSpan }
func (ie *InfixExpression) TokenLiteral() string { return ie.Operator }
func (ie *InfixExpression) expressionNode() {}

type CallExpression struct {
	Function Expr // Identifier representing the calc def
	Args     []Expr
	
	nodeSpan SourceSpan
}
func (ce *CallExpression) Span() SourceSpan { return ce.nodeSpan }
func (ce *CallExpression) TokenLiteral() string { return "(" }
func (ce *CallExpression) expressionNode() {}
