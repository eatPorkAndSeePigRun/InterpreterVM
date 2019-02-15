package luna

import "InterpreterVM/Source/io/text"

type ModuleManager struct {
	state   *State
	modules *Table
}

func NewModuleManager(state *State, modules *Table) *ModuleManager {
	return &ModuleManager{state, modules}
}

func (m ModuleManager) load(lexer *Lexer) {
	// Parse to AST
	ast := Parse(lexer)

	// Semantic analysis
	SemanticAnalysis(ast, m.state)

	// Generate code
	CodeGenerate(ast, m.state)
}

// Check module loaded or not
func (m ModuleManager) IsLoaded(moduleName string) bool {
	value := m.GetModuleClosure(moduleName)
	return !value.IsNil()
}

// Get module closure when module loaded
// if the module is not loaded, return nil value
func (m ModuleManager) GetModuleClosure(moduleName string) Value {
	key := NewValueString(m.state.GetString(moduleName))
	return m.modules.GetValue(key)
}

// Load module, when loaded success, push the closure of the module onto stack
func (m ModuleManager) LoadModule(moduleName string) error {
	is := text.NewInStream(moduleName)
	if !is.IsOpen() {
		return NewOpenFileFail(moduleName)
	}

	lexer := NewLexer(m.state, m.state.GetString(moduleName), func() uint8 { return is.GetChar() })
	m.load(&lexer)

	// Add to modules' table
	key := NewValueString(m.state.GetString(moduleName))
	value :=
}

// Load module, when loaded success, push the closure of the string onto stack
func (m ModuleManager) LoadString(str, name string) {
	is := text.NewInStringStream(str)
	lexer := NewLexer(m.state, m.state.GetString(name), func() uint8 { return is.GetChar() })
	m.load(&lexer)
}
