package lexer

import (
	"testing"

	"github.com/XRastlinX/logos-formal/pkg/sysml/token"
)

func TestLexer_BasicTokens(t *testing.T) {
	input := `package CTPhase0ScalarChain {
    // Morphism f
    calc def AltitudeToPressure {
        in attribute altitude :> ISQ::length;
        return :> ISQ::pressure = (1 - altitude / 10[km]);
    }
}`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.PACKAGE, "package"},
		{token.IDENT, "CTPhase0ScalarChain"},
		{token.LBRACE, "{"},
		{token.CALC, "calc"},
		{token.DEF, "def"},
		{token.IDENT, "AltitudeToPressure"},
		{token.LBRACE, "{"},
		{token.IN, "in"},
		{token.ATTRIBUTE, "attribute"},
		{token.IDENT, "altitude"},
		{token.COLON_GT, ":>"},
		{token.QIDENT, "ISQ::length"},
		{token.SEMICOLON, ";"},
		{token.RETURN, "return"},
		{token.COLON_GT, ":>"},
		{token.QIDENT, "ISQ::pressure"},
		{token.ASSIGN, "="},
		{token.LPAREN, "("},
		{token.NUMBER, "1"},
		{token.MINUS, "-"},
		{token.IDENT, "altitude"},
		{token.SLASH, "/"},
		{token.UNIT, "10[km]"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	l := New([]byte(input))

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexer_OperatorsAndBoundaries(t *testing.T) {
	input := `10[km] ISQ::length :> >= == != - 0.5`
	
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.UNIT, "10[km]"},
		{token.QIDENT, "ISQ::length"},
		{token.COLON_GT, ":>"},
		{token.GTE, ">="},
		{token.EQ, "=="},
		{token.NOT_EQ, "!="},
		{token.MINUS, "-"},
		{token.NUMBER, "0.5"},
		{token.EOF, ""},
	}

	l := New([]byte(input))
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexer_SourceSpans(t *testing.T) {
	input := "calc def\n//comment\n 10[km]"
	l := New([]byte(input))
	
	tok1 := l.NextToken()
	if tok1.Line != 1 || tok1.Column != 1 || tok1.Start != 0 || tok1.End != 4 {
		t.Fatalf("tok1 span incorrect: %+v", tok1)
	}

	tok2 := l.NextToken()
	if tok2.Line != 1 || tok2.Column != 6 || tok2.Start != 5 || tok2.End != 8 {
		t.Fatalf("tok2 span incorrect: %+v", tok2)
	}

	tok3 := l.NextToken()
	if tok3.Type != token.UNIT || tok3.Line != 3 || tok3.Column != 2 {
		t.Fatalf("tok3 span incorrect (comments shouldn't alter subsequent relative spans if correct): %+v", tok3)
	}
}

func TestLexer_Illegal(t *testing.T) {
	input := `calc $ def`
	l := New([]byte(input))
	
	_ = l.NextToken() // calc
	tok := l.NextToken()
	if tok.Type != token.ILLEGAL || tok.Literal != "$" {
		t.Fatalf("expected ILLEGAL '$', got %q", tok.Literal)
	}
}
