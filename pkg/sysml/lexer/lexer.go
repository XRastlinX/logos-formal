package lexer

import (
	"github.com/XRastlinX/logos-formal/pkg/sysml/token"
)

type Lexer struct {
	input        []byte
	position     int
	readPosition int
	ch           byte

	line       int
	column     int
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

	if errTok, ok := l.skipWhitespaceAndComments(); ok {
		return errTok
	}

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
			literal, tokType := l.readNumberOrUnit()
			tok.Type = tokType
			tok.Literal = literal
			tok.Start = startByte
			tok.End = l.position
			tok.Line = startLine
			tok.Column = startColumn
			return tok
		} else if isLetter(l.ch) {
			literal, tokType := l.readIdentifier()
			tok.Type = tokType
			tok.Literal = literal
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

func (l *Lexer) skipWhitespaceAndComments() (token.Token, bool) {
	for {
		if l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			l.readChar()
			continue
		}
		if l.ch == '/' && l.peekChar() == '/' {
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			continue
		}
		if l.ch == '/' && l.peekChar() == '*' {
			startByte := l.position
			startLine := l.line
			startCol := l.column
			l.readChar()
			l.readChar()
			terminated := false
			for l.ch != 0 {
				if l.ch == '*' && l.peekChar() == '/' {
					l.readChar()
					l.readChar()
					terminated = true
					break
				}
				l.readChar()
			}
			if !terminated {
				return token.Token{
					Type:    token.ILLEGAL,
					Literal: "unterminated block comment",
					Line:    startLine,
					Column:  startCol,
					Start:   startByte,
					End:     l.position,
				}, true
			}
			continue
		}
		break
	}
	return token.Token{}, false
}

func (l *Lexer) readIdentifier() (string, token.TokenType) {
	position := l.position
	hasColons := false

	for {
		for isLetter(l.ch) || isDigit(l.ch) {
			l.readChar()
		}
		
		// If we see ::, we might have a qualified identifier
		if l.ch == ':' && l.peekChar() == ':' {
			hasColons = true
			l.readChar() // consume :
			l.readChar() // consume :
			// After :: we MUST have a letter
			if !isLetter(l.ch) {
				// If not a letter, this is malformed. Just return ILLEGAL
				for isLetter(l.ch) || isDigit(l.ch) || l.ch == ':' {
					l.readChar()
				}
				return string(l.input[position:l.position]), token.ILLEGAL
			}
			continue
		}
		break
	}

	literal := string(l.input[position:l.position])
	if hasColons {
		return literal, token.QIDENT
	}
	return literal, token.LookupIdent(literal)
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readNumberOrUnit() (string, token.TokenType) {
	position := l.position
	dotCount := 0

	for isDigit(l.ch) || l.ch == '.' {
		if l.ch == '.' {
			dotCount++
		}
		l.readChar()
	}

	if dotCount > 1 {
		return string(l.input[position:l.position]), token.ILLEGAL
	}

	if l.ch == '[' {
		l.readChar()
		for l.ch != ']' && l.ch != 0 && l.ch != '\n' {
			l.readChar()
		}
		if l.ch == ']' {
			l.readChar() // consume ']'
			return string(l.input[position:l.position]), token.UNIT
		}
		// Unterminated unit
		return string(l.input[position:l.position]), token.ILLEGAL
	}

	return string(l.input[position:l.position]), token.NUMBER
}

