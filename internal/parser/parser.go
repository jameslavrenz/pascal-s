package parser

import (
	"fmt"
	"strings"

	"github.com/JimLavrenz/pascal-s/internal/ast"
	"github.com/JimLavrenz/pascal-s/internal/lexer"
	"github.com/JimLavrenz/pascal-s/internal/token"
)

type Parser struct {
	l   *lexer.Lexer
	cur token.Token
}

func ParseProgram(input string) (*ast.Program, error) {
	p := &Parser{l: lexer.New(input)}
	p.next()
	return p.parseProgram()
}

func ParseREPL(input string) (*ast.Block, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil, nil
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "program ") {
		prog, err := ParseProgram(input)
		if err != nil {
			return nil, err
		}
		return prog.Block, nil
	}
	p := &Parser{l: lexer.New(input)}
	p.next()
	if p.cur.Kind == token.Const || p.cur.Kind == token.Type || p.cur.Kind == token.Var ||
		p.cur.Kind == token.Procedure || p.cur.Kind == token.Function {
		return p.parseBlockDeclsOnly()
	}
	stmts, err := p.parseStatementSequence(token.EOF)
	if err != nil {
		return nil, err
	}
	return &ast.Block{Stmts: stmts}, nil
}

func (p *Parser) next() {
	p.cur = p.l.Next()
}

func (p *Parser) expect(k token.Kind) error {
	if p.cur.Kind != k {
		return fmt.Errorf("%s: expected %s, got %s", lexer.FormatError(p.cur.Pos, ""), k, p.cur.Kind)
	}
	p.next()
	return nil
}

func (p *Parser) parseProgram() (*ast.Program, error) {
	if err := p.expect(token.Program); err != nil {
		return nil, err
	}
	if p.cur.Kind != token.Ident {
		return nil, fmt.Errorf("%s: expected program name", lexer.FormatError(p.cur.Pos, ""))
	}
	name := p.cur.Lit
	p.next()
	if err := p.expect(token.LParen); err != nil {
		return nil, err
	}
	var params []string
	for p.cur.Kind == token.Ident {
		params = append(params, p.cur.Lit)
		p.next()
		if p.cur.Kind == token.Comma {
			p.next()
		}
	}
	if err := p.expect(token.RParen); err != nil {
		return nil, err
	}
	if err := p.expect(token.Semicolon); err != nil {
		return nil, err
	}
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.Period); err != nil {
		return nil, err
	}
	return &ast.Program{Name: name, Params: params, Block: block}, nil
}

func (p *Parser) parseBlockDeclsOnly() (*ast.Block, error) {
	block := &ast.Block{}
	for p.cur.Kind == token.Const || p.cur.Kind == token.Type || p.cur.Kind == token.Var ||
		p.cur.Kind == token.Procedure || p.cur.Kind == token.Function {
		if err := p.parseDeclPart(block); err != nil {
			return nil, err
		}
	}
	return block, nil
}

func (p *Parser) parseBlock() (*ast.Block, error) {
	block := &ast.Block{}
	for p.cur.Kind == token.Const || p.cur.Kind == token.Type || p.cur.Kind == token.Var ||
		p.cur.Kind == token.Procedure || p.cur.Kind == token.Function {
		if err := p.parseDeclPart(block); err != nil {
			return nil, err
		}
	}
	if err := p.expect(token.Begin); err != nil {
		return nil, err
	}
	stmts, err := p.parseStatementSequence(token.End)
	if err != nil {
		return nil, err
	}
	block.Stmts = stmts
	if err := p.expect(token.End); err != nil {
		return nil, err
	}
	return block, nil
}

func (p *Parser) parseDeclPart(block *ast.Block) error {
	switch p.cur.Kind {
	case token.Const:
		p.next()
		for p.cur.Kind == token.Ident {
			name := p.cur.Lit
			p.next()
			if err := p.expect(token.Equal); err != nil {
				return err
			}
			val, err := p.parseConstant()
			if err != nil {
				return err
			}
			block.Consts = append(block.Consts, ast.ConstDecl{Name: name, Value: val})
			if p.cur.Kind == token.Semicolon {
				p.next()
			}
		}
	case token.Type:
		p.next()
		for p.cur.Kind == token.Ident {
			name := p.cur.Lit
			p.next()
			if err := p.expect(token.Equal); err != nil {
				return err
			}
			typ, err := p.parseType()
			if err != nil {
				return err
			}
			block.Types = append(block.Types, ast.TypeDecl{Name: name, Type: typ})
			if p.cur.Kind == token.Semicolon {
				p.next()
			}
		}
	case token.Var:
		p.next()
		for p.cur.Kind == token.Ident {
			names, err := p.parseIdentList()
			if err != nil {
				return err
			}
			if err := p.expect(token.Colon); err != nil {
				return err
			}
			typ, err := p.parseType()
			if err != nil {
				return err
			}
			block.Vars = append(block.Vars, ast.VarDecl{Names: names, Type: typ})
			if p.cur.Kind == token.Semicolon {
				p.next()
			}
		}
	case token.Procedure:
		proc, err := p.parseProcedure()
		if err != nil {
			return err
		}
		block.Procs = append(block.Procs, *proc)
		if p.cur.Kind == token.Semicolon {
			p.next()
		}
	case token.Function:
		fn, err := p.parseFunction()
		if err != nil {
			return err
		}
		block.Funcs = append(block.Funcs, *fn)
		if p.cur.Kind == token.Semicolon {
			p.next()
		}
	}
	return nil
}

func (p *Parser) parseIdentList() ([]string, error) {
	var names []string
	if p.cur.Kind != token.Ident {
		return nil, fmt.Errorf("%s: expected identifier", lexer.FormatError(p.cur.Pos, ""))
	}
	names = append(names, p.cur.Lit)
	p.next()
	for p.cur.Kind == token.Comma {
		p.next()
		if p.cur.Kind != token.Ident {
			return nil, fmt.Errorf("%s: expected identifier", lexer.FormatError(p.cur.Pos, ""))
		}
		names = append(names, p.cur.Lit)
		p.next()
	}
	return names, nil
}

func (p *Parser) parseType() (*ast.TypeDesc, error) {
	switch p.cur.Kind {
	case token.Ident:
		name := p.cur.Lit
		p.next()
		return &ast.TypeDesc{Kind: ast.TypeNamed, Name: name}, nil
	case token.Array:
		p.next()
		if err := p.expect(token.LBracket); err != nil {
			return nil, err
		}
		low, err := p.parseConstant()
		if err != nil {
			return nil, err
		}
		if err := p.expect(token.Range); err != nil {
			return nil, err
		}
		high, err := p.parseConstant()
		if err != nil {
			return nil, err
		}
		if err := p.expect(token.RBracket); err != nil {
			return nil, err
		}
		if err := p.expect(token.Of); err != nil {
			return nil, err
		}
		elem, err := p.parseType()
		if err != nil {
			return nil, err
		}
		return &ast.TypeDesc{
			Kind: ast.TypeArray,
			Low:  constInt(low),
			High: constInt(high),
			Elem: elem,
		}, nil
	case token.Record:
		p.next()
		var fields []ast.FieldDesc
		for p.cur.Kind == token.Ident {
			names, err := p.parseIdentList()
			if err != nil {
				return nil, err
			}
			if err := p.expect(token.Colon); err != nil {
				return nil, err
			}
			typ, err := p.parseType()
			if err != nil {
				return nil, err
			}
			fields = append(fields, ast.FieldDesc{Names: names, Type: typ})
			if p.cur.Kind == token.Semicolon {
				p.next()
			}
		}
		if err := p.expect(token.End); err != nil {
			return nil, err
		}
		return &ast.TypeDesc{Kind: ast.TypeRecord, Fields: fields}, nil
	default:
		return nil, fmt.Errorf("%s: expected type", lexer.FormatError(p.cur.Pos, ""))
	}
}

func constInt(e ast.Expr) int {
	switch v := e.(type) {
	case *ast.IntLit:
		return v.Value
	case *ast.CharLit:
		return int(v.Value)
	default:
		return 0
	}
}

func (p *Parser) parseConstant() (ast.Expr, error) {
	if p.cur.Kind == token.Plus || p.cur.Kind == token.Minus {
		op := p.cur.Lit
		p.next()
		expr, err := p.parseConstant()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: op, Expr: expr}, nil
	}
	if p.cur.Kind == token.Ident {
		name := p.cur.Lit
		p.next()
		return &ast.Ident{Name: name}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parseProcedure() (*ast.ProcDecl, error) {
	p.next()
	if p.cur.Kind != token.Ident {
		return nil, fmt.Errorf("%s: expected procedure name", lexer.FormatError(p.cur.Pos, ""))
	}
	name := p.cur.Lit
	p.next()
	params, err := p.parseFormalParams()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.Semicolon); err != nil {
		return nil, err
	}
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.ProcDecl{Name: name, Params: params, Block: block}, nil
}

func (p *Parser) parseFunction() (*ast.FuncDecl, error) {
	p.next()
	if p.cur.Kind != token.Ident {
		return nil, fmt.Errorf("%s: expected function name", lexer.FormatError(p.cur.Pos, ""))
	}
	name := p.cur.Lit
	p.next()
	params, err := p.parseFormalParams()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.Colon); err != nil {
		return nil, err
	}
	result, err := p.parseType()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.Semicolon); err != nil {
		return nil, err
	}
	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.FuncDecl{Name: name, Params: params, ResultType: result, Block: block}, nil
}

func (p *Parser) parseFormalParams() ([]ast.ParamSection, error) {
	if p.cur.Kind != token.LParen {
		return nil, nil
	}
	p.next()
	var sections []ast.ParamSection
	for p.cur.Kind == token.Ident || p.cur.Kind == token.Var {
		sec := ast.ParamSection{}
		if p.cur.Kind == token.Var {
			sec.IsVar = true
			p.next()
		}
		names, err := p.parseIdentList()
		if err != nil {
			return nil, err
		}
		sec.Names = names
		if err := p.expect(token.Colon); err != nil {
			return nil, err
		}
		typ, err := p.parseType()
		if err != nil {
			return nil, err
		}
		sec.Type = typ
		sections = append(sections, sec)
		if p.cur.Kind == token.Semicolon {
			p.next()
		}
	}
	if err := p.expect(token.RParen); err != nil {
		return nil, err
	}
	return sections, nil
}

func (p *Parser) parseStatementSequence(end token.Kind) ([]ast.Stmt, error) {
	var stmts []ast.Stmt
	for p.cur.Kind != end && p.cur.Kind != token.EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		if p.cur.Kind == token.Semicolon {
			p.next()
		}
	}
	return stmts, nil
}

func (p *Parser) parseStatement() (ast.Stmt, error) {
	switch p.cur.Kind {
	case token.Begin:
		return p.parseCompound()
	case token.If:
		return p.parseIf()
	case token.While:
		return p.parseWhile()
	case token.Repeat:
		return p.parseRepeat()
	case token.For:
		return p.parseFor()
	case token.Case:
		return p.parseCase()
	case token.Ident:
		return p.parseSimpleStatement()
	default:
		return nil, fmt.Errorf("%s: unexpected token %s", lexer.FormatError(p.cur.Pos, ""), p.cur.Kind)
	}
}

func (p *Parser) parseSimpleStatement() (ast.Stmt, error) {
	name := p.cur.Lit
	p.next()
	expr, err := p.parseDesignatorSuffix(&ast.Ident{Name: name})
	if err != nil {
		return nil, err
	}
	if p.cur.Kind == token.Assign {
		p.next()
		val, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		return &ast.AssignStmt{Target: expr, Value: val}, nil
	}
	if call, ok := expr.(*ast.CallExpr); ok {
		return &ast.ProcCallStmt{Name: call.Name, Args: call.Args}, nil
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return &ast.ProcCallStmt{Name: ident.Name, Args: nil}, nil
	}
	return nil, fmt.Errorf("%s: expected assignment or procedure call", lexer.FormatError(p.cur.Pos, ""))
}

func (p *Parser) parseDesignatorSuffix(base ast.Expr) (ast.Expr, error) {
	expr := base
	for {
		switch p.cur.Kind {
		case token.LBracket:
			p.next()
			idx, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			if err := p.expect(token.RBracket); err != nil {
				return nil, err
			}
			expr = &ast.IndexExpr{Array: expr, Index: idx}
		case token.Dot:
			p.next()
			if p.cur.Kind != token.Ident {
				return nil, fmt.Errorf("%s: expected field name", lexer.FormatError(p.cur.Pos, ""))
			}
			expr = &ast.FieldExpr{Record: expr, Field: p.cur.Lit}
			p.next()
		case token.LParen:
			name := ""
			if ident, ok := expr.(*ast.Ident); ok {
				name = ident.Name
			} else if call, ok := expr.(*ast.CallExpr); ok {
				name = call.Name
			} else {
				return nil, fmt.Errorf("%s: invalid call target", lexer.FormatError(p.cur.Pos, ""))
			}
			p.next()
			var args []ast.Expr
			if p.cur.Kind != token.RParen {
				for {
					arg, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
					if p.cur.Kind != token.Comma {
						break
					}
					p.next()
				}
			}
			if err := p.expect(token.RParen); err != nil {
				return nil, err
			}
			expr = &ast.CallExpr{Name: name, Args: args}
		default:
			return expr, nil
		}
	}
}

func (p *Parser) parseCompound() (ast.Stmt, error) {
	p.next()
	stmts, err := p.parseStatementSequence(token.End)
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.End); err != nil {
		return nil, err
	}
	return &ast.CompoundStmt{Stmts: stmts}, nil
}

func (p *Parser) parseIf() (ast.Stmt, error) {
	p.next()
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.Then); err != nil {
		return nil, err
	}
	thenStmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	var elseStmt ast.Stmt
	if p.cur.Kind == token.Else {
		p.next()
		elseStmt, err = p.parseStatement()
		if err != nil {
			return nil, err
		}
	}
	return &ast.IfStmt{Cond: cond, Then: thenStmt, Else: elseStmt}, nil
}

func (p *Parser) parseWhile() (ast.Stmt, error) {
	p.next()
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.Do); err != nil {
		return nil, err
	}
	body, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	return &ast.WhileStmt{Cond: cond, Body: body}, nil
}

func (p *Parser) parseRepeat() (ast.Stmt, error) {
	p.next()
	body, err := p.parseStatementSequence(token.Until)
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.Until); err != nil {
		return nil, err
	}
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return &ast.RepeatStmt{Body: body, Cond: cond}, nil
}

func (p *Parser) parseFor() (ast.Stmt, error) {
	p.next()
	if p.cur.Kind != token.Ident {
		return nil, fmt.Errorf("%s: expected control variable", lexer.FormatError(p.cur.Pos, ""))
	}
	v := p.cur.Lit
	p.next()
	if err := p.expect(token.Assign); err != nil {
		return nil, err
	}
	start, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	down := false
	if p.cur.Kind == token.To {
		p.next()
	} else if p.cur.Kind == token.Downto {
		down = true
		p.next()
	} else {
		return nil, fmt.Errorf("%s: expected to or downto", lexer.FormatError(p.cur.Pos, ""))
	}
	end, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.Do); err != nil {
		return nil, err
	}
	body, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	return &ast.ForStmt{Var: v, Start: start, End: end, Down: down, Body: body}, nil
}

func (p *Parser) parseCase() (ast.Stmt, error) {
	p.next()
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(token.Of); err != nil {
		return nil, err
	}
	var branches []ast.CaseBranch
	for p.cur.Kind != token.End && p.cur.Kind != token.EOF {
		labels, err := p.parseCaseLabels()
		if err != nil {
			return nil, err
		}
		if err := p.expect(token.Colon); err != nil {
			return nil, err
		}
		body, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		branches = append(branches, ast.CaseBranch{Labels: labels, Body: body})
		if p.cur.Kind == token.Semicolon {
			p.next()
		}
	}
	if err := p.expect(token.End); err != nil {
		return nil, err
	}
	return &ast.CaseStmt{Expr: expr, Cases: branches}, nil
}

func (p *Parser) parseCaseLabels() ([]ast.Expr, error) {
	var labels []ast.Expr
	for {
		c, err := p.parseConstant()
		if err != nil {
			return nil, err
		}
		labels = append(labels, c)
		if p.cur.Kind != token.Comma {
			break
		}
		p.next()
	}
	return labels, nil
}

func (p *Parser) parseExpression() (ast.Expr, error) {
	left, err := p.parseSimpleExpression()
	if err != nil {
		return nil, err
	}
	if p.cur.Kind >= token.Equal && p.cur.Kind <= token.Ge {
		op := p.cur.Lit
		p.next()
		right, err := p.parseSimpleExpression()
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpr{Op: op, Left: left, Right: right}, nil
	}
	return left, nil
}

func (p *Parser) parseSimpleExpression() (ast.Expr, error) {
	var left ast.Expr
	var err error
	if p.cur.Kind == token.Plus || p.cur.Kind == token.Minus {
		op := p.cur.Lit
		p.next()
		term, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &ast.UnaryExpr{Op: op, Expr: term}
	} else {
		left, err = p.parseTerm()
		if err != nil {
			return nil, err
		}
	}
	for p.cur.Kind == token.Plus || p.cur.Kind == token.Minus || p.cur.Kind == token.Or {
		op := p.cur.Lit
		p.next()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseTerm() (ast.Expr, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for p.cur.Kind == token.Star || p.cur.Kind == token.Slash ||
		p.cur.Kind == token.Div || p.cur.Kind == token.Mod || p.cur.Kind == token.And {
		op := p.cur.Lit
		p.next()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseFactor() (ast.Expr, error) {
	if p.cur.Kind == token.Not {
		p.next()
		expr, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: "not", Expr: expr}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (ast.Expr, error) {
	switch p.cur.Kind {
	case token.IntLit:
		v, _ := lexer.ParseInt(p.cur.Lit)
		p.next()
		return &ast.IntLit{Value: v}, nil
	case token.RealLit:
		v, _ := lexer.ParseReal(p.cur.Lit)
		p.next()
		return &ast.RealLit{Value: v}, nil
	case token.CharLit:
		ch := byte(0)
		if len(p.cur.Lit) > 0 {
			ch = p.cur.Lit[0]
		}
		p.next()
		return &ast.CharLit{Value: ch}, nil
	case token.StringLit:
		s := p.cur.Lit
		p.next()
		return &ast.StringLit{Value: s}, nil
	case token.Ident:
		if p.cur.Lit == "true" {
			p.next()
			return &ast.BoolLit{Value: true}, nil
		}
		if p.cur.Lit == "false" {
			p.next()
			return &ast.BoolLit{Value: false}, nil
		}
		name := p.cur.Lit
		p.next()
		return p.parseDesignatorSuffix(&ast.Ident{Name: name})
	case token.LParen:
		p.next()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.expect(token.RParen); err != nil {
			return nil, err
		}
		return expr, nil
	default:
		return nil, fmt.Errorf("%s: expected expression", lexer.FormatError(p.cur.Pos, ""))
	}
}
