package token

type Kind int

const (
	EOF Kind = iota
	Ident
	IntLit
	RealLit
	CharLit
	StringLit

	// keywords
	Program
	Array
	Begin
	Case
	Const
	Div
	Do
	Downto
	Else
	End
	For
	Function
	If
	Mod
	Not
	Of
	Or
	Procedure
	Record
	Repeat
	Then
	To
	Type
	Until
	Var
	While
	And

	// symbols
	Assign   // :=
	Equal    // =
	Ne       // <>
	Lt
	Le
	Gt
	Ge
	Plus
	Minus
	Star
	Slash
	Dot
	Comma
	Semicolon
	Colon
	LParen
	RParen
	LBracket
	RBracket
	Range    // ..
	Period   // .
)

var keywordMap = map[string]Kind{
	"program":   Program,
	"array":     Array,
	"begin":     Begin,
	"case":      Case,
	"const":     Const,
	"div":       Div,
	"do":        Do,
	"downto":    Downto,
	"else":      Else,
	"end":       End,
	"for":       For,
	"function":  Function,
	"if":        If,
	"mod":       Mod,
	"not":       Not,
	"of":        Of,
	"or":        Or,
	"procedure": Procedure,
	"record":    Record,
	"repeat":    Repeat,
	"then":      Then,
	"to":        To,
	"type":      Type,
	"until":     Until,
	"var":       Var,
	"while":     While,
	"and":       And,
}

func LookupKeyword(word string) (Kind, bool) {
	k, ok := keywordMap[word]
	return k, ok
}

type Token struct {
	Kind Kind
	Lit  string
	Pos  int
}

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "unknown"
}

var kindNames = []string{
	"EOF", "ident", "int", "real", "char", "string",
	"program", "array", "begin", "case", "const", "div", "do", "downto",
	"else", "end", "for", "function", "if", "mod", "not", "of", "or",
	"procedure", "record", "repeat", "then", "to", "type", "until", "var",
	"while", "and",
	":=", "=", "<>", "<", "<=", ">", ">=",
	"+", "-", "*", "/", ".", ",", ";", ":", "(", ")", "[", "]", "..", ".",
}
