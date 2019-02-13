package luna

func numberToStr(num *Value) string {
	if num.Type != ValueTNumber {
		panic("assert")
	}
	// TODO
}

type VM struct {
	state *State
}

func (vm VM) executeFrame() {
	call := vm.state.calls.Back().Value.(*CallInfo)
	cl := call.Func_.Closure
	proto := cl.GetPrototype()
	var a, b, c *Value

	for call.Instruction < call.End {
		vm.state.CheckRunGC()
		i := *c
	}

}

// Execute next frame if return true
func (vm VM) call(a *Value, i Instruction) {

}

func (vm VM) generateClosure(a *Value, i Instruction) {

}

func (vm VM) copyVarArg(a *Value, i Instruction) {

}

func (vm VM) return_(a *Value, i Instruction) {

}

func (vm VM) concat(dst, op1, op2 *Value) {

}

func (vm VM) forInit(var_, limit, step *Value) {

}

// Debug help functions
func (vm VM) getOperandNameAndScope(a *Value) (string, string) {

}

func (vm VM) getCurrentInstructionPos() (string, int64) {

}

func (vm VM) checkType(v *Value, type_ ValueT, op string) {

}

func (vm VM) checkArithType(v1, v2 *Value, op string) {

}

func (vm VM) checkInequalityType(v1, v2 *Value, op string) {

}

func (vm VM) checkTableType(t, k *Value, op, desc string) {

}

func (vm VM) reportTypeError(v *Value, op string) {

}

func (vm VM) Execute() {
	if vm.state.calls.Len() == 0 {
		panic("assert")
	}

	for vm.state.calls.Len() != 0 {
		// If current stack frame is a frame of a c function,
		// do not continue execute instructions, just return
		if vm.state.calls.Back().Value.(CallInfo).Func_.Type == ValueTCFunction {
			return
		}
		vm.executeFrame()
	}
}
