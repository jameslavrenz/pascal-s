package lexer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/JimLavrenz/pascal-s/internal/token"
)

type Lexer struct {
	input string
	pos   int
}

func New(input string) *Lexer {
	return &Lexer{input: input}
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) next() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	return ch
}

func (l *Lexer) skipWhitespace() {
	for {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			l.next()
			continue
		}
		break
	}
}

func (l *Lexer) skipComment() bool {
	if l.peek() == '{' {
		l.next()
		for l.peek() != 0 && l.peek() != '}' {
			l.next()
		}
		if l.peek() == '}' {
			l.next()
		}
		return true
	}
	if l.peek() == '(' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '*' {
		l.next()
		l.next()
		for l.peek() != 0 {
			if l.peek() == '*' && l.pos+1 < len(l.input) && l.input[l.pos+1] == ')' {
				l.next()
				l.next()
				break
			}
			l.next()
		}
		return true
	}
	return false
}

func (l *Lexer) Next() token.Token {
	for {
		l.skipWhitespace()
		for l.skipComment() {
			l.skipWhitespace()
		}

		start := l.pos
		ch := l.peek()
		if ch == 0 {
			return token.Token{Kind: token.EOF, Pos: start}
		}

		switch ch {
		case '+':
			l.next()
			return token.Token{Kind: token.Plus, Lit: "+", Pos: start}
		case '-':
			l.next()
			return token.Token{Kind: token.Minus, Lit: "-", Pos: start}
		case '*':
			l.next()
			return token.Token{Kind: token.Star, Lit: "*", Pos: start}
		case '/':
			l.next()
			return token.Token{Kind: token.Slash, Lit: "/", Pos: start}
		case '(':
			l.next()
			return token.Token{Kind: token.LParen, Lit: "(", Pos: start}
		case ')':
			l.next()
			return token.Token{Kind: token.RParen, Lit: ")", Pos: start}
		case '[':
			l.next()
			return token.Token{Kind: token.LBracket, Lit: "[", Pos: start}
		case ']':
			l.next()
			return token.Token{Kind: token.RBracket, Lit: "]", Pos: start}
		case ',':
			l.next()
			return token.Token{Kind: token.Comma, Lit: ",", Pos: start}
		case ';':
			l.next()
			return token.Token{Kind: token.Semicolon, Lit: ";", Pos: start}
		case ':':
			l.next()
			if l.peek() == '=' {
				l.next()
				return token.Token{Kind: token.Assign, Lit: ":=", Pos: start}
			}
			return token.Token{Kind: token.Colon, Lit: ":", Pos: start}
		case '.':
			l.next()
			if l.peek() == '.' {
				l.next()
				return token.Token{Kind: token.Range, Lit: "..", Pos: start}
			}
			return token.Token{Kind: token.Period, Lit: ".", Pos: start}
		case '=':
			l.next()
			return token.Token{Kind: token.Equal, Lit: "=", Pos: start}
		case '<':
			l.next()
			if l.peek() == '>' {
				l.next()
				return token.Token{Kind: token.Ne, Lit: "<>", Pos: start}
			}
			if l.peek() == '=' {
				l.next()
				return token.Token{Kind: token.Le, Lit: "<=", Pos: start}
			}
			return token.Token{Kind: token.Lt, Lit: "<", Pos: start}
		case '>':
			l.next()
			if l.peek() == '=' {
				l.next()
				return token.Token{Kind: token.Ge, Lit: ">=", Pos: start}
			}
			return token.Token{Kind: token.Gt, Lit: ">", Pos: start}
		case '\'':
			return l.readSingleQuoted(start)
		case '"':
			return l.readStringLit(start)
		default:
			if ch >= '0' && ch <= '9' {
				return l.readNumber(start)
			}
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
				return l.readIdent(start)
			}
			l.next()
			return token.Token{Kind: token.Ident, Lit: string(ch), Pos: start}
		}
	}
}

func (l *Lexer) readIdent(start int) token.Token {
	for {
		ch := l.peek()
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' {
			l.next()
			continue
		}
		break
	}
	lit := strings.ToLower(l.input[start:l.pos])
	if k, ok := token.LookupKeyword(lit); ok {
		return token.Token{Kind: k, Lit: lit, Pos: start}
	}
	return token.Token{Kind: token.Ident, Lit: lit, Pos: start}
}

func (l *Lexer) readNumber(start int) token.Token {
	isReal := false
	for unicode.IsDigit(rune(l.peek())) {
		l.next()
	}
	if l.peek() == '.' && l.pos+1 < len(l.input) && l.input[l.pos+1] >= '0' && l.input[l.pos+1] <= '9' {
		isReal = true
		l.next()
		for unicode.IsDigit(rune(l.peek())) {
			l.next()
		}
	}
	lit := l.input[start:l.pos]
	if isReal {
		return token.Token{Kind: token.RealLit, Lit: lit, Pos: start}
	}
	return token.Token{Kind: token.IntLit, Lit: lit, Pos: start}
}

func (l *Lexer) readSingleQuoted(start int) token.Token {
	l.next()
	var b strings.Builder
	for l.peek() != 0 {
		if l.peek() == '\'' {
			if l.pos+1 < len(l.input) && l.input[l.pos+1] == '\'' {
				b.WriteByte('\'')
				l.next()
				l.next()
				continue
			}
			l.next()
			break
		}
		b.WriteByte(l.next())
	}
	lit := b.String()
	if len(lit) == 1 {
		return token.Token{Kind: token.CharLit, Lit: lit, Pos: start}
	}
	return token.Token{Kind: token.StringLit, Lit: lit, Pos: start}
}

func (l *Lexer) readStringLit(start int) token.Token {
	l.next()
	begin := l.pos
	for l.peek() != 0 && l.peek() != '"' {
		l.next()
	}
	lit := l.input[begin:l.pos]
	if l.peek() == '"' {
		l.next()
	}
	return token.Token{Kind: token.StringLit, Lit: lit, Pos: start}
}

func ParseInt(lit string) (int, error) {
	return strconv.Atoi(lit)
}

func ParseReal(lit string) (float64, error) {
	return strconv.ParseFloat(lit, 64)
}

func FormatError(pos int, msg string) string {
	if msg == "" {
		return fmt.Sprintf("parse error at position %d", pos)
	}
	return fmt.Sprintf("parse error at position %d: %s", pos, msg)
}
