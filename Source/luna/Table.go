package luna

type Array []Value
type Hash map[Value]Value

type Table struct {
	GCObject
	array *Array
	hash  *Hash
}

func (table Table) appendAndMergeFromHashToArray(value Value) {

}

func (table Table) appendToArray(value Value) {

}

func (table Table) mergeFromHashToArray() {

}

func (table Table) moveHashToArray(key Value) bool {
	return false
}

func (table Table) Accept(v *GCObjectVisitor) {

}

func (table Table) SetArrayValue(index int64, value Value) bool {
	return false
}

func (table Table) InsertArrayValue(index int64, value Value) bool {
	return false
}

func (table Table) EraseArrayValue(index int64) bool {
	return false
}

func (table Table) SetValue(key, value Value) {

}

func (table Table) GetValue(key Value) Value {
	return Value{}
}

func (table Table) FirstKeyValue(key, value Value) bool {
	return false
}

func (table Table) NextKeyValue(key, nextKey, nextValue Value) bool {
	return false
}

func (table Table) ArraySize() int64 {

}
