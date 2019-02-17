package datatype

import "InterpreterVM/Source/vm"

// Function prototype class, all runtime functions(closures) reference this
// class object. This class contains some static information generated after
// parse.
type Function struct {
	gcObjectField
	opCodes     []vm.Instruction // function instruction opCodes
	opCodeLines []int64          // opCodes' line number
	constValues []Value          // const values in function
	localVars   []localVarInfo   // debug info
	childFuncs  []*Function      // child functions
	upvalues    []UpvalueInfo    // upvalues
	module      *String          // function define module name
	line        int64            // function define line at module
	args        int64            // count of args
	isVararg    bool             // has '...' param or not
	superior    *Function        // superior function pointer
}

func NewFunction() *Function {
	return &Function{}
}

// For debug
type localVarInfo struct {
	Name       *String // Local variable name
	RegisterId int64   // Register id in function
	BeginPc    int64   // Begin instruction index of variable
	EndPc      int64   // The past-the-end instruction index
}

type UpvalueInfo struct {
	// Upvalue name
	Name *String

	// This upvalue is parent function's local variable
	// when value is true, otherwise it is parent parent
	// (... and so on) function's local variable
	ParentLocal bool

	// Register id when this upvalue is parent function's
	// local variable, otherwise it is index of upvalue list
	// of parent function
	RegisterIndex int64
}

func (f *Function) Accept(v GCObjectVisitor) {
	if v.VisitFunction(&f) {
		if f.module != nil {
			f.module.Accept(v)
		}
		if f.superior != nil {
			f.superior.Accept(v)
		}

		for _, value := range f.constValues {
			value.Accept(v)
		}

		for _, var_ := range f.localVars {
			var_.Name.Accept(v)
		}

		for _, child := range f.childFuncs {
			child.Accept(v)
		}

		for _, upvalue := range f.upvalues {
			upvalue.Name.Accept(v)
		}
	}
}

// Get function instructions and size
func (f *Function) GetOpCodes() *Instruction {
	if len(f.opCodes) == 0 {
		return nil
	} else {
		return &f.opCodes[0]
	}
}

func (f *Function) OpCodeSize() int64 {
	return int64(len(f.opCodes))
}

// Get instruction pointer, then it can be changed
func (f *Function) GetMutableInstruction(index int64) *Instruction {
	return &f.opCodes[index]
}

// Add instruction, 'line' is line number of the instruction 'i',
// return index of the new instruction
func (f *Function) AddInstruction(i Instruction, line int64) int64 {
	f.opCodes = append(f.opCodes, i)
	f.opCodeLines = append(f.opCodeLines, line)
	return int64(len(f.opCodes)) - 1
}

// Set this function has vararg
func (f *Function) SetHasVararg() {
	f.isVararg = true
}

// Get this function has vararg
func (f *Function) HasVararg() bool {
	return f.isVararg
}

// Add fixed arg count
func (f *Function) AddFixedArgCount(count int64) {
	f.args += count
}

// get fixed arg count
func (f *Function) FixedArgCount() int64 {
	return f.args
}

// Set module and function define start line
func (f *Function) SetModuleName(module *String) {
	f.module = module
}

func (f *Function) SetLine(line int64) {
	f.line = line
}

// Set superior function
func (f *Function) SetSuperior(superior *Function) {
	f.superior = superior
}

// Add const number and return index of the const value
func (f *Function) AddConstNumber(num float64) int64 {
	v := Value{Type: ValueTNumber, Num: num}
	return f.AddConstValue(&v)
}

// Add const String and return index of the const value
func (f *Function) AddConstString(str *String) int64 {
	v := Value{Type: ValueTString, Str: str}
	return f.AddConstValue(&v)
}

// Add const Value and return index of the const value
func (f *Function) AddConstValue(v *Value) int64 {
	f.constValues = append(f.constValues, *v)
	return int64(len(f.constValues)) - 1
}

// Add local variable debug info
func (f *Function) AddLocalVar(name *String, registerId, beginPc, endPc int64) {
	f.localVars = append(f.localVars, localVarInfo{name, registerId, beginPc, endPc})
}

// Add child function, return index of the function
func (f *Function) AddChildFunction(child *Function) int64 {
	f.childFuncs = append(f.childFuncs, child)
	return int64(len(f.childFuncs)) - 1
}

// Add a upvalue, return index of the upvalue
func (f *Function) AddUpvalue(name *String, parentLocal bool, registerIndex int64) int64 {
	f.upvalues = append(f.upvalues, UpvalueInfo{name, parentLocal, registerIndex})
	return int64(len(f.upvalues)) - 1
}

// Get upvalue index when the name upvalue existed, otherwise return -1
func (f *Function) SearchUpvalue(name *String) int64 {
	size := len(f.upvalues)
	for i := 0; i < size; i++ {
		if f.upvalues[i].Name == name {
			return int64(i)
		}
	}

	return -1
}

// Get child function by index
func (f *Function) GetChildFunction(index int64) *Function {
	return f.childFuncs[index]
}

// Search local variable name from local variable list
func (f *Function) SearchLocalVar(registerId, pc int64) *String {
	var name *String
	endPc := int(^uint(0) >> 1)
	beginPc := ^endPc

	for _, var_ := range f.localVars {
		if var_.RegisterId == registerId &&
			var_.BeginPc <= pc && pc < var_.EndPc {
			if int(var_.BeginPc) >= beginPc && int(var_.EndPc) <= endPc {
				name = var_.Name
				beginPc = int(var_.BeginPc)
				endPc = int(var_.EndPc)
			}
		}
	}

	return name
}

// Get const Value by index
func (f *Function) GetConstValue(i int64) *Value {
	return &f.constValues[i]
}

// Get instruction line by instruction index
func (f *Function) GetInstructionLine(i int64) int64 {
	return f.opCodeLines[i]
}

// Get upvalue count
func (f *Function) GetUpvalueCount() int {
	return len(f.upvalues)
}

// Get upvalue info by index
func (f *Function) GetUpvalue(index int64) *UpvalueInfo {
	return &f.upvalues[index]
}

// Get module name
func (f *Function) GetModule() *String {
	return f.module
}

// Get line of function define
func (f *Function) GetLine() int64 {
	return f.line
}

// All runtime function are closures, this class object pointer to a
// prototype Function object and its upvalues.
type Closure struct {
	gcObjectField
	prototype *Function  // prototype Function
	upvalues  []*Upvalue // upvalues
}

func NewClosure() *Closure {
	return &Closure{}
}

func (c *Closure) Accept(visitor GCObjectVisitor) {
	if visitor.VisitClosure(&c) {
		c.prototype.Accept(visitor)
		for _, v := range c.upvalues {
			v.Accept(visitor)
		}
	}
}

// Get closure prototype Function
func (c *Closure) GetPrototype() *Function {
	return c.prototype
}

// Set closure prototype Function
func (c *Closure) SetPrototype(prototype *Function) {
	c.prototype = prototype
}

// Add upvalue
func (c *Closure) AddUpvalue(upvalue *Upvalue) {
	c.upvalues = append(c.upvalues, upvalue)
}

// Get upvalue by index
func (c *Closure) GetUpvalue(index int64) *Upvalue {
	return c.upvalues[index]
}
