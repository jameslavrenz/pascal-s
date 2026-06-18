package ast

type TypeKind int

const (
	TypeInteger TypeKind = iota
	TypeReal
	TypeBoolean
	TypeChar
	TypeArray
	TypeRecord
	TypeNamed
)

type TypeDesc struct {
	Kind   TypeKind
	Name   string
	Elem   *TypeDesc
	Low    int
	High   int
	Fields []FieldDesc
}

type FieldDesc struct {
	Names []string
	Type  *TypeDesc
}

type Program struct {
	Name   string
	Params []string
	Block  *Block
}

type Block struct {
	Consts  []ConstDecl
	Types   []TypeDecl
	Vars    []VarDecl
	Procs   []ProcDecl
	Funcs   []FuncDecl
	Stmts   []Stmt
}

type ConstDecl struct {
	Name  string
	Value Expr
}

type TypeDecl struct {
	Name string
	Type *TypeDesc
}

type VarDecl struct {
	Names []string
	Type  *TypeDesc
}

type ParamSection struct {
	IsVar  bool
	Names  []string
	Type   *TypeDesc
}

type ProcDecl struct {
	Name   string
	Params []ParamSection
	Block  *Block
}

type FuncDecl struct {
	Name       string
	Params     []ParamSection
	ResultType *TypeDesc
	Block      *Block
}

type Stmt interface {
	stmtNode()
}

type AssignStmt struct {
	Target Expr
	Value  Expr
}

func (*AssignStmt) stmtNode() {}

type ProcCallStmt struct {
	Name string
	Args []Expr
}

func (*ProcCallStmt) stmtNode() {}

type CompoundStmt struct {
	Stmts []Stmt
}

func (*CompoundStmt) stmtNode() {}

type IfStmt struct {
	Cond Expr
	Then Stmt
	Else Stmt
}

func (*IfStmt) stmtNode() {}

type WhileStmt struct {
	Cond Expr
	Body Stmt
}

func (*WhileStmt) stmtNode() {}

type RepeatStmt struct {
	Body []Stmt
	Cond Expr
}

func (*RepeatStmt) stmtNode() {}

type ForStmt struct {
	Var   string
	Start Expr
	End   Expr
	Down  bool
	Body  Stmt
}

func (*ForStmt) stmtNode() {}

type CaseStmt struct {
	Expr  Expr
	Cases []CaseBranch
}

type CaseBranch struct {
	Labels []Expr
	Body   Stmt
}

func (*CaseStmt) stmtNode() {}

type Expr interface {
	exprNode()
}

type IntLit struct{ Value int }

func (*IntLit) exprNode() {}

type RealLit struct{ Value float64 }

func (*RealLit) exprNode() {}

type BoolLit struct{ Value bool }

func (*BoolLit) exprNode() {}

type CharLit struct{ Value byte }

func (*CharLit) exprNode() {}

type StringLit struct{ Value string }

func (*StringLit) exprNode() {}

type Ident struct{ Name string }

func (*Ident) exprNode() {}

type UnaryExpr struct {
	Op   string
	Expr Expr
}

func (*UnaryExpr) exprNode() {}

type BinaryExpr struct {
	Op    string
	Left  Expr
	Right Expr
}

func (*BinaryExpr) exprNode() {}

type CallExpr struct {
	Name string
	Args []Expr
}

func (*CallExpr) exprNode() {}

type IndexExpr struct {
	Array Expr
	Index Expr
}

func (*IndexExpr) exprNode() {}

type FieldExpr struct {
	Record Expr
	Field  string
}

func (*FieldExpr) exprNode() {}
