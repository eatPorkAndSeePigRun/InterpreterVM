package datatype

import "InterpreterVM/Source/vm"

const ExpvalueCountAny = -1

type CFunctionType func(state *vm.State) int64

type ValueT int64

const (
	ValueTNil = iota
	ValueTBool
	ValueTNumber
	ValueTObj
	ValueTString
	ValueTClosure
	ValueTUpvalue
	ValueTTable
	ValueTUserData
	ValueTCFunction
)

// Value type of vm
type Value struct {
	Obj      GCObject
	Str      *String
	Closure  *Closure
	Upvalue  *Upvalue
	Table    *Table
	UserDate *UserData
	CFunc    CFunctionType
	Num      float64
	BValue   bool

	Type ValueT
}

func NewValueObj() Value {
	return Value{Obj: nil, Type: ValueTNil}
}

func NewValueBValue(bValue bool) Value {
	return Value{BValue: bValue, Type: ValueTBool}
}

func NewValueNum(num float64) Value {
	return Value{Num: num, Type: ValueTNumber}
}

func NewValueString(str *String) Value {
	return Value{Str: str, Type: ValueTString}
}

func NewValueClosure(closure *Closure) Value {
	return Value{Closure: closure, Type: ValueTClosure}
}

func NewValueUpvalue(upvalue *Upvalue) Value {
	return Value{Upvalue: upvalue, Type: ValueTUpvalue}
}

func NewValueTable(table *Table) Value {
	return Value{Table: table, Type: ValueTTable}
}

func NewValueUserData(userData *UserData) Value {
	return Value{UserDate: userData, Type: ValueTUserData}
}

func NewValueCFunction(cFunc CFunctionType) Value {
	return Value{CFunc: cFunc, Type: ValueTCFunction}
}

func (v *Value) SetNil() {
	v.Obj = nil
	v.Type = ValueTNil
}

func (v *Value) SetBool(bValue bool) {
	v.BValue = bValue
	v.Type = ValueTBool
}

func (v *Value) IsNil() bool {
	return v.Type == ValueTNil
}

func (v *Value) IsFalse() bool {
	return (v.Type == ValueTNil) || (v.Type == ValueTBool && !v.BValue)
}

func (v *Value) Accept(visitor GCObjectVisitor) {
	switch v.Type {
	case ValueTNil, ValueTBool, ValueTNumber, ValueTCFunction:
	case ValueTObj:
		v.Obj.Accept(visitor)
	case ValueTString:
		v.Str.Accept(visitor)
	case ValueTClosure:
		v.Closure.Accept(visitor)
	case ValueTUpvalue:
		v.Upvalue.Accept(visitor)
	case ValueTTable:
		v.Table.Accept(visitor)
	case ValueTUserData:
		v.UserDate.Accept(visitor)
	}
}

func (v *Value) TypeName() string {
	return v.GetTypeName(v.Type)
}

func (v *Value) isEqual(v1 *Value) bool {
	if v.Type != v1.Type {
		return false
	}

	switch v.Type {
	case ValueTNil:
		return true
	case ValueTBool:
		return v.BValue == v1.BValue
	case ValueTNumber:
		return v.Num == v1.Num
	case ValueTObj:
		return v.Obj == v1.Obj
	case ValueTString:
		return v.Str == v1.Str
	case ValueTClosure:
		return v.Closure == v1.Closure
	case ValueTUpvalue:
		return v.Upvalue == v1.Upvalue
	case ValueTTable:
		return v.Table == v1.Table
	case ValueTUserData:
		return v.UserDate == v1.UserDate
		//case ValueTCFunction:
		//	return v.CFunc == v1.CFunc
	}

	return false
}

func (v *Value) GetTypeName(vType ValueT) string {
	switch vType {
	case ValueTNil:
		return "nil"
	case ValueTBool:
		return "bool"
	case ValueTNumber:
		return "number"
	case ValueTCFunction:
		return "C-Function"
	case ValueTString:
		return "string"
	case ValueTClosure:
		return "function"
	case ValueTUpvalue:
		return "upvalue"
	case ValueTTable:
		return "table"
	case ValueTUserDate:
		return "userdata"
	default:
		return "unknown type"
	}
}
