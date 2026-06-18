package runtime

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/JimLavrenz/pascal-s/internal/ast"
)

type ValueKind int

const (
	ValInt ValueKind = iota
	ValReal
	ValBool
	ValChar
	ValString
	ValArray
	ValRecord
)

type Value struct {
	Kind   ValueKind
	Int    int
	Real   float64
	Bool   bool
	Char   byte
	Str    string
	Array  []Value
	Fields map[string]Value
}

type Runtime struct {
	env       *Env
	input     *bufio.Reader
	output    io.Writer
	inputData []string
	inputIdx  int
}

type Env struct {
	parent      *Env
	vars        map[string]Value
	consts      map[string]Value
	types       map[string]*ast.TypeDesc
	arrayLow    map[string]int
	procs       map[string]*ast.ProcDecl
	funcs       map[string]*ast.FuncDecl
	funcName    string
}

func New() *Runtime {
	return &Runtime{
		env:    newEnv(nil),
		input:  bufio.NewReader(os.Stdin),
		output: os.Stdout,
	}
}

func (r *Runtime) SetInput(reader io.Reader) {
	r.input = bufio.NewReader(reader)
}

func (r *Runtime) SetOutput(w io.Writer) {
	r.output = w
}

func (r *Runtime) SetInputLines(lines []string) {
	r.inputData = lines
	r.inputIdx = 0
}

func newEnv(parent *Env) *Env {
	return &Env{
		parent:   parent,
		vars:     make(map[string]Value),
		consts:   make(map[string]Value),
		types:    make(map[string]*ast.TypeDesc),
		arrayLow: make(map[string]int),
		procs:    make(map[string]*ast.ProcDecl),
		funcs:    make(map[string]*ast.FuncDecl),
	}
}

func (r *Runtime) RunProgram(prog *ast.Program) error {
	r.installBuiltins()
	r.declareBlock(prog.Block)
	return r.execBlock(prog.Block)
}

func (r *Runtime) RunBlock(block *ast.Block) error {
	r.installBuiltins()
	r.declareBlock(block)
	return r.execBlock(block)
}

func (r *Runtime) installBuiltins() {
	r.env.consts["true"] = Value{Kind: ValBool, Bool: true}
	r.env.consts["false"] = Value{Kind: ValBool, Bool: false}
	r.env.types["integer"] = &ast.TypeDesc{Kind: ast.TypeInteger}
	r.env.types["real"] = &ast.TypeDesc{Kind: ast.TypeReal}
	r.env.types["boolean"] = &ast.TypeDesc{Kind: ast.TypeBoolean}
	r.env.types["char"] = &ast.TypeDesc{Kind: ast.TypeChar}
}

func (r *Runtime) declareBlock(block *ast.Block) {
	for _, c := range block.Consts {
		val, err := r.evalExpr(c.Value)
		if err != nil {
			panic(err)
		}
		r.env.consts[c.Name] = val
	}
	for _, t := range block.Types {
		r.env.types[t.Name] = t.Type
	}
	for _, v := range block.Vars {
		typ := r.resolveType(v.Type)
		for _, name := range v.Names {
			r.env.vars[name] = r.defaultValue(typ)
			if typ.Kind == ast.TypeArray {
				r.env.arrayLow[name] = typ.Low
			}
		}
	}
	for i := range block.Procs {
		p := block.Procs[i]
		r.env.procs[p.Name] = &p
	}
	for i := range block.Funcs {
		f := block.Funcs[i]
		r.env.funcs[f.Name] = &f
	}
}

func (r *Runtime) defaultValue(desc *ast.TypeDesc) Value {
	desc = r.resolveType(desc)
	switch desc.Kind {
	case ast.TypeInteger:
		return Value{Kind: ValInt}
	case ast.TypeReal:
		return Value{Kind: ValReal}
	case ast.TypeBoolean:
		return Value{Kind: ValBool}
	case ast.TypeChar:
		return Value{Kind: ValChar}
	case ast.TypeArray:
		size := desc.High - desc.Low + 1
		arr := make([]Value, size)
		for i := range arr {
			arr[i] = r.defaultValue(desc.Elem)
		}
		return Value{Kind: ValArray, Array: arr}
	case ast.TypeRecord:
		fields := make(map[string]Value)
		for _, f := range desc.Fields {
			for _, name := range f.Names {
				fields[name] = r.defaultValue(f.Type)
			}
		}
		return Value{Kind: ValRecord, Fields: fields}
	default:
		return Value{Kind: ValInt}
	}
}

func (r *Runtime) resolveType(desc *ast.TypeDesc) *ast.TypeDesc {
	if desc.Kind == ast.TypeNamed {
		if t, ok := r.env.types[desc.Name]; ok {
			return r.resolveType(t)
		}
		switch desc.Name {
		case "integer":
			return &ast.TypeDesc{Kind: ast.TypeInteger}
		case "real":
			return &ast.TypeDesc{Kind: ast.TypeReal}
		case "boolean":
			return &ast.TypeDesc{Kind: ast.TypeBoolean}
		case "char":
			return &ast.TypeDesc{Kind: ast.TypeChar}
		}
	}
	return desc
}

func (r *Runtime) execBlock(block *ast.Block) error {
	for _, s := range block.Stmts {
		if err := r.execStmt(s); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runtime) execStmt(s ast.Stmt) error {
	switch s := s.(type) {
	case *ast.AssignStmt:
		val, err := r.evalExpr(s.Value)
		if err != nil {
			return err
		}
		return r.assign(s.Target, val)
	case *ast.ProcCallStmt:
		return r.callProc(s.Name, s.Args)
	case *ast.CompoundStmt:
		for _, st := range s.Stmts {
			if err := r.execStmt(st); err != nil {
				return err
			}
		}
		return nil
	case *ast.IfStmt:
		cond, err := r.evalExpr(s.Cond)
		if err != nil {
			return err
		}
		if toBool(cond) {
			return r.execStmt(s.Then)
		}
		if s.Else != nil {
			return r.execStmt(s.Else)
		}
		return nil
	case *ast.WhileStmt:
		for {
			cond, err := r.evalExpr(s.Cond)
			if err != nil {
				return err
			}
			if !toBool(cond) {
				break
			}
			if err := r.execStmt(s.Body); err != nil {
				return err
			}
		}
		return nil
	case *ast.RepeatStmt:
		for {
			for _, st := range s.Body {
				if err := r.execStmt(st); err != nil {
					return err
				}
			}
			cond, err := r.evalExpr(s.Cond)
			if err != nil {
				return err
			}
			if toBool(cond) {
				break
			}
		}
		return nil
	case *ast.ForStmt:
		start, err := r.evalExpr(s.Start)
		if err != nil {
			return err
		}
		end, err := r.evalExpr(s.End)
		if err != nil {
			return err
		}
		iStart := toInt(start)
		iEnd := toInt(end)
		if s.Down {
			for i := iStart; i >= iEnd; i-- {
				r.env.vars[s.Var] = Value{Kind: ValInt, Int: i}
				if err := r.execStmt(s.Body); err != nil {
					return err
				}
			}
		} else {
			for i := iStart; i <= iEnd; i++ {
				r.env.vars[s.Var] = Value{Kind: ValInt, Int: i}
				if err := r.execStmt(s.Body); err != nil {
					return err
				}
			}
		}
		return nil
	case *ast.CaseStmt:
		val, err := r.evalExpr(s.Expr)
		if err != nil {
			return err
		}
		for _, branch := range s.Cases {
			for _, label := range branch.Labels {
				lv, err := r.evalExpr(label)
				if err != nil {
					return err
				}
				if valuesEqual(val, lv) {
					return r.execStmt(branch.Body)
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown statement")
	}
}

func (r *Runtime) assign(target ast.Expr, val Value) error {
	switch t := target.(type) {
	case *ast.Ident:
		if r.env.funcName != "" && t.Name == r.env.funcName {
			r.env.vars[t.Name] = val
			return nil
		}
		r.env.vars[t.Name] = val
		return nil
	case *ast.IndexExpr:
		arrVal, err := r.evalExpr(t.Array)
		if err != nil {
			return err
		}
		idx, err := r.evalExpr(t.Index)
		if err != nil {
			return err
		}
		i := toInt(idx)
		if arrIdent, ok := t.Array.(*ast.Ident); ok {
			v := r.env.vars[arrIdent.Name]
			low := r.env.arrayLow[arrIdent.Name]
			if low == 0 {
				low = 1
			}
			pos := i - low
			if pos >= 0 && pos < len(v.Array) {
				v.Array[pos] = val
				r.env.vars[arrIdent.Name] = v
			}
		}
		_ = arrVal
		return nil
	case *ast.FieldExpr:
		if recIdent, ok := t.Record.(*ast.Ident); ok {
			v := r.env.vars[recIdent.Name]
			v.Fields[t.Field] = val
			r.env.vars[recIdent.Name] = v
		}
		return nil
	default:
		return fmt.Errorf("invalid assignment target")
	}
}

func arrayLow(r *Runtime, name string) int {
	if l, ok := r.env.arrayLow[name]; ok {
		return l
	}
	return 1
}

func (r *Runtime) lookupFunc(name string) (*ast.FuncDecl, bool) {
	for e := r.env; e != nil; e = e.parent {
		if f, ok := e.funcs[name]; ok {
			return f, true
		}
	}
	return nil, false
}

func (r *Runtime) lookupProc(name string) (*ast.ProcDecl, bool) {
	for e := r.env; e != nil; e = e.parent {
		if p, ok := e.procs[name]; ok {
			return p, true
		}
	}
	return nil, false
}

func (r *Runtime) callProc(name string, args []ast.Expr) error {
	switch name {
	case "write":
		return r.builtinWrite(args, false)
	case "writeln":
		return r.builtinWrite(args, true)
	case "read":
		return r.builtinRead(args, false)
	case "readln":
		return r.builtinRead(args, true)
	}
	if proc, ok := r.lookupProc(name); ok {
		return r.callUserProc(proc, args)
	}
	if _, ok := r.lookupFunc(name); ok {
		_, err := r.callUserFunc(name, args)
		return err
	}
	return fmt.Errorf("unknown procedure: %s", name)
}

func (r *Runtime) lookupVar(name string) (Value, bool) {
	for e := r.env; e != nil; e = e.parent {
		if v, ok := e.vars[name]; ok {
			return v, true
		}
	}
	return Value{}, false
}

func (r *Runtime) lookupConst(name string) (Value, bool) {
	for e := r.env; e != nil; e = e.parent {
		if v, ok := e.consts[name]; ok {
			return v, true
		}
	}
	return Value{}, false
}

func (r *Runtime) evalExpr(e ast.Expr) (Value, error) {
	switch e := e.(type) {
	case *ast.IntLit:
		return Value{Kind: ValInt, Int: e.Value}, nil
	case *ast.RealLit:
		return Value{Kind: ValReal, Real: e.Value}, nil
	case *ast.BoolLit:
		return Value{Kind: ValBool, Bool: e.Value}, nil
	case *ast.CharLit:
		return Value{Kind: ValChar, Char: e.Value}, nil
	case *ast.StringLit:
		return Value{Kind: ValString, Str: e.Value}, nil
	case *ast.Ident:
		if v, ok := r.lookupConst(e.Name); ok {
			return v, nil
		}
		if v, ok := r.lookupVar(e.Name); ok {
			return v, nil
		}
		return Value{}, fmt.Errorf("undefined identifier: %s", e.Name)
	case *ast.UnaryExpr:
		v, err := r.evalExpr(e.Expr)
		if err != nil {
			return Value{}, err
		}
		return r.evalUnary(e.Op, v)
	case *ast.BinaryExpr:
		left, err := r.evalExpr(e.Left)
		if err != nil {
			return Value{}, err
		}
		right, err := r.evalExpr(e.Right)
		if err != nil {
			return Value{}, err
		}
		return r.evalBinary(e.Op, left, right)
	case *ast.CallExpr:
		return r.evalCall(e)
	case *ast.IndexExpr:
		arr, err := r.evalExpr(e.Array)
		if err != nil {
			return Value{}, err
		}
		idx, err := r.evalExpr(e.Index)
		if err != nil {
			return Value{}, err
		}
		pos := toInt(idx)
		low := 1
		if arrIdent, ok := e.Array.(*ast.Ident); ok {
			if l, ok := r.env.arrayLow[arrIdent.Name]; ok {
				low = l
			}
		}
		i := pos - low
		if i >= 0 && i < len(arr.Array) {
			return arr.Array[i], nil
		}
		return Value{}, fmt.Errorf("array index out of bounds")
	case *ast.FieldExpr:
		rec, err := r.evalExpr(e.Record)
		if err != nil {
			return Value{}, err
		}
		if v, ok := rec.Fields[e.Field]; ok {
			return v, nil
		}
		return Value{}, fmt.Errorf("unknown field: %s", e.Field)
	default:
		return Value{}, fmt.Errorf("unknown expression")
	}
}

func (r *Runtime) evalCall(e *ast.CallExpr) (Value, error) {
	if v, ok := r.evalBuiltinFunc(e.Name, e.Args); ok {
		return v, nil
	}
	return r.callUserFunc(e.Name, e.Args)
}

func (r *Runtime) callUserProc(proc *ast.ProcDecl, args []ast.Expr) error {
	child := newEnv(r.env)
	old := r.env
	r.env = child
	r.declareBlock(proc.Block)
	for k, v := range old.funcs {
		r.env.funcs[k] = v
	}
	for k, v := range old.procs {
		r.env.procs[k] = v
	}
	r.env.funcName = ""

	argIdx := 0
	for _, sec := range proc.Params {
		for _, name := range sec.Names {
			if argIdx < len(args) {
				val, err := r.evalExpr(args[argIdx])
				if err != nil {
					r.env = old
					return err
				}
				child.vars[name] = val
				argIdx++
			}
		}
	}

	err := r.execBlock(proc.Block)
	r.env = old
	return err
}

func (r *Runtime) callUserFunc(name string, args []ast.Expr) (Value, error) {
	proc, ok := r.lookupFunc(name)
	if !ok {
		return Value{}, fmt.Errorf("unknown function: %s", name)
	}
	child := newEnv(r.env)
	old := r.env
	r.env = child
	r.declareBlock(proc.Block)
	// inherit callable definitions for recursion
	for k, v := range old.funcs {
		r.env.funcs[k] = v
	}
	for k, v := range old.procs {
		r.env.procs[k] = v
	}
	r.env.funcName = name
	r.env.vars[name] = r.defaultValue(proc.ResultType)

	argIdx := 0
	for _, sec := range proc.Params {
		for _, pname := range sec.Names {
			if argIdx < len(args) {
				val, err := r.evalExpr(args[argIdx])
				if err != nil {
					r.env = old
					return Value{}, err
				}
				child.vars[pname] = val
				argIdx++
			}
		}
	}

	err := r.execBlock(proc.Block)
	result := r.env.vars[name]
	r.env = old
	return result, err
}

func (r *Runtime) builtinWrite(args []ast.Expr, newline bool) error {
	for i, arg := range args {
		if str, ok := arg.(*ast.StringLit); ok {
			fmt.Fprint(r.output, str.Value)
		} else {
			val, err := r.evalExpr(arg)
			if err != nil {
				return err
			}
			switch val.Kind {
			case ValInt:
				fmt.Fprint(r.output, val.Int)
			case ValReal:
				fmt.Fprint(r.output, val.Real)
			case ValBool:
				fmt.Fprint(r.output, val.Bool)
			case ValChar:
				fmt.Fprintf(r.output, "%c", val.Char)
			case ValString:
				fmt.Fprint(r.output, val.Str)
			}
		}
		if i < len(args)-1 {
			fmt.Fprint(r.output, " ")
		}
	}
	if newline {
		fmt.Fprintln(r.output)
	}
	return nil
}

func (r *Runtime) builtinRead(args []ast.Expr, newline bool) error {
	for _, arg := range args {
		varName := ""
		switch a := arg.(type) {
		case *ast.Ident:
			varName = a.Name
		default:
			continue
		}
		line, err := r.readLine()
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		v := r.env.vars[varName]
		switch v.Kind {
		case ValInt:
			var n int
			fmt.Sscanf(line, "%d", &n)
			v.Int = n
		case ValReal:
			var f float64
			fmt.Sscanf(line, "%f", &f)
			v.Real = f
		case ValChar:
			if len(line) > 0 {
				v.Char = line[0]
			}
		case ValBool:
			v.Bool = line == "true"
		}
		r.env.vars[varName] = v
	}
	if newline && len(args) == 0 {
		_, _ = r.readLine()
	}
	return nil
}

func (r *Runtime) readLine() (string, error) {
	if r.inputIdx < len(r.inputData) {
		line := r.inputData[r.inputIdx]
		r.inputIdx++
		return line, nil
	}
	return r.input.ReadString('\n')
}

func toBool(v Value) bool {
	if v.Kind == ValBool {
		return v.Bool
	}
	return v.Int != 0
}

func toInt(v Value) int {
	switch v.Kind {
	case ValInt:
		return v.Int
	case ValReal:
		return int(v.Real)
	case ValChar:
		return int(v.Char)
	case ValBool:
		if v.Bool {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func toReal(v Value) float64 {
	switch v.Kind {
	case ValReal:
		return v.Real
	case ValInt:
		return float64(v.Int)
	default:
		return 0
	}
}

func valuesEqual(a, b Value) bool {
	if a.Kind != b.Kind {
		return toInt(a) == toInt(b)
	}
	switch a.Kind {
	case ValInt:
		return a.Int == b.Int
	case ValBool:
		return a.Bool == b.Bool
	case ValChar:
		return a.Char == b.Char
	default:
		return false
	}
}

func (r *Runtime) evalUnary(op string, v Value) (Value, error) {
	switch op {
	case "+":
		return v, nil
	case "-":
		if v.Kind == ValReal {
			return Value{Kind: ValReal, Real: -v.Real}, nil
		}
		return Value{Kind: ValInt, Int: -v.Int}, nil
	case "not":
		return Value{Kind: ValBool, Bool: !toBool(v)}, nil
	default:
		return Value{}, fmt.Errorf("unknown unary op: %s", op)
	}
}

func (r *Runtime) evalBinary(op string, left, right Value) (Value, error) {
	switch op {
	case "or":
		return Value{Kind: ValBool, Bool: toBool(left) || toBool(right)}, nil
	case "and":
		return Value{Kind: ValBool, Bool: toBool(left) && toBool(right)}, nil
	case "+":
		if left.Kind == ValReal || right.Kind == ValReal {
			return Value{Kind: ValReal, Real: toReal(left) + toReal(right)}, nil
		}
		return Value{Kind: ValInt, Int: toInt(left) + toInt(right)}, nil
	case "-":
		if left.Kind == ValReal || right.Kind == ValReal {
			return Value{Kind: ValReal, Real: toReal(left) - toReal(right)}, nil
		}
		return Value{Kind: ValInt, Int: toInt(left) - toInt(right)}, nil
	case "*":
		if left.Kind == ValReal || right.Kind == ValReal {
			return Value{Kind: ValReal, Real: toReal(left) * toReal(right)}, nil
		}
		return Value{Kind: ValInt, Int: toInt(left) * toInt(right)}, nil
	case "/":
		return Value{Kind: ValReal, Real: toReal(left) / toReal(right)}, nil
	case "div":
		d := toInt(right)
		if d == 0 {
			return Value{}, fmt.Errorf("division by zero")
		}
		return Value{Kind: ValInt, Int: toInt(left) / d}, nil
	case "mod":
		d := toInt(right)
		if d == 0 {
			return Value{}, fmt.Errorf("division by zero")
		}
		return Value{Kind: ValInt, Int: toInt(left) % d}, nil
	case "=":
		return Value{Kind: ValBool, Bool: compareValues(left, right) == 0}, nil
	case "<>":
		return Value{Kind: ValBool, Bool: compareValues(left, right) != 0}, nil
	case "<":
		return Value{Kind: ValBool, Bool: compareValues(left, right) < 0}, nil
	case "<=":
		return Value{Kind: ValBool, Bool: compareValues(left, right) <= 0}, nil
	case ">":
		return Value{Kind: ValBool, Bool: compareValues(left, right) > 0}, nil
	case ">=":
		return Value{Kind: ValBool, Bool: compareValues(left, right) >= 0}, nil
	default:
		return Value{}, fmt.Errorf("unknown binary op: %s", op)
	}
}

func compareValues(a, b Value) int {
	if a.Kind == ValBool || b.Kind == ValBool {
		ai, bi := 0, 0
		if toBool(a) {
			ai = 1
		}
		if toBool(b) {
			bi = 1
		}
		return ai - bi
	}
	if a.Kind == ValReal || b.Kind == ValReal {
		ar, br := toReal(a), toReal(b)
		if ar < br {
			return -1
		}
		if ar > br {
			return 1
		}
		return 0
	}
	return toInt(a) - toInt(b)
}

func (r *Runtime) evalBuiltinFunc(name string, args []ast.Expr) (Value, bool) {
	evalArgs := func() ([]Value, error) {
		var vals []Value
		for _, a := range args {
			v, err := r.evalExpr(a)
			if err != nil {
				return nil, err
			}
			vals = append(vals, v)
		}
		return vals, nil
	}

	switch name {
	case "abs":
		vals, err := evalArgs()
		if err != nil || len(vals) != 1 {
			return Value{}, false
		}
		if vals[0].Kind == ValReal {
			return Value{Kind: ValReal, Real: math.Abs(vals[0].Real)}, true
		}
		n := toInt(vals[0])
		if n < 0 {
			n = -n
		}
		return Value{Kind: ValInt, Int: n}, true
	case "sqr":
		vals, err := evalArgs()
		if err != nil || len(vals) != 1 {
			return Value{}, false
		}
		f := toReal(vals[0])
		return Value{Kind: ValReal, Real: f * f}, true
	case "odd":
		vals, err := evalArgs()
		if err != nil || len(vals) != 1 {
			return Value{}, false
		}
		return Value{Kind: ValBool, Bool: toInt(vals[0])%2 != 0}, true
	case "chr":
		vals, err := evalArgs()
		if err != nil || len(vals) != 1 {
			return Value{}, false
		}
		return Value{Kind: ValChar, Char: byte(toInt(vals[0]))}, true
	case "ord":
		vals, err := evalArgs()
		if err != nil || len(vals) != 1 {
			return Value{}, false
		}
		return Value{Kind: ValInt, Int: toInt(vals[0])}, true
	case "succ":
		vals, err := evalArgs()
		if err != nil || len(vals) != 1 {
			return Value{}, false
		}
		return Value{Kind: ValChar, Char: vals[0].Char + 1}, true
	case "pred":
		vals, err := evalArgs()
		if err != nil || len(vals) != 1 {
			return Value{}, false
		}
		return Value{Kind: ValChar, Char: vals[0].Char - 1}, true
	case "sqrt":
		vals, err := evalArgs()
		if err != nil || len(vals) != 1 {
			return Value{}, false
		}
		return Value{Kind: ValReal, Real: math.Sqrt(toReal(vals[0]))}, true
	case "sin", "cos", "exp", "ln", "arctan":
		vals, err := evalArgs()
		if err != nil || len(vals) != 1 {
			return Value{}, false
		}
		f := toReal(vals[0])
		switch name {
		case "sin":
			return Value{Kind: ValReal, Real: math.Sin(f)}, true
		case "cos":
			return Value{Kind: ValReal, Real: math.Cos(f)}, true
		case "exp":
			return Value{Kind: ValReal, Real: math.Exp(f)}, true
		case "ln":
			return Value{Kind: ValReal, Real: math.Log(f)}, true
		case "arctan":
			return Value{Kind: ValReal, Real: math.Atan(f)}, true
		}
	}
	return Value{}, false
}
