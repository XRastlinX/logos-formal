package parser

import (
	"fmt"

	"github.com/XRastlinX/logos-formal/pkg/sysml/ast"
	"github.com/XRastlinX/logos-formal/pkg/sysml/lexer"
	"github.com/XRastlinX/logos-formal/pkg/sysml/token"
)

// Status is the bounded parser outcome consumed by the later 010 evaluator.
// It is not an authorization decision.
type Status string

const (
	StatusParsed     Status = "PARSED"
	StatusReject     Status = "REJECT"
	StatusNoDecision Status = "NO_DECISION"
)

// Diagnostic records why parsing could not produce a usable Phase 0 tree.
type Diagnostic struct {
	Status  Status
	Code    string
	Message string
	Span    ast.SourceSpan
}

// Result is the complete deterministic parser result. A Package is usable by
// later evaluation only when Status is PARSED and Diagnostics is empty.
type Result struct {
	Package     *ast.Package
	Status      Status
	Diagnostics []Diagnostic
}

const (
	precedenceLowest = iota
	precedenceAnd
	precedenceEquals
	precedenceCompare
	precedenceSum
	precedenceProduct
	precedencePrefix
	precedenceCall
)

var precedences = map[token.TokenType]int{
	token.AND:      precedenceAnd,
	token.EQ:       precedenceEquals,
	token.NOT_EQ:   precedenceEquals,
	token.GT:       precedenceCompare,
	token.LT:       precedenceCompare,
	token.GTE:      precedenceCompare,
	token.LTE:      precedenceCompare,
	token.PLUS:     precedenceSum,
	token.MINUS:    precedenceSum,
	token.ASTERISK: precedenceProduct,
	token.SLASH:    precedenceProduct,
	token.LPAREN:   precedenceCall,
}

type parser struct {
	lexer *lexer.Lexer

	current token.Token
	peek    token.Token

	status      Status
	diagnostics []Diagnostic
}

// ParsePackage parses only the canonical Phase 0 fragment:
//
//	package IDENT {
//	  calc def IDENT {
//	    [in attribute IDENT :> QIDENT;]*
//	    return :> QIDENT = Expr;
//	  }
//	  constraint def IDENT {
//	    [in attribute IDENT :> QIDENT;]*
//	    Expr
//	  }
//	}
//
// Legal SysML forms outside this fragment produce NO_DECISION. Malformed
// instances of the supported fragment produce REJECT.
func ParsePackage(input []byte) Result {
	p := newParser(input)
	pkg := p.parsePackage()

	if p.status == "" {
		p.status = StatusParsed
	}
	if p.status != StatusParsed {
		pkg = nil
	}

	return Result{
		Package:     pkg,
		Status:      p.status,
		Diagnostics: append([]Diagnostic(nil), p.diagnostics...),
	}
}

func newParser(input []byte) *parser {
	p := &parser{lexer: lexer.New(input)}
	p.advance()
	p.advance()
	return p
}

func (p *parser) advance() {
	p.current = p.peek
	p.peek = p.lexer.NextToken()

	if p.peek.Type == token.ILLEGAL && p.status == "" {
		p.addDiagnostic(
			StatusReject,
			"LEXICAL_ERROR",
			fmt.Sprintf("illegal token %q", p.peek.Literal),
			spanFromToken(p.peek),
		)
	}
}

func (p *parser) parsePackage() *ast.Package {
	if p.failed() {
		return nil
	}
	if !p.expectCurrent(token.PACKAGE, "package declaration") {
		return nil
	}

	start := spanFromToken(p.current)
	if !p.expectPeek(token.IDENT, "package name") {
		return nil
	}
	name := p.current.Literal

	if !p.expectPeek(token.LBRACE, "package body") {
		return nil
	}
	p.advance()

	pkg := &ast.Package{
		Name:     name,
		NodeSpan: start,
	}

	for !p.failed() && p.current.Type != token.RBRACE && p.current.Type != token.EOF {
		switch p.current.Type {
		case token.CALC:
			calc := p.parseCalcDef()
			if calc != nil {
				pkg.Calcs = append(pkg.Calcs, *calc)
			}
		case token.CONSTRAINT:
			constraint := p.parseConstraintDef()
			if constraint != nil {
				pkg.Constraints = append(pkg.Constraints, *constraint)
			}
		case token.PRIVATE, token.IMPORT:
			p.unsupported(
				"UNSUPPORTED_IMPORT",
				"imports are outside the currently implemented canonical parser fragment",
				p.current,
			)
		default:
			p.unsupported(
				"UNSUPPORTED_PACKAGE_MEMBER",
				fmt.Sprintf("unsupported package member beginning with %q", p.current.Literal),
				p.current,
			)
		}

		if !p.failed() {
			p.advance()
		}
	}

	if p.failed() {
		return nil
	}
	if p.current.Type == token.EOF {
		p.reject("UNTERMINATED_PACKAGE", "expected } before end of input", p.current)
		return nil
	}

	pkg.NodeSpan.EndByte = p.current.End
	if p.peek.Type != token.EOF {
		p.unsupported(
			"TRAILING_CONTENT",
			fmt.Sprintf("content after package is outside Phase 0: %q", p.peek.Literal),
			p.peek,
		)
		return nil
	}

	return pkg
}

func (p *parser) parseCalcDef() *ast.CalcDef {
	start := spanFromToken(p.current)
	if !p.expectPeek(token.DEF, "calc def") ||
		!p.expectPeek(token.IDENT, "calculation name") {
		return nil
	}
	name := p.current.Literal
	if !p.expectPeek(token.LBRACE, "calculation body") {
		return nil
	}
	p.advance()

	calc := &ast.CalcDef{Name: name, NodeSpan: start}
	for !p.failed() && p.current.Type == token.IN {
		parameter := p.parseInputParameter()
		if parameter == nil {
			return nil
		}
		calc.Inputs = append(calc.Inputs, *parameter)
		p.advance()
	}

	if p.failed() {
		return nil
	}
	if p.current.Type == token.RBRACE {
		p.reject("MISSING_RETURN", "calc def requires exactly one final return", p.current)
		return nil
	}
	if p.current.Type != token.RETURN {
		p.unsupported(
			"UNSUPPORTED_CALC_MEMBER_ORDER",
			"calc def supports only input attributes followed by one final return",
			p.current,
		)
		return nil
	}

	returnType, body := p.parseReturn()
	if p.failed() {
		return nil
	}
	calc.Return = ast.Parameter{
		Quantity: returnType.Literal,
		NodeSpan: spanFromToken(returnType),
	}
	calc.Body = body

	if p.peek.Type == token.IN {
		p.unsupported(
			"UNSUPPORTED_CALC_MEMBER_ORDER",
			"input attributes may not follow the final return in Phase 0",
			p.peek,
		)
		return nil
	}
	if p.peek.Type == token.RETURN {
		p.reject("MULTIPLE_RETURNS", "calc def permits exactly one return", p.peek)
		return nil
	}
	if !p.expectPeek(token.RBRACE, "end of calculation") {
		return nil
	}
	calc.NodeSpan.EndByte = p.current.End
	return calc
}

func (p *parser) parseConstraintDef() *ast.ConstraintDef {
	start := spanFromToken(p.current)
	if !p.expectPeek(token.DEF, "constraint def") ||
		!p.expectPeek(token.IDENT, "constraint name") {
		return nil
	}
	name := p.current.Literal
	if !p.expectPeek(token.LBRACE, "constraint body") {
		return nil
	}
	p.advance()

	constraint := &ast.ConstraintDef{Name: name, NodeSpan: start}
	for !p.failed() && p.current.Type == token.IN {
		parameter := p.parseInputParameter()
		if parameter == nil {
			return nil
		}
		constraint.Inputs = append(constraint.Inputs, *parameter)
		p.advance()
	}

	if p.failed() {
		return nil
	}
	if p.current.Type == token.RBRACE {
		p.reject("MISSING_CONSTRAINT_BODY", "constraint def requires one final expression", p.current)
		return nil
	}
	if p.current.Type == token.RETURN {
		p.unsupported(
			"UNSUPPORTED_CONSTRAINT_RETURN",
			"constraint def requires a bare final expression, not return syntax",
			p.current,
		)
		return nil
	}

	constraint.Body = p.parseExpression(precedenceLowest)
	if constraint.Body == nil || p.failed() {
		return nil
	}
	if p.peek.Type == token.SEMICOLON {
		p.unsupported(
			"UNSUPPORTED_CONSTRAINT_TERMINATOR",
			"the canonical Phase 0 constraint expression has no trailing semicolon",
			p.peek,
		)
		return nil
	}
	if p.peek.Type == token.IN {
		p.unsupported(
			"UNSUPPORTED_CONSTRAINT_MEMBER_ORDER",
			"input attributes may not follow the final constraint expression",
			p.peek,
		)
		return nil
	}
	if !p.expectPeek(token.RBRACE, "end of constraint") {
		return nil
	}
	constraint.NodeSpan.EndByte = p.current.End
	return constraint
}

func (p *parser) parseInputParameter() *ast.Parameter {
	start := spanFromToken(p.current)
	if !p.expectPeek(token.ATTRIBUTE, "input attribute") ||
		!p.expectPeek(token.IDENT, "input name") {
		return nil
	}
	name := p.current.Literal
	if !p.expectPeek(token.COLON_GT, "input type binding") {
		return nil
	}
	if !p.expectPeek(token.QIDENT, "qualified input quantity") {
		return nil
	}
	quantity := p.current.Literal
	if !p.expectPeek(token.SEMICOLON, "input terminator") {
		return nil
	}

	start.EndByte = p.current.End
	return &ast.Parameter{
		Name:     name,
		Quantity: quantity,
		NodeSpan: start,
	}
}

func (p *parser) parseReturn() (token.Token, ast.Expr) {
	if !p.expectPeek(token.COLON_GT, "return type binding") {
		return token.Token{}, nil
	}
	if !p.expectPeek(token.QIDENT, "qualified return quantity") {
		return token.Token{}, nil
	}
	returnType := p.current
	if !p.expectPeek(token.ASSIGN, "return expression assignment") {
		return token.Token{}, nil
	}
	p.advance()
	body := p.parseExpression(precedenceLowest)
	if body == nil || p.failed() {
		return token.Token{}, nil
	}
	if !p.expectPeek(token.SEMICOLON, "return terminator") {
		return token.Token{}, nil
	}
	return returnType, body
}

func (p *parser) parseExpression(precedence int) ast.Expr {
	if p.failed() {
		return nil
	}

	var left ast.Expr
	switch p.current.Type {
	case token.IDENT:
		left = &ast.Identifier{
			Value:    p.current.Literal,
			NodeSpan: spanFromToken(p.current),
		}
	case token.NUMBER:
		left = &ast.NumberLiteral{
			Lexeme:   p.current.Literal,
			NodeSpan: spanFromToken(p.current),
		}
	case token.UNIT:
		left = &ast.UnitLiteral{
			Lexeme:   p.current.Literal,
			NodeSpan: spanFromToken(p.current),
		}
	case token.MINUS:
		left = p.parsePrefixExpression()
	case token.LPAREN:
		left = p.parseGroupedExpression()
	default:
		p.reject(
			"EXPECTED_EXPRESSION",
			fmt.Sprintf("expected expression, found %q", p.current.Literal),
			p.current,
		)
		return nil
	}

	for !p.failed() &&
		p.peek.Type != token.SEMICOLON &&
		p.peek.Type != token.RBRACE &&
		p.peek.Type != token.EOF &&
		precedence < p.peekPrecedence() {
		switch p.peek.Type {
		case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH,
			token.EQ, token.NOT_EQ, token.GT, token.LT, token.GTE, token.LTE,
			token.AND:
			p.advance()
			left = p.parseInfixExpression(left)
		case token.LPAREN:
			p.advance()
			left = p.parseCallExpression(left)
		default:
			return left
		}
	}

	return left
}

func (p *parser) parsePrefixExpression() ast.Expr {
	start := spanFromToken(p.current)
	operator := p.current.Literal
	p.advance()
	right := p.parseExpression(precedencePrefix)
	if right == nil {
		return nil
	}
	start.EndByte = right.Span().EndByte
	return &ast.PrefixExpression{
		Operator: operator,
		Right:    right,
		NodeSpan: start,
	}
}

func (p *parser) parseGroupedExpression() ast.Expr {
	p.advance()
	expression := p.parseExpression(precedenceLowest)
	if expression == nil || p.failed() {
		return nil
	}
	if !p.expectPeek(token.RPAREN, "grouped expression") {
		return nil
	}
	return expression
}

func (p *parser) parseInfixExpression(left ast.Expr) ast.Expr {
	operatorToken := p.current
	precedence := p.currentPrecedence()
	p.advance()
	right := p.parseExpression(precedence)
	if right == nil {
		return nil
	}

	return &ast.InfixExpression{
		Left:     left,
		Operator: operatorToken.Literal,
		Right:    right,
		NodeSpan: ast.SourceSpan{
			StartByte: left.Span().StartByte,
			EndByte:   right.Span().EndByte,
			Line:      left.Span().Line,
			Column:    left.Span().Column,
		},
	}
}

func (p *parser) parseCallExpression(function ast.Expr) ast.Expr {
	if _, ok := function.(*ast.Identifier); !ok {
		p.unsupported(
			"UNSUPPORTED_CALL_TARGET",
			"Phase 0 calculation calls require a declared calculation identifier",
			p.current,
		)
		return nil
	}

	start := function.Span()
	call := &ast.CallExpression{
		Function: function,
		NodeSpan: start,
	}

	if p.peek.Type == token.RPAREN {
		p.advance()
		call.NodeSpan.EndByte = p.current.End
		return call
	}

	p.advance()
	first := p.parseExpression(precedenceLowest)
	if first == nil {
		return nil
	}
	call.Args = append(call.Args, first)

	for p.peek.Type == token.COMMA {
		p.advance()
		p.advance()
		arg := p.parseExpression(precedenceLowest)
		if arg == nil {
			return nil
		}
		call.Args = append(call.Args, arg)
	}

	if !p.expectPeek(token.RPAREN, "call expression") {
		return nil
	}
	call.NodeSpan.EndByte = p.current.End
	return call
}

func (p *parser) currentPrecedence() int {
	if precedence, ok := precedences[p.current.Type]; ok {
		return precedence
	}
	return precedenceLowest
}

func (p *parser) peekPrecedence() int {
	if precedence, ok := precedences[p.peek.Type]; ok {
		return precedence
	}
	return precedenceLowest
}

func (p *parser) expectCurrent(expected token.TokenType, context string) bool {
	if p.current.Type == expected {
		return true
	}
	p.reject(
		"SYNTAX_ERROR",
		fmt.Sprintf("expected %s token %q, found %q", context, expected, p.current.Literal),
		p.current,
	)
	return false
}

func (p *parser) expectPeek(expected token.TokenType, context string) bool {
	if p.peek.Type == expected {
		p.advance()
		return !p.failed()
	}
	if p.peek.Type == token.ILLEGAL {
		return false
	}
	p.reject(
		"SYNTAX_ERROR",
		fmt.Sprintf("expected %s token %q, found %q", context, expected, p.peek.Literal),
		p.peek,
	)
	return false
}

func (p *parser) reject(code, message string, tok token.Token) {
	p.addDiagnostic(StatusReject, code, message, spanFromToken(tok))
}

func (p *parser) unsupported(code, message string, tok token.Token) {
	p.addDiagnostic(StatusNoDecision, code, message, spanFromToken(tok))
}

func (p *parser) addDiagnostic(status Status, code, message string, span ast.SourceSpan) {
	if p.status != "" {
		return
	}
	p.status = status
	p.diagnostics = append(p.diagnostics, Diagnostic{
		Status:  status,
		Code:    code,
		Message: message,
		Span:    span,
	})
}

func (p *parser) failed() bool {
	return p.status == StatusReject || p.status == StatusNoDecision
}

func spanFromToken(tok token.Token) ast.SourceSpan {
	return ast.SourceSpan{
		StartByte: tok.Start,
		EndByte:   tok.End,
		Line:      tok.Line,
		Column:    tok.Column,
	}
}
