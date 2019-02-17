package compiler

// Lexical block data in LexicalFunction for finding
type lexicalBlock struct {
	Parent *lexicalBlock
	// Local names
	// Same names are the same instance String, so using String
	// pointer as key is fine
	Names map[*String]int
}

func newLexicalBlock() lexicalBlock {
	return lexicalBlock{}
}

// Lexical function data for name finding
type lexicalFunction struct {
	Parent       *lexicalFunction
	CurrentBlock *lexicalBlock
	CurrentLoop  *SyntaxTree
	HasVararg    bool
}

func newLexicalFunction() lexicalFunction {
	return lexicalFunction{}
}

type semanticAnalysisVisitor struct {
	state           *State
	currentFunction *lexicalFunction // Current lexical function for all names finding
}

// TODO

// For NameList AST
type nameListData struct {
	NameCount int
}

func newNameListData() nameListData {
	return nameListData{}
}

// For VarList AST
type varListData struct {
	VarCount int
}

func newVarListData() varListData {
	return varListData{}
}

// For ExpList AST
type expListData struct {
	ExpValueCount int
}

func newExpListData() expListData {
	return expListData{}
}

// Expression type for semantic
const (
	ExpTypeUnknown = iota // unknown type at semantic analysis phase
	ExpTypeNil
	ExpTypeBool
	ExpTypeNumber
	ExpTypeString
	ExpTypeVarArg
	ExpTypeTable
)

// Expression or variable data for semantic analysis
type expVarData struct {
	SemanticOp      int
	ExpType         int
	ResultsAnyCount bool
}

func newExpVarData(semanticOp int) expVarData {
	return expVarData{semanticOp, ExpTypeUnknown, false}
}

// For FunctionName
type functionNameData struct {
	HasMemberToken bool
}

func newFunctionNameData() functionNameData {
	return functionNameData{}
}

func SemanticAnalysis(root *SyntaxTree, state *State) {
	if root == nil && state == nil {
		panic("assert")
	}

}
