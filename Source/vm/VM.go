package vm

import (
	"InterpreterVM/Source/datatype"
	"fmt"
	"math"
	"unsafe"
)

func numberToStr(num *datatype.Value) string {
	if num.Type != datatype.ValueTNumber {
		panic("assert")
	}
	if math.Floor(num.Num) == num.Num {
		return fmt.Sprintf("%d", int64(num.Num))
	} else {
		return fmt.Sprintf("%G", num.Num)
	}
}

func getConstValue(i Instruction, proto *datatype.Function) *datatype.Value {
	return proto.GetConstValue(int(GetParamBx(i)))
}

func getRegisterA(i Instruction, call *CallInfo) *datatype.Value {
	return vPointerAdd(call.Register, GetParamA(i))
}

func getRegisterB(i Instruction, call *CallInfo) *datatype.Value {
	return vPointerAdd(call.Register, GetParamB(i))
}

func getRegisterC(i Instruction, call *CallInfo) *datatype.Value {
	return vPointerAdd(call.Register, GetParamC(i))
}

func getUpvalueB(i Instruction, cl *datatype.Closure) *datatype.Upvalue {
	return cl.GetUpvalue(GetParamB(i))
}

func getRealValue(a *datatype.Value) *datatype.Value {
	if a.Type == datatype.ValueTUpvalue {
		return a.Upvalue.GetValue()
	} else {
		return a
	}
}

func getCallInfoAndProto(vm *VM) (*CallInfo, *datatype.Function) {
	if vm.state.calls.Len() == 0 {
		panic("assert")
	}
	call := vm.state.calls.Back().Value.(*CallInfo)
	if call.Func == nil && call.Func.Closure == nil {
		panic("assert")
	}
	proto := call.Func.Closure.GetPrototype()
	return call, proto
}

func getRegisterABC(i Instruction, call *CallInfo) (a, b, c *datatype.Value) {
	return getRegisterA(i, call), getRegisterB(i, call), getRegisterC(i, call)
}

type VM struct {
	state *State
}

func NewVM(state *State) VM {
	return VM{state}
}

func (vm *VM) executeFrame() error {
	call := vm.state.calls.Back().Value.(*CallInfo)
	cl := call.Func.Closure
	proto := cl.GetPrototype()
	var a, b *datatype.Value

	for uintptr(unsafe.Pointer(call.Instruction)) < uintptr(unsafe.Pointer(call.End)) {
		vm.state.CheckRunGC()
		i := *call.Instruction
		call.Instruction = iPointerAdd(call.Instruction, 1)

		switch GetOpCode(i) {
		case OpTypeLoadNil:
			a = getRegisterA(i, call)
			getRealValue(a).SetNil()
		case OpTypeFillNil:
			a = getRegisterA(i, call)
			b = getRegisterB(i, call)
			for uintptr(unsafe.Pointer(a)) < uintptr(unsafe.Pointer(b)) {
				a.SetNil()
				a = vPointerAdd(a, 1)
			}
		case OpTypeLoadBool:
			a = getRegisterA(i, call)
			getRealValue(a).SetBool(GetParamB(i) == 0)
		case OpTypeLoadInt:
			a = getRegisterA(i, call)
			if uintptr(unsafe.Pointer(call.Instruction)) > uintptr(unsafe.Pointer(call.End)) {
				panic("assert")
			}
			a.Num = (float64)((*call.Instruction).OpCode)
			a.Type = datatype.ValueTNumber
		case OpTypeLoadConst:
			a = getRegisterA(i, call)
			b = getConstValue(i, proto)
			*getRealValue(a) = *b
		case OpTypeMove:
			a = getRegisterA(i, call)
			b = getRegisterB(i, call)
			*getRealValue(a) = *getRealValue(b)
		case OpTypeCall:
			a = getRegisterA(i, call)
			res, err := vm.call(a, i)
			if err != nil {
				panic(err)
			}
			if res {
				return nil
			}
		case OpTypeGetUpvalue:
			a = getRegisterA(i, call)
			b = getUpvalueB(i, cl).GetValue()
			*getRealValue(a) = *b
		case OpTypeSetUpvalue:
			a = getRegisterA(i, call)
			b = getConstValue(i, proto)
			*getRealValue(a) = vm.state.global.Table.GetValue(*b)
		case OpTypeGetGlobal:
			a = getRegisterA(i, call)
			b = getConstValue(i, proto)
			*getRealValue(a) = vm.state.global.Table.GetValue(*b)
		case OpTypeSetGlobal:
			a = getRegisterA(i, call)
			b = getConstValue(i, proto)
			vm.state.global.Table.SetValue(*b, *a)
		case OpTypeClosure:
			a = getRegisterA(i, call)
			b = getConstValue(i, proto)
			vm.state.global.Table.SetValue(*b, *a)
		case OpTypeVarArg:
			a = getRegisterA(i, call)
			vm.copyVarArg(a, i)
		case OpTypeRet:
			a = getRegisterA(i, call)
			vm.return_(a, i)
			return nil
		case OpTypeJmpFalse:
			a = getRegisterA(i, call)
			if getRealValue(a).IsFalse() {
				call.Instruction = iPointerAdd(call.Instruction, -1+int(GetParamBx(i)))
			}
		case OpTypeJmpTrue:
			a = getRegisterA(i, call)
			if !getRealValue(a).IsFalse() {
				call.Instruction = iPointerAdd(call.Instruction, -1+int(GetParamBx(i)))
			}
		case OpTypeJmpNil:
			a = getRegisterA(i, call)
			if a.Type == datatype.ValueTNil {
				call.Instruction = iPointerAdd(call.Instruction, -1+int(GetParamBx(i)))
			}
		case OpTypeJmp:
			call.Instruction = iPointerAdd(call.Instruction, -1+int(GetParamBx(i)))
		case OpTypeNeg:
			a = getRegisterA(i, call)
			if err := vm.checkType(a, datatype.ValueTNumber, "neg"); err != nil {
				panic(err)
			}
			a.Num = -a.Num
		case OpTypeNot:
			a = getRegisterA(i, call)
			a.SetBool(a.IsFalse())
		case OpTypeLen:
			a = getRegisterA(i, call)
			if a.Type == datatype.ValueTTable {
				a.Num = float64(a.Table.ArraySize())
			} else if a.Type == datatype.ValueTString {
				a.Num = float64(a.Str.GetLength())
			} else {
				return vm.reportTypeError(a, "length of")
			}
			a.Type = datatype.ValueTNumber
		case OpTypeAdd:
			a, b, c := getRegisterABC(i, call)
			if err := vm.checkArithType(*b, *c, "add"); err != nil {
				panic(err)
			}
			if a.Type == datatype.ValueTTable {
				a.Num = float64(a.Table.ArraySize())
				a.Type = datatype.ValueTNumber
			}
		case OpTypeSub:
			a, b, c := getRegisterABC(i, call)
			if err := vm.checkArithType(*b, *c, "sub"); err != nil {
				panic(err)
			}
			a.Num = b.Num - c.Num
			a.Type = datatype.ValueTNumber
		case OpTypeMul:
			a, b, c := getRegisterABC(i, call)
			if err := vm.checkArithType(*b, *c, "multiply"); err != nil {
				panic(err)
			}
			a.Num = b.Num * c.Num
			a.Type = datatype.ValueTNumber
		case OpTypeDiv:
			a, b, c := getRegisterABC(i, call)
			if err := vm.checkArithType(*b, *c, "div"); err != nil {
				panic(err)
			}
			a.Num = b.Num / c.Num
			a.Type = datatype.ValueTNumber
		case OpTypePow:
			a, b, c := getRegisterABC(i, call)
			if err := vm.checkArithType(*b, *c, "power"); err != nil {
				panic(err)
			}
			a.Num = math.Pow(b.Num, c.Num)
		case OpTypeMod:
			a, b, c := getRegisterABC(i, call)
			if err := vm.checkArithType(*b, *c, "mod"); err != nil {
				panic(err)
			}
			a.Num = math.Mod(b.Num, c.Num)
			a.Type = datatype.ValueTNumber
		case OpTypeConcat:
			a, b, c := getRegisterABC(i, call)
			if err := vm.concat(a, b, c); err != nil {
				panic(err)
			}
		case OpTypeLess:
			a, b, c := getRegisterABC(i, call)
			if err := vm.checkInequalityType(*b, *c, "compare(<)"); err != nil {
				panic(err)
			}
			if b.Type == datatype.ValueTNumber {
				a.SetBool(b.Num < c.Num)
			} else {
				a.SetBool(b.Str.IsLess(*c.Str))
			}
		case OpTypeGreater:
			a, b, c := getRegisterABC(i, call)
			if err := vm.checkInequalityType(*b, *c, "compare(>)"); err != nil {
				panic(err)
			}
			if b.Type == datatype.ValueTNumber {
				a.SetBool(b.Num > c.Num)
			} else {
				a.SetBool(c.Str.IsLess(*b.Str))
			}
		case OpTypeEqual:
			a, b, c := getRegisterABC(i, call)
			a.SetBool(b.IsEqual(c))
		case OpTypeUnEqual:
			a, b, c := getRegisterABC(i, call)
			a.SetBool(!b.IsEqual(c))
		case OpTypeLessEqual:
			a, b, c := getRegisterABC(i, call)
			if err := vm.checkInequalityType(*b, *c, "compare(<=)"); err != nil {
				panic(err)
			}
			if b.Type == datatype.ValueTNumber {
				a.SetBool(b.Num <= c.Num)
			} else {
				a.SetBool(b.Str.IsLess(*c.Str))
			}
		case OpTypeGreaterEqual:
			a, b, c := getRegisterABC(i, call)
			if err := vm.checkInequalityType(*b, *c, "compare(>=)"); err != nil {
				panic(err)
			}
			if b.Type == datatype.ValueTNumber {
				a.SetBool(b.Num >= c.Num)
			} else {
				a.SetBool(!b.Str.IsLess(*c.Str))
			}
		case OpTypeNewTable:
			a = getRegisterA(i, call)
			a.Table = vm.state.NewTable()
			a.Type = datatype.ValueTTable
		case OpTypeSetTable:
			a, b, c := getRegisterABC(i, call)
			if err := vm.checkTableType(*a, *b, "set", "to"); err != nil {
				panic(err)
			}
			if a.Type == datatype.ValueTTable {
				a.Table.SetValue(*b, *c)
			} else if a.Type == datatype.ValueTUserData {
				a.UserDate.GetMetaTable().SetValue(*b, *c)
			} else {
				panic("assert")
			}
		case OpTypeGetTable:
			a, b, c := getRegisterABC(i, call)
			if a.Type == datatype.ValueTTable {
				*c = a.Table.GetValue(*b)
			} else if a.Type == datatype.ValueTUserData {
				*c = a.UserDate.GetMetaTable().GetValue(*b)
			} else {
				panic("assert")
			}
		case OpTypeForInit:
			a, b, c := getRegisterABC(i, call)
			if err := vm.forInit(a, b, c); err != nil {
				panic(err)
			}
		case OpTypeForStep:
			a, b, c := getRegisterABC(i, call)
			i = *call.Instruction
			call.Instruction = iPointerAdd(call.Instruction, 1)
			if (c.Num > 0.0 && a.Num > b.Num) || (c.Num <= 0.0 && a.Num < b.Num) {
				call.Instruction = iPointerAdd(call.Instruction, -1+int(GetParamBx(i)))
			}
		}
	}

	newTop := call.Func
	// Reset top value
	vm.state.stack.SetNewTop(newTop)
	// Set expect results
	if call.ExpectResult != datatype.ExpValueCountAny {
		vm.state.stack.SetNewTop(vPointerAdd(newTop, call.ExpectResult))
	}
	// Pop current CallInfo, and return to last CallInfo
	vm.state.calls.Remove(vm.state.calls.Back())
	return nil
}

// Execute next frame if return true
func (vm *VM) call(a *datatype.Value, i Instruction) (bool, error) {
	if a.Type != datatype.ValueTClosure && a.Type != datatype.ValueTCFunction {
		panic(vm.reportTypeError(a, "call"))
		return true, nil
	}

	argCount := GetParamB(i) - 1
	expectResult := GetParamC(i) - 1
	res, err := vm.state.CallFunction(a, argCount, int64(expectResult))
	if e, ok := err.(*CallCFuncError); ok {
		// Calculate line number of the call
		pos1, pos2 := vm.getCurrentInstructionPos()
		return false, NewRuntimeError1(pos1, pos2, e.what)
	}
	return res, nil
}

func (vm *VM) generateClosure(a *datatype.Value, i Instruction) {
	call, proto := getCallInfoAndProto(vm)
	aProto := proto.GetChildFunction(int(GetParamBx(i)))
	a.Type = datatype.ValueTClosure
	a.Closure = vm.state.NewClosure()
	a.Closure.SetPrototype(aProto)

	// Prepare all upvalues
	newClosure := a.Closure
	closure := call.Func.Closure
	count := aProto.GetUpvalueCount()
	for i := 0; i < count; i++ {
		upvalueInfo := aProto.GetUpvalue(i)
		if upvalueInfo.ParentLocal {
			reg := vPointerAdd(call.Register, upvalueInfo.RegisterIndex)
			if reg.Type != datatype.ValueTUpvalue {
				upvalue := vm.state.NewUpvalue()
				upvalue.SetValue(reg)
				reg.Type = datatype.ValueTUpvalue
				reg.Upvalue = upvalue
				newClosure.AddUpvalue(upvalue)
			} else {
				newClosure.AddUpvalue(reg.Upvalue)
			}
		} else {
			// Get upvalue from parent upvalue list
			upvalue := closure.GetUpvalue(upvalueInfo.RegisterIndex)
			newClosure.AddUpvalue(upvalue)
		}
	}
}

func (vm *VM) copyVarArg(a *datatype.Value, i Instruction) {
	call, proto := getCallInfoAndProto(vm)
	arg := vPointerAdd(call.Func, 1)
	// totalArgs represents the number of datatype.Value between call.Register and arg
	totalArgs := int((uintptr(unsafe.Pointer(call.Register)) - uintptr(unsafe.Pointer(arg))) /
		unsafe.Sizeof(datatype.Value{}))
	varargCount := totalArgs - proto.FixedArgCount()

	arg = vPointerAdd(arg, proto.FixedArgCount())
	expectCount := int(GetParamsBx(i))
	if expectCount == datatype.ExpValueCountAny {
		for i := 0; i < varargCount; i++ {
			*a = *arg
			arg = vPointerAdd(arg, 1)
		}
		vm.state.stack.SetNewTop(a)
	} else {
		i := 0
		for ; i < varargCount && i < expectCount; i++ {
			*a = *arg
			arg = vPointerAdd(arg, 1)
		}
		for ; i < expectCount; i++ {
			a.SetNil()
			a = vPointerAdd(a, 1)
		}
	}
}

func (vm *VM) return_(a *datatype.Value, i Instruction) {
	// Set stack top when return value count i is fixed
	retValueCount := int(GetParamsBx(i))
	if retValueCount != datatype.ExpValueCountAny {
		vm.state.stack.Top = vPointerAdd(a, retValueCount)
	}

	if vm.state.calls.Len() == 0 {
		panic("assert")
	}
	call := vm.state.calls.Back().Value.(*CallInfo)

	src := a
	dst := call.Func

	expectResult := call.ExpectResult
	resultCount := int(uintptr(unsafe.Pointer(vm.state.stack.Top)) - uintptr(unsafe.Pointer(a)))
	if expectResult == datatype.ExpValueCountAny {
		for i := 0; i < resultCount; i++ {
			*dst = *src
			src = vPointerAdd(src, 1)
		}
	} else {
		i := 0
		count := int(math.Min(float64(expectResult), float64(resultCount)))
		for i < count {
			*dst = *src
			dst = vPointerAdd(dst, 1)
			src = vPointerAdd(src, i)
			i++
		}
		// No enough results for expect results, set remain as nil
		for i < expectResult {
			dst.SetNil()
			dst = vPointerAdd(dst, 1)
			i++
		}
	}

	// Set new top and pop current CallInfo
	vm.state.stack.SetNewTop(dst)
	vm.state.calls.Remove(vm.state.calls.Back())
}

func (vm *VM) concat(dst, op1, op2 *datatype.Value) error {
	if op1.Type == datatype.ValueTString && op2.Type == datatype.ValueTString {
		dst.Str = vm.state.GetString(op1.Str.GetStdString() + op2.Str.GetCStr())
	} else if op1.Type == datatype.ValueTString && op2.Type == datatype.ValueTNumber {
		dst.Str = vm.state.GetString(op1.Str.GetCStr() + numberToStr(op2))
	} else if op1.Type == datatype.ValueTNumber && op2.Type == datatype.ValueTString {
		dst.Str = vm.state.GetString(numberToStr(op2) + op2.Str.GetCStr())
	} else {
		pos1, pos2 := vm.getCurrentInstructionPos()
		return NewRuntimeError4(pos1, pos2, *op1, *op2, "concat")
	}
	return nil
}

func (vm *VM) forInit(var_, limit, step *datatype.Value) error {
	if var_.Type != datatype.ValueTNumber {
		pos1, pos2 := vm.getCurrentInstructionPos()
		return NewRuntimeError2(pos1, pos2, *var_, "'for' init", "number")
	}

	if limit.Type != datatype.ValueTNumber {
		pos1, pos2 := vm.getCurrentInstructionPos()
		return NewRuntimeError2(pos1, pos2, *var_, "'for' limit", "number")
	}

	if step.Type != datatype.ValueTNumber {
		pos1, pos2 := vm.getCurrentInstructionPos()
		return NewRuntimeError2(pos1, pos2, *var_, "'for' step", "number")
	}
	return nil
}

// Debug help functions
func (vm *VM) getOperandNameAndScope(a *datatype.Value) (string, string) {
	call, proto := getCallInfoAndProto(vm)
	reg := int(uintptr(unsafe.Pointer(a)) - uintptr(unsafe.Pointer(call.Register)))
	instruction := iPointerAdd(call.Instruction, -1)
	base := proto.GetOpCodes()
	pc := int(uintptr(unsafe.Pointer(instruction)) - uintptr(unsafe.Pointer(base)))
	unknownName := "?"
	scopeGlobal := "global"
	scopeLocal := "local"
	scopeUpvalue := "upvalue"
	scopeTable := "table member"
	scopeNil := ""

	// Search last instruction which dst register is reg,
	// and get the name base on the instruction
	for uintptr(unsafe.Pointer(instruction)) > uintptr(unsafe.Pointer(base)) {
		instruction = iPointerAdd(instruction, -1)
		switch GetOpCode(*instruction) {
		case OpTypeGetGlobal:
			if reg == GetParamA(*instruction) {
				index := GetParamBx(*instruction)
				key := proto.GetConstValue(int(index))
				if key.Type == datatype.ValueTString {
					return key.Str.GetCStr(), scopeGlobal
				} else {
					return unknownName, scopeNil
				}
			}
		case OpTypeMove:
			if reg == GetParamA(*instruction) {
				src := GetParamB(*instruction)
				name := proto.SearchLocalVar(src, pc)
				if name != nil {
					return name.GetCStr(), scopeLocal
				} else {
					return unknownName, scopeNil
				}
			}
		case OpTypeGetUpvalue:
			if reg == GetParamA(*instruction) {
				index := GetParamB(*instruction)
				upvalueInfo := proto.GetUpvalue(index)
				return upvalueInfo.Name.GetCStr(), scopeUpvalue
			}
		case OpTypeGetTable:
			if reg == GetParamC(*instruction) {
				key := GetParamB(*instruction)
				keyReg := vPointerAdd(call.Register, key)
				if keyReg.Type == datatype.ValueTString {
					return keyReg.Str.GetCStr(), scopeTable
				} else {
					return unknownName, scopeTable
				}
			}
		}
	}
	return unknownName, scopeNil
}

func (vm *VM) getCurrentInstructionPos() (string, int) {
	call, proto := getCallInfoAndProto(vm)
	index := uintptr(unsafe.Pointer(call.Instruction)) - uintptr(unsafe.Pointer(proto.GetOpCodes())) - 1
	return proto.GetModule().GetCStr(), proto.GetInstructionLine(int(index))
}

func (vm *VM) checkType(v *datatype.Value, vType int, op string) error {
	if v.Type != vType {
		return vm.reportTypeError(v, op)
	}
	return nil
}

func (vm *VM) checkArithType(v1, v2 datatype.Value, op string) error {
	if v1.Type != datatype.ValueTNumber || v2.Type != datatype.ValueTNumber {
		pos1, pos2 := vm.getCurrentInstructionPos()
		return NewRuntimeError4(pos1, pos2, v1, v2, op)
	}
	return nil
}

func (vm *VM) checkInequalityType(v1, v2 datatype.Value, op string) error {
	if (v1.Type != v2.Type) ||
		(v1.Type != datatype.ValueTNumber && v1.Type != datatype.ValueTString) {
		pos1, pos2 := vm.getCurrentInstructionPos()
		return NewRuntimeError4(pos1, pos2, v1, v2, op)
	}
	return nil
}

func (vm *VM) checkTableType(t, k datatype.Value, op, desc string) error {
	if (t.Type == datatype.ValueTTable) ||
		(t.Type == datatype.ValueTUserData && t.UserDate.GetMetaTable() != nil) {
		return nil
	}

	n, s := vm.getOperandNameAndScope(&t)
	pos1, pos2 := vm.getCurrentInstructionPos()
	var keyName string
	if k.Type == datatype.ValueTString {
		keyName = k.Str.GetCStr()
	} else {
		keyName = "?"
	}
	opDesc := fmt.Sprintf("%s table key '%s' %s", op, keyName, desc)
	return NewRuntimeError3(pos1, pos2, t, n, s, opDesc)
}

func (vm *VM) reportTypeError(v *datatype.Value, op string) error {
	n, s := vm.getOperandNameAndScope(v)
	pos1, pos2 := vm.getCurrentInstructionPos()
	return NewRuntimeError3(pos1, pos2, *v, n, s, op)
}

func (vm *VM) Execute() {
	if vm.state.calls.Len() == 0 {
		panic("assert")
	}

	for vm.state.calls.Len() != 0 {
		// If current stack frame is a frame of a c function,
		// do not continue execute instructions, just return
		call := vm.state.calls.Back().Value.(*CallInfo)
		if call.Func.Type == datatype.ValueTCFunction {
			return
		}

		if err := vm.executeFrame(); err != nil {
			panic(err)
		}
	}
}
