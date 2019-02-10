package luna

const ExpValueCountAny = -1

type CFunctionType func(*State) int64

type ValueT int64

const (
	ValueTNil = iota
	ValueTBool
	ValueTNumber
	ValueTObj
	ValueTString
	ValueTClosure
	ValueTUpValue
	ValueTTable
	ValueTUserDate
	ValueTCFunction
)

// Value type of luna
type Value struct {
	Obj      *GCObject
	Str      *String
	Closure  *Closure
	UpValue  *UpValue
	Table    *Table
	UserDate *UserData
	CFunc    *CFunctionType
	Num      float64
	BValue   bool

	Type ValueT
}

func (v Value) SetNil() {
	v.Obj = nil
	v.Type = ValueTNil
}

func (v Value) SetBool(bValue bool) {
	v.BValue = bValue
	v.Type = ValueTBool
}

func (v Value) IsNil() bool {
	return v.Type == ValueTNil
}

func (v Value) IsFalse() bool {
	return (v.Type == ValueTNil) || (v.Type == ValueTBool && !v.BValue)
}

func (v Value) Accept(visitor GCObjectVisitor) {
	switch v.Type {
	case ValueTNil, ValueTBool, ValueTNumber, ValueTCFunction:
	case ValueTObj:
		v.Obj.Accept(visitor)
	case ValueTString:
		v.Str.Accept(visitor)
	case ValueTClosure:
		v.Closure.Accept(visitor)
	case ValueTUpValue:
		v.UpValue.Accept(visitor)
	case ValueTTable:
		v.Table.Accept(visitor)
	case ValueTUserDate:
		v.UserDate.Accept(visitor)
	}
}

func (v Value) TypeName() string {
	return v.GetTypeName(v.Type)
}

func (v Value) GetTypeName(vType ValueT) string {
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
	case ValueTUpValue:
		return "upvalue"
	case ValueTTable:
		return "table"
	case ValueTUserDate:
		return "userdata"
	default:
		return "unknown type"
	}
}
