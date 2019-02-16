package luna

import "unsafe"

func numberToStr(num *Value) string {
	if num.Type != ValueTNumber {
		panic("assert")
	}
	// TODO
}

func getConstValue(i Instruction) *Value {

}

func getRegisterA(i Instruction) *Value {

}

func getRegisterB(i Instruction) *Value {

}

func getRegisterC(i Instruction) *Value {

}

func getUpValueB(i Instruction) {

}

func getRealValue(a *Value) *Value {
	if a.Type == ValueTUpValue {
		return a.UpValue.GetValue()
	} else {
		return a
	}
}

func getCallinfoAndProto(vm VM) {
	if vm.state.calls.Len() == 0 {
		panic("assert")
	}
	call := vm.state.calls.Back().Value.(*CallInfo)
	if call.Func_ == nil && call.Func_.Closure == nil {
		panic("assert")
	}
	proto := call.Func_.Closure.GetPrototype()
	return
}

func getRegisterABC(i Instruction) (a, b, c *Value) {
	return getRegisterA(i), getRegisterB(i), getRegisterC(i)
}

type VM struct {
	state *State
}

func NewVM(state *State) VM {
	return VM{state}
}

func (vm VM) executeFrame() {
	call := vm.state.calls.Back().Value.(*CallInfo)
	cl := call.Func_.Closure
	proto := cl.GetPrototype()
	var a, b, c *Value

	for uintptr(unsafe.Pointer(call.Instruction)) < uintptr(unsafe.Pointer(call.End)) {
		vm.state.CheckRunGC()
		i := *call.Instruction
		temp := uintptr(unsafe.Pointer(call.Instruction))
		temp += unsafe.Sizeof(Instruction{})
		call.Instruction = (*Instruction)(unsafe.Pointer(temp))

		switch GetOpCode(i) {
		case OpTypeLoadNil:
			a = getRegisterA(i)
			getRealValue(a).SetNil()
		case OpTypeFillNil:
			a = getRegisterA(i)
			b = getRegisterB(i)
			for uintptr(unsafe.Pointer(a)) < uintptr(unsafe.Pointer(b)) {
				a.SetNil()
				a = (*Value)(unsafe.Pointer(uintptr(unsafe.Pointer(a)) + unsafe.Sizeof(Value{})))
			}
		case OpTypeLoadBool:
			a = getRegisterA(i)
			getRealValue(a).SetBool(GetParamB(i) == 0)
		case OpTypeLoadInt:
			a = getRegisterA(i)
			if uintptr(unsafe.Pointer(call.Instruction)) > uintptr(unsafe.Pointer(call.End)) {
				panic("assert")
			}
			a.Num = (float64)((*call.Instruction).OpCode)
			a.Type = ValueTNumber
		case OpTypeLoadConst:
			a = getRegisterA(i)
			b = getConstValue(i)
			*getRealValue(a) = *b
		case OpTypeMove:
			a = getRegisterA(i)
			b = getRegisterB(i)
			*getRealValue(a) = *getRealValue(b)
		case OpTypeCall:
			a = getRegisterA(i)
			if vm.call(a, i) {
				return
			}
		case OpTypeGetUpvalue:
		case OpTypeSetUpvalue:
		case OpTypeGetGlobal:
		case OpTypeSetGlobal:
		case OpTypeClosure:
		case OpTypeVarArg:
		case OpTypeRet:
		case OpTypeJmpFalse:
		case OpTypeJmpTrue:
		case OpTypeJmpNil:
		case OpTypeJmp:
		case OpTypeNeg:
		case OpTypeNot:
		case OpTypeLen:
		case OpTypeAdd:
		case OpTypeSub:
		case OpTypeMul:
		case OpTypeDiv:
		case OpTypePow:
		case OpTypeMod:
		case OpTypeConcat:
		case OpTypeLess:
		case OpTypeGreater:
		case OpTypeEqual:
		case OpTypeUnEqual:
		case OpTypeLessEqual:
		case OpTypeGreaterEqual:
		case OpTypeNewTable:
		case OpTypeSetTable:
		case OpTypeGetTable:
		case OpTypeForInit:
		case OpTypeForStep:
		}
	}

	newTop := call.Func_
	// Reset top value
	vm.state.stack.SetNewTop(newTop)
	// Set expect results
	if call.ExpectResult != ExpValueCountAny {
		vm.state.stack.SetNewTop(newTop + call.ExpectResult)
	}
	// Pop current CallInfo, and return to last CallInfo
	//vm.state.calls TODO
}

// Execute next frame if return true
func (vm VM) call(a *Value, i Instruction) bool {
	if a.Type != ValueTClosure && a.Type != ValueTCFunction {
		vm.reportTypeError(a, "call")
		return true
	}

	//argCount :=
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

func (vm VM) reportTypeError(v *Value, op string) error {
	n, s := vm.getOperandNameAndScope(v)
	pos1, pos2 := vm.getCurrentInstructionPos()
	return
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
