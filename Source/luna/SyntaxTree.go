package luna

import "unsafe"

// Expression or variable operation semantic
const (
	SemanticOpNone = iota
	SemanticOpRead
	SemanticOpWrite
)

// Expression or variable lexical scoping
const (
	LexicalScopingUnknown = iota
	LexicalScopingGlobal  // Expression or variable in global table
	LexicalScopingUpValue // Expression or variable in upValue
	LexicalScopingLocal   // Expression or variable in current function
)

// AST base class, all AST node derived from this class and
// provide Visitor to Accept itself.
type SyntaxTree interface {
	Accept(v Visitor, data unsafe.Pointer)
}

type Chunk struct {
	Block  SyntaxTree
	Module *String
}

func NewChunk(block SyntaxTree, module *String) *Chunk {
	return &Chunk{block, module}
}

func (c Chunk) Accept(v Visitor, data unsafe.Pointer) {

}

type Block struct {
	Statements []*SyntaxTree
	ReturnStmt *SyntaxTree
}

func (b Block) Accept(v Visitor, data unsafe.Pointer) {

}

type ReturnStatement struct {
	ExpList       *SyntaxTree
	Line          int64
	ExpValueCount int64
}

func (r ReturnStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type BreakStatement struct {
	Break TokenDetail
	Loop  *SyntaxTree // For semantic
}

func (b BreakStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type DoStatement struct {
	Block *SyntaxTree
}

func (d DoStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type WhileStatement struct {
	Exp   *SyntaxTree
	Block *SyntaxTree

	FirstLine int64
	LastLine  int64
}

func (w WhileStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type RepeatStatement struct {
	Block *SyntaxTree
	Exp   *SyntaxTree

	Line int64 // Line of until
}

func (r RepeatStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type IfStatement struct {
	Exp         *SyntaxTree
	TrueBranch  *SyntaxTree
	FalseBranch *SyntaxTree

	Line         int64 // Line of if
	BlockEndLine int64 // End line of block
}

func (i IfStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type ElseIfStatement struct {
	Exp        *SyntaxTree
	TrueBranch *SyntaxTree
	False      *SyntaxTree

	Line         int64 // Line of elseif
	BlockEndLine int64 // End line of block
}

func (e ElseIfStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type ElseStatement struct {
	Block *SyntaxTree
}

func (e ElseStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type NumericForStatement struct {
	Name  TokenDetail
	Exp1  *SyntaxTree
	Exp2  *SyntaxTree
	Exp3  *SyntaxTree
	Block *SyntaxTree
}

func (n NumericForStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type GenericForStatement struct {
	NameList *SyntaxTree
	ExpList  *SyntaxTree
	Block    *SyntaxTree

	Line int64
}

func (g GenericForStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type FunctionStatement struct {
	FuncName *SyntaxTree
	FuncBody *SyntaxTree
}

func (f FunctionStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type FunctionName struct {
	Names      []TokenDetail
	MemberName TokenDetail
	Scoping    int // First token scoping
}

func (f FunctionName) Accept(v Visitor, data unsafe.Pointer) {

}

type LocalFunctionStatement struct {
	Name     TokenDetail
	FuncBody *SyntaxTree
}

func (l LocalFunctionStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type LocalNameListStatement struct {
	NameList  *SyntaxTree
	ExpList   *SyntaxTree
	Line      int64 // Start Line
	NameCount int64 // For semantic and code generate
}

func (l LocalNameListStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type AssignmentStatement struct {
	VarList  *SyntaxTree
	ExpList  *SyntaxTree
	Line     int64 // Start line
	VarCount int64 // For semantic
}

func (a AssignmentStatement) Accept(v Visitor, data unsafe.Pointer) {

}

type VarList struct {
	VarList []*SyntaxTree
}

func (vl VarList) Accept(v Visitor, data unsafe.Pointer) {

}

type Terminator struct {
	Token    TokenDetail
	Semantic int
	Scoping  int
}

func (t Terminator) Accept(v Visitor, data unsafe.Pointer) {

}

type BinaryExpression struct {
	Left    *SyntaxTree
	Right   *SyntaxTree
	OpToken TokenDetail
}

func (b BinaryExpression) Accept(v Visitor, data unsafe.Pointer) {

}

type UnaryExpression struct {
	Exp     *SyntaxTree
	OpToken TokenDetail
}

func (u UnaryExpression) Accept(v Visitor, data unsafe.Pointer) {

}

type FunctionBody struct {
	ParamList *SyntaxTree
	BLock     *SyntaxTree
	HasSelf   bool // For code generate, has 'self' param or not
	Line      int64
}

func (f FunctionBody) Accept(v Visitor, data unsafe.Pointer) {

}

type ParamList struct {
	NameList    *SyntaxTree
	Vararg      bool
	FixArgCount int64 // For semantic and code generate
}

func (p ParamList) Accept(v Visitor, data unsafe.Pointer) {

}

type NameList struct {
	Names []TokenDetail
}

func (n NameList) Accept(v Visitor, data unsafe.Pointer) {

}

type TableDefine struct {
	Fields []*SyntaxTree
	Line   int64
}

func (t TokenDetail) Accept(v Visitor, data unsafe.Pointer) {

}

type TableIndexField struct {
	Index *SyntaxTree
	Value *SyntaxTree
	Line  int64
}

func (t TableIndexField) Accept(v Visitor, data unsafe.Pointer) {

}

type TableNameField struct {
	Name  TokenDetail
	Value *SyntaxTree
}

func (t TableNameField) Accept(v Visitor, data unsafe.Pointer) {

}

type TableArrayField struct {
	Value *SyntaxTree
	Line  int64
}

func (t TableArrayField) Accept(v Visitor, data unsafe.Pointer) {

}

type IndexAccessor struct {
	Table    *SyntaxTree
	Index    *SyntaxTree
	Line     int64
	Semantic int // For semantic
}

func (i IndexAccessor) Accept(v Visitor, data unsafe.Pointer) {

}

type MemberAccessor struct {
	Table  *SyntaxTree
	Member TokenDetail
}

func (m MemberAccessor) Accept(v Visitor, data unsafe.Pointer) {

}

type NormalFuncCall struct {
	Caller *SyntaxTree
	Args   *SyntaxTree
	Line   int64 // Function call line in source
}

func (n NormalFuncCall) Accept(v Visitor, data unsafe.Pointer) {

}

type MemberFuncCall struct {
	Caller *SyntaxTree
	Member TokenDetail
	Args   *SyntaxTree
	Line   int64 // Function call line in source
}

func (m MemberFuncCall) Accept(v Visitor, data unsafe.Pointer) {
}

type FuncCallArgs struct {
	Arg           *SyntaxTree
	Type          ArgType
	ArgValueCount int64 // For code generate
}

type ArgType int64

const (
	ArgTypeExpList = iota
	ArgTypeTable
	ArgTypeString
)

func (f FuncCallArgs) Accept(v Visitor, data unsafe.Pointer) {

}

type ExpressionList struct {
	ExpList []*SyntaxTree
	Line    int64 // Start line
}

func (e ExpressionList) Accept(v Visitor, data unsafe.Pointer) {

}
