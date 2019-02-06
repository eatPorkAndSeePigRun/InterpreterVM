package luna

type ModuleManager struct {
	state   *State
	modules *Table
}

func (moduleManager ModuleManager) load(lexer Lexer) {

}

// Check module loaded or not
func (moduleManager ModuleManager) IsLoaded(moduleName string) bool {
	return false
}

// Get module closure when module loaded
// if the module is not loaded, return nil value
func (moduleManager ModuleManager) GetModuleClosure(moduleName string) Value {
	return Value{}
}

// Load module, when loaded success, push the closure of the module onto stack
func (moduleManager ModuleManager) LoadModule(moduleName string) {

}

// Load module, when loaded success, push the closure of the string onto stack
func (moduleManager ModuleManager) LoadString(str, name string) {

}
