package luna

type UpValue struct {
	GCObject
	value Value
}

func (u UpValue) Accept(v GCObjectVisitor) {
	if v.VisitUpValue(&u) {
		u.value.Accept(v)
	}
}

func (u *UpValue) SetValue(value *Value) {
	u.value = *value
}

func (u UpValue) GetValue() *Value {
	return &u.value
}
