package lexer

import (
	"github.com/XRastlinX/logos-formal/pkg/sysml/token"
)

type Lexer struct {
	input        []byte
	position     int // current position in input (points to current char)
	readPosition int // current reading position in input (after current char)
	ch           byte

	// Tracking positions for SourceSpan
	line   int
	column int

	prevLine   int
	prevColumn int
}

func New(input []byte) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	l.prevLine = l.line
	l.prevColumn = l.column

	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespaceAndComments()

	startByte := l.position
	startLine := l.line
	startColumn := l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case ':':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.COLON_GT, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.COLON, l.ch)
		}
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isDigit(l.ch) {
			tok.Literal = l.readNumberOrUnit()
			// Check if it has a unit suffix [xxx] included
			if len(tok.Literal) > 0 && tok.Literal[len(tok.Literal)-1] == ']' {
				tok.Type = token.UNIT
			} else {
				tok.Type = token.NUMBER
			}
			tok.Start = startByte
			tok.End = l.position // l.position is currently the char AFTER the number/unit because readNumberOrUnit advances
			tok.Line = startLine
			tok.Column = startColumn
			return tok
		} else if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			if isQualified(tok.Literal) {
				tok.Type = token.QIDENT
			} else {
				tok.Type = token.LookupIdent(tok.Literal)
			}
			tok.Start = startByte
			tok.End = l.position
			tok.Line = startLine
			tok.Column = startColumn
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	tok.Start = startByte
	tok.End = l.position
	if l.ch != 0 {
		tok.End = l.position + 1
	}
	tok.Line = startLine
	tok.Column = startColumn

	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		if l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			l.readChar()
			continue
		}
		// Check for // comment
		if l.ch == '/' && l.peekChar() == '/' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			continue
		}
		// Check for /* comment */
		if l.ch == '/' && l.peekChar() == '*' {
			l.readChar()
			l.readChar()
			for l.ch != 0 {
				if l.ch == '*' && l.peekChar() == '/' {
					l.readChar()
					l.readChar()
					break
				}
				l.readChar()
			}
			continue
		}
		break
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == ':' {
		l.readChar()
	}
	return string(l.input[position:l.position])
}

func isQualified(ident string) bool {
	for i := 0; i < len(ident)-1; i++ {
		if ident[i] == ':' && ident[i+1] == ':' {
			return true
		}
	}
	return false
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readNumberOrUnit() string {
	position := l.position
	// Read number part
	for isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	// Check if unit syntax [km] follows immediately
	if l.ch == '[' {
		l.readChar()
		for l.ch != ']' && l.ch != 0 {
			l.readChar()
		}
		if l.ch == ']' {
			l.readChar() // consume ']'
		}
	}
	return string(l.input[position:l.position])
}

