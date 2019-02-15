package luna

import "unsafe"

const KBaseStackSize int = 10000

// Runtime stack, registers of each funciton is one part of stack.
type Stack struct {
	Stack_ [KBaseStackSize]Value
	Top    *Value
}

func NewStack() Stack {
	var s Stack
	s.Top = &s.Stack_[0]
	return s
}

// Set new top pointer, and [new top, old top) will be set nil
func (s Stack) SetNewTop(top *Value) {
	old := s.Top
	s.Top = top

	// Clear values between new top to old
	for uintptr(unsafe.Pointer(top)) < uintptr(unsafe.Pointer(old)) {
		top.SetNil()
		uintptr(unsafe.Pointer(top))++
	}
}

// Function call stack info
type CallInfo struct {
	Register     *Value       // register base pointer which points to Stack
	Func_        *Value       // current closure, pointer to stack Value
	Instruction  *Instruction // current Instruction
	End          *Instruction // Instruction end
	ExpectResult int64        // expect result of this function call
}

func NewCallInfo() CallInfo {
	return CallInfo{}
}
