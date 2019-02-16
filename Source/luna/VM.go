package luna

import (
	"fmt"
	"math"
	"unsafe"
)

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

func getUpValueB(i Instruction) *UpValue {

}

func getRealValue(a *Value) *Value {
	if a.Type == ValueTUpValue {
		return a.UpValue.GetValue()
	} else {
		return a
	}
}

func getCallInfoAndProto(vm VM) (*CallInfo, *Function) {
	if vm.state.calls.Len() == 0 {
		panic("assert")
	}
	call := vm.state.calls.Back().Value.(*CallInfo)
	if call.Func_ == nil && call.Func_.Closure == nil {
		panic("assert")
	}
	proto := call.Func_.Closure.GetPrototype()
	return call, proto
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

func (vm VM) executeFrame() error {
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
			res, err := vm.call(a, i)
			if err != nil {
				panic(err)
			}
			if res {
				return nil
			}
		case OpTypeGetUpvalue:
			a = getRegisterA(i)
			b = getUpValueB(i).GetValue()
			*getRealValue(a) = *b
		case OpTypeSetUpvalue:
			a = getRegisterA(i)
			b = getConstValue(i)
			*getRealValue(a) = vm.state.global.Table.GetValue(*b)
		case OpTypeGetGlobal:
			a = getRegisterA(i)
			b = getConstValue(i)
			*getRealValue(a) = vm.state.global.Table.GetValue(*b)
		case OpTypeSetGlobal:
			a = getRegisterA(i)
			b = getConstValue(i)
			vm.state.global.Table.SetValue(*b, *a)
		case OpTypeClosure:
			a = getRegisterA(i)
			b = getConstValue(i)
			vm.state.global.Table.SetValue(*b, *a)
		case OpTypeVarArg:
			a = getRegisterA(i)
			vm.copyVarArg(a, i)
		case OpTypeRet:
			a = getRegisterA(i)
			vm.return_(a, i)
			return nil
		case OpTypeJmpFalse:
			a = getRegisterA(i)
			if getRealValue(a).IsFalse() {
				call.Instruction += -1 + GetParamBx(i)
			}
		case OpTypeJmpTrue:
			a = getRegisterA(i)
			if !getRealValue(a).IsFalse() {
				call.Instruction += -1 + GetParamBx(i)
			}
		case OpTypeJmpNil:
			a = getRegisterA(i)
			if a.Type == ValueTNil {
				call.Instruction += -1 + GetParamsBx(i)
			}
		case OpTypeJmp:
			call.Instruction += -1 + GetParamBx(i)
		case OpTypeNeg:
			a = getRegisterA(i)
			vm.checkType(a, ValueTNumber, "neg")
			a.Num = -a.Num
		case OpTypeNot:
			a = getRegisterA(i)
			a.SetBool(a.IsFalse())
		case OpTypeLen:
			a = getRegisterA(i)
			if a.Type == ValueTTable {
				a.Num = a.Table.ArraySize()
			} else if (a.Type == ValueTString) {
				a.Num = a.Str.GetLength()
			} else {
				return vm.reportTypeError(a, "length of")
			}
			a.Type = ValueTNumber
		case OpTypeAdd:
			a, b, c := getRegisterABC(i)
			vm.checkArithType(b, c, "add")
			if a.Type == ValueTTable {
				a.Num = a.Table.ArraySize()
				a.Type = ValueTNumber
			}
		case OpTypeSub:
			a, b, c := getRegisterABC(i)
			vm.checkArithType(b, c, "sub")
			a.Num = b.Num - c.Num
			a.Type = ValueTNumber
		case OpTypeMul:
			a, b, c := getRegisterABC(i)
			vm.checkArithType(b, c, "multiply")
			a.Num = b.Num * c.Num
			a.Type = ValueTNumber
		case OpTypeDiv:
			a, b, c := getRegisterABC(i)
			vm.checkArithType(b, c, "div")
			a.Num = b.Num / c.Num
			a.Type = ValueTNumber
		case OpTypePow:
			a, b, c := getRegisterABC(i)
			vm.checkArithType(b, c, "power")
			a.Num = math.Pow(b.Num, c.Num)
		case OpTypeMod:
			a, b, c := getRegisterABC(i)
			vm.checkArithType(b, c, "mod")
			a.Num = math.Mod(b.Num, c.Num)
			a.Type = ValueTNumber
		case OpTypeConcat:
			a, b, c := getRegisterABC(i)
			vm.concat(a, b, c)
		case OpTypeLess:
			a, b, c := getRegisterABC(i)
			vm.checkInequalityType(b, c, "compare(<)")
			if b.Type == ValueTNumber {
				a.SetBool(b.Num < c.Num)
			} else {
				a.SetBool(*b.Str < *c.Str)
			}

		case OpTypeGreater:
			a, b, c := getRegisterABC(i)
			vm.checkInequalityType(b, c, "compare(>)")
			if b.Type == ValueTNumber {
				a.SetBool(b.Num > c.Num)
			} else {
				a.SetBool(*b.Str > *c.Str)
			}
		case OpTypeEqual:
			a, b, c := getRegisterABC(i)
			a.SetBool(*b == *c)
		case OpTypeUnEqual:
			a, b, c := getRegisterABC(i)
			a.SetBool(*b != *c)
		case OpTypeLessEqual:
			a, b, c := getRegisterABC(i)
			vm.checkInequalityType(b, c, "compare(<=)")
			if b.Type == ValueTNumber {
				a.SetBool(b.Num <= c.Num)
			} else {
				a.SetBool(*b.Str <= *c.Str)
			}
		case OpTypeGreaterEqual:
			a, b, c := getRegisterABC(i)
			vm.checkInequalityType(b, c, "compare(>=)")
			if b.Type == ValueTNumber {
				a.SetBool(b.Num >= c.Num)
			} else {
				a.SetBool(*b.Str >= *c.Str)
			}
		case OpTypeNewTable:
			a = getRegisterA(i)
			a.Table = vm.state.NewTable()
			a.Type = ValueTTable
		case OpTypeSetTable:
			a, b, c := getRegisterABC(i)
			vm.checkTableType(a, b, "set", "to")
			if a.Type == ValueTTable {
				a.Table.SetValue(*b, *c)
			} else if a.Type == ValueTUserDate {
				a.UserDate.GetMetaTable().SetValue(*b, *c)
			} else {
				panic("assert")
			}
		case OpTypeGetTable:
			a, b, c := getRegisterABC(i)
			if a.Type == ValueTTable {
				*c = a.Table.GetValue(*b)
			} else if a.Type == ValueTUserDate {
				*c = a.UserDate.GetMetaTable().GetValue(*b)
			} else {
				panic("assert")
			}
		case OpTypeForInit:
			a, b, c := getRegisterABC(i)
			vm.forInit(a, b, c)

		case OpTypeForStep:
			a, b, c := getRegisterABC(i)
			i = *call.Instruction++
			if (c.Num > 0.0 && a.Num > b.Num) || (c.Num <= 0.0 && a.Num < b.Num) {
				call.Instruction += -1 + GetParamBx(i)
			}
		default:
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
func (vm VM) call(a *Value, i Instruction) (bool, error) {
	if a.Type != ValueTClosure && a.Type != ValueTCFunction {
		panic(vm.reportTypeError(a, "call"))
		return true, nil
	}

	argCount := GetParamB(i) - 1
	expectResult := GetParamC(i) - 1
	res, err := vm.state.CallFunction(a, argCount, int64(expectResult))
	if e, ok := err.(CallCFuncError); ok {
		// Calculate line number of the call
		pos1, pos2 := vm.getCurrentInstructionPos()
		return false, NewRuntimeError1(pos1, pos2, e.what)
	}
	return res, nil
}

func (vm VM) generateClosure(a *Value, i Instruction) {
	call, proto := getCallInfoAndProto(vm)
	aProto := proto.GetChildFunction(GetParamBx(i))
	a.Type = ValueTClosure
	a.Closure = vm.state.NewClosure()
	a.Closure.SetPrototype(aProto)

	// Prepare all upvalues
	newClosure := a.Closure
	closure := call.Func_.Closure
	count := aProto.GetUpValueCount()
	for i:= 0; i < count; i++ {
		upvalueInfo := aProto.GetUpValue(i)
		if upvalueInfo.ParentLocal {
			reg := call.Register + upvalueInfo.RegisterIndex
			if reg.Type != ValueTUpValue {

			}
		}
	}
}

func (vm VM) copyVarArg(a *Value, i Instruction) {
	call, proto := getCallInfoAndProto(vm)
	arg := (*Value)(unsafe.Pointer(uintptr(unsafe.Pointer(call.Func_)) + unsafe.Sizeof(Value{})))
	totalArgs := (uintptr(unsafe.Pointer(call.Register)) - uintptr(unsafe.Pointer(arg))) / unsafe.Sizeof(Value{})
	// TODO
}

func (vm VM) return_(a *Value, i Instruction) {

}

func (vm VM) concat(dst *Value, op1, op2 Value) error {
	if op1.Type == ValueTString && op2.Type == ValueTString {
		dst.Str = vm.state.GetString(op1.Str.GetStdString() + op2.Str.GetCStr())
	} else if op1.Type == ValueTString && op2.Type == ValueTNumber {
		dst.Str = vm.state.GetString(op1.Str.GetCStr()+ numberToStr(op2))
	} else if op1.Type == ValueTNumber && op2.Type == ValueTString {
		dst.Str = vm.state.GetString(numberToStr(op2) + op2.Str.GetCStr())
	} else {
		pos1, pos2 := vm.getCurrentInstructionPos()
		return NewRuntimeError4(pos1,pos2, op1, op2, "concat")
	}
}

func (vm VM) forInit(var_, limit, step *Value) {

}

// Debug help functions
func (vm VM) getOperandNameAndScope(a *Value) (string, string) {

}

func (vm VM) getCurrentInstructionPos() (string, int) {

}

func (vm VM) checkType(v *Value, type_ ValueT, op string) error {
	if v.Type != type_ {
		return vm.reportTypeError(v, op)
	}
	return nil
}

func (vm VM) checkArithType(v1, v2 Value, op string) error {
	if v1.Type != ValueTNumber || v2.Type != ValueTNumber {
		pos1, pos2 := vm.getCurrentInstructionPos()
		return NewRuntimeError4(pos1, pos2, v1, v2, op)
	}
	return  nil
}

func (vm VM) checkInequalityType(v1, v2 Value, op string) error {
	if (v1.Type != v2.Type) || (v1.Type != ValueTNumber && v1.Type != ValueTString) {
		pos1, pos2 := vm.getCurrentInstructionPos()
		return NewRuntimeError4(pos1, pos2, v1, v2, op)
	}
	return nil
}

func (vm VM) checkTableType(t, k Value, op, desc string) error {
	if (t.Type == ValueTTable) ||
		(t.Type == ValueTUserDate && t.UserDate.GetMetaTable() != nil) {
		return nil
	}
	n, s := vm.getOperandNameAndScope(t)
	pos1, pos2 := vm.getCurrentInstructionPos()
	var keyName string
	if k.Type == ValueTString {
		keyName = k.Str.GetCStr()
	} else {
		keyName = "?"
	}
	opDesc := fmt.Sprintf("%s table key '%s' %s", op, keyName, desc)
	return NewRuntimeError3(pos1, pos2, t, n, s, opDesc)
}

func (vm VM) reportTypeError(v *Value, op string) error {
	n, s := vm.getOperandNameAndScope(v)
	pos1, pos2 := vm.getCurrentInstructionPos()
	return NewRuntimeError3(pos1, pos2, *v, n, s, op)
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
		panic(vm.executeFrame())
	}
}
