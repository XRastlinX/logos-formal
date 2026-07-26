package lexer

import (
	"os"
	"testing"

	"github.com/XRastlinX/logos-formal/pkg/sysml/token"
)

func TestLexer_1_BasicTokens(t *testing.T) {
	input := `calc def AltitudeToPressure { in attribute altitude :> ISQ::length; }`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
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
		{token.RBRACE, "}"},
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

func TestLexer_2_UnitAndQualifiedBoundaries(t *testing.T) {
	input := `10[km] ISQ::length :>`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.UNIT, "10[km]"},
		{token.QIDENT, "ISQ::length"},
		{token.COLON_GT, ":>"},
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

func TestLexer_3_MultiCharOperators(t *testing.T) {
	input := `>= == != <= :>`
	tests := []struct {
		expectedType    token.TokenType
	}{
		{token.GTE},
		{token.EQ},
		{token.NOT_EQ},
		{token.LTE},
		{token.COLON_GT},
		{token.EOF},
	}
	l := New([]byte(input))
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
	}
}

func TestLexer_4_MinusOperator(t *testing.T) {
	input := `1 - 0.5`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.NUMBER, "1"},
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
	}
}

func TestLexer_5_CommentSourceSpans(t *testing.T) {
	input := "calc\n//comment\n 10[km]"
	l := New([]byte(input))
	
	tok1 := l.NextToken()
	if tok1.Line != 1 || tok1.Column != 1 || tok1.Start != 0 || tok1.End != 4 {
		t.Fatalf("tok1 span incorrect: %+v", tok1)
	}

	tok2 := l.NextToken()
	if tok2.Type != token.UNIT || tok2.Line != 3 || tok2.Column != 2 {
		t.Fatalf("tok2 span incorrect (comments shouldn't alter subsequent relative spans): %+v", tok2)
	}
}

func TestLexer_6_IllegalCharactersAndMalformed(t *testing.T) {
	tests := []struct {
		input   string
		badLex  string
	}{
		{"calc $ def", "$"},
		{"1.2.3", "1.2.3"},
		{"10[km", "10[km"},
		{"/* unterminated", "unterminated block comment"},
		{"ISQ::", "ISQ::"},
		{"ISQ:::", "ISQ:::"},
	}
	
	for _, tt := range tests {
		l := New([]byte(tt.input))
		var illegalTok token.Token
		for {
			tok := l.NextToken()
			if tok.Type == token.ILLEGAL {
				illegalTok = tok
				break
			}
			if tok.Type == token.EOF {
				break
			}
		}
		if illegalTok.Type != token.ILLEGAL {
			t.Fatalf("expected ILLEGAL for %q, got none", tt.input)
		}
		if illegalTok.Literal != tt.badLex {
			t.Fatalf("expected ILLEGAL literal %q, got %q", tt.badLex, illegalTok.Literal)
		}
	}
}

func TestLexer_7_RepeatedRuns(t *testing.T) {
	input := `ISQ::length == 10[km]`
	
	l1 := New([]byte(input))
	l2 := New([]byte(input))
	
	for {
		t1 := l1.NextToken()
		t2 := l2.NextToken()
		if t1 != t2 {
			t.Fatalf("tokens diverged: %+v != %+v", t1, t2)
		}
		if t1.Type == token.EOF {
			break
		}
	}
}

func TestLexer_8_GoldenSysMLFixture(t *testing.T) {
	// Need to step out from pkg/sysml/lexer to the root
	bytes, err := os.ReadFile("../../../04_Dual_Projection/CTPhase0ScalarChain.sysml")
	if err != nil {
		t.Fatalf("failed to read golden fixture: %v", err)
	}
	
	l := New(bytes)
	tokCount := 0
	for {
		tok := l.NextToken()
		if tok.Type == token.ILLEGAL {
			t.Fatalf("golden fixture yielded ILLEGAL token: %+v", tok)
		}
		if tok.Type == token.EOF {
			break
		}
		tokCount++
	}
	
	if tokCount < 50 {
		t.Fatalf("expected at least 50 tokens from golden fixture, got %d", tokCount)
	}
}
