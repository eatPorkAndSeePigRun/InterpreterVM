package datatype

type Upvalue struct {
	gcObjectField
	value Value
}

func NewUpvalue() *Upvalue {
	return &Upvalue{}
}

func (u *Upvalue) Accept(v GCObjectVisitor) {
	if v.VisitUpvalue(u) {
		u.value.Accept(v)
	}
}

func (u *Upvalue) SetValue(value *Value) {
	u.value = *value
}

func (u *Upvalue) GetValue() *Value {
	return &u.value
}
