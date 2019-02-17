package vm

import (
	"InterpreterVM/Source/compiler"
	"InterpreterVM/Source/datatype"
	"InterpreterVM/Source/io/text"
)

// Load and manager all modules or load string
type ModuleManager struct {
	state   *State
	modules *datatype.Table
}

func NewModuleManager(state *State, modules *datatype.Table) *ModuleManager {
	return &ModuleManager{state, modules}
}

// Load and push the closure onto stack
func (mm ModuleManager) load(lexer *compiler.Lexer) {
	// Parse to AST
	ast := compiler.Parse(lexer)

	// Semantic analysis
	compiler.SemanticAnalysis(ast, mm.state)

	// Generate code
	compiler.CodeGenerate(ast, mm.state)
}

// Check module loaded or not
func (mm ModuleManager) IsLoaded(moduleName string) bool {
	value := mm.GetModuleClosure(moduleName)
	return !value.IsNil()
}

// Get module closure when module loaded
// if the module is not loaded, return nil value
func (mm ModuleManager) GetModuleClosure(moduleName string) datatype.Value {
	key := datatype.NewValueString(mm.state.GetString(moduleName))
	return mm.modules.GetValue(key)
}

// Load module, when loaded success, push the closure of the module onto stack
func (mm ModuleManager) LoadModule(moduleName string) error {
	is := text.NewInStream(moduleName)
	if !is.IsOpen() {
		return NewOpenFileFail(moduleName)
	}

	lexer := compiler.NewLexer(mm.state, mm.state.GetString(moduleName),
		func() byte { return is.GetChar() })
	mm.load(&lexer)

	// Add to modules' table
	key := datatype.NewValueString(mm.state.GetString(moduleName))
	value := mm.state.stack.ValueStack[mm.state.stack.Top-1]
	mm.modules.SetValue(key, value)

	return nil
}

// Load module, when loaded success, push the closure of the string onto stack
func (mm ModuleManager) LoadString(str, name string) {
	is := text.NewInStringStream(str)
	lexer := compiler.NewLexer(mm.state, mm.state.GetString(name),
		func() byte { return is.GetChar() })
	mm.load(&lexer)
}
