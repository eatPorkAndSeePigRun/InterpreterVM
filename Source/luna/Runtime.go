package luna

// Runtime stack, registers of each funciton is one part of stack.
type Stack struct {
	KBaseStackSize int64
	Stack_         []Value
	Top            *Value
}

// Set new top pointer, and [new top, old top) will be set nil
func (s Stack) SetNewTop(top *Value) {

}

// Function call stack info
type CallInfo struct {
	Register     *Value       // register base pointer which points to Stack
	Func_        *Value       // current closure, pointer to stack Value
	Instruction  *Instruction // current Instruction
	End          *Instruction // Instruction end
	ExpectResult int64        // expect result of this function call
}
