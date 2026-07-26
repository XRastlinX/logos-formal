package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
	Start   int
	End     int
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers and literals
	IDENT   = "IDENT"
	QIDENT  = "QIDENT" // Qualified identifier e.g. ISQ::length
	NUMBER  = "NUMBER" // e.g. 0, 0.5, 10
	UNIT    = "UNIT"   // e.g. 10[km]

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"
	EQ       = "=="
	NOT_EQ   = "!="
	GT       = ">"
	LT       = "<"
	GTE      = ">="
	LTE      = "<="
	COLON_GT = ":>" // Directed attribute

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"

	// Keywords
	PACKAGE    = "PACKAGE"
	PRIVATE    = "PRIVATE"
	IMPORT     = "IMPORT"
	CALC       = "CALC"
	DEF        = "DEF"
	CONSTRAINT = "CONSTRAINT"
	IN         = "IN"
	ATTRIBUTE  = "ATTRIBUTE"
	RETURN     = "RETURN"
	AND        = "AND"
)

var keywords = map[string]TokenType{
	"package":    PACKAGE,
	"private":    PRIVATE,
	"import":     IMPORT,
	"calc":       CALC,
	"def":        DEF,
	"constraint": CONSTRAINT,
	"in":         IN,
	"attribute":  ATTRIBUTE,
	"return":     RETURN,
	"and":        AND,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
