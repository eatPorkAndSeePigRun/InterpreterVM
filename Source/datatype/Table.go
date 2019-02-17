package datatype

import "math"

func isInt(x float64) bool {
	return math.Floor(x) == x
}

// Table has array part and hash table part.
type Table struct {
	gcObjectField
	array *array // array part of table
	hash  *hash  // hash table of table
}

func NewTable() *Table {
	return &Table{}
}

type array []Value
type hash map[Value]Value

// Combine AppendToArray and MergeFromHashToArray
func (t *Table) appendAndMergeFromHashToArray(value Value) {
	t.appendToArray(value)
	t.mergeFromHashToArray()
}

// Append value to array.
func (t *Table) appendToArray(value Value) {
	if t.array == nil {
		t.array = new(array)
	}
	*t.array = append(*(t.array), value)
}

// Try to move values from hash to array which keys start from
// ArraySize() + 1
func (t *Table) mergeFromHashToArray() {
	index := t.ArraySize()
	index++
	key := Value{Num: float64(index), Type: ValueTNumber}

	for t.moveHashToArray(key) {
		index++
		key.Num = float64(index)
	}
}

// Move hash table key-value pair to array which key is number and key
// fit with array, return true if move success.
func (t *Table) moveHashToArray(key Value) bool {
	if t.hash == nil {
		return false
	}

	value, ok := (*t.hash)[key]
	if !ok {
		return false
	}

	t.appendToArray(value)
	delete(*t.hash, value)
	return true
}

func (t *Table) Accept(v GCObjectVisitor) {
	if v.VisitTable(t) {
		// Visit all array members
		if t.array != nil {
			for _, value := range *t.array {
				value.Accept(v)
			}
		}

		// Visit all keys and values in hash table.
		if t.hash != nil {
			for key, value := range *t.hash {
				key.Accept(v)
				value.Accept(v)
			}
		}
	}
}

// Set array value by index, return true if success.
// 'index' start from 1, if 'index' == ArraySize() + 1,
// then append value to array.
func (t *Table) SetArrayValue(index int64, value Value) bool {
	if index < 1 {
		return false
	}

	arraySize := t.ArraySize()
	if index > arraySize+1 {
		return false
	}

	if index == arraySize+1 {
		t.appendAndMergeFromHashToArray(value)
	} else {
		(*t.array)[index-1] = value
	}

	return true
}

// If 'index' == ArraySize() + 1, then append value to array,
// otherwise shifting up all values which start from 'index',
// and insert value to 'index' of array.
// Return true when insert success.
func (t *Table) InsertArrayValue(index int64, value Value) bool {
	if index < 1 {
		return false
	}

	arraySize := t.ArraySize()
	if index > arraySize+1 {
		return false
	}

	if index == arraySize+1 {
		t.appendAndMergeFromHashToArray(value)
	} else {
		// Insert value
		*t.array = append(*t.array, Value{})
		copy((*t.array)[index:], (*t.array)[index-1:])
		(*t.array)[index] = value
		// Try to merge from hash to array
		t.mergeFromHashToArray()
	}

	return true
}

// Erase the value by 'index' in array if 'index' is legal,
// shifting down all values which start from 'index' + 1.
// Return true when erase success.
func (t *Table) EraseArrayValue(index int64) bool {
	if index < 1 || index > t.ArraySize() {
		return false
	}

	*t.array = append((*t.array)[:index], (*t.array)[index+1:]...)
	return true
}

// Add key-value into table.
// If key is number and key fit with array, then insert into array,
// otherwise insert into hash table.
func (t *Table) SetValue(key, value Value) {
	// Try array part
	if key.Type == ValueTNumber && isInt(key.Num) {
		if t.SetArrayValue(int64(key.Num), value) {
			return
		}
	}

	// Hash part
	if t.hash == nil {
		// If value is nil and hash part is not existed, then do nothing
		if value.IsNil() {
			return
		}
		t.hash = new(hash)
	}

	v, ok := (*t.hash)[key]
	if ok {
		// If value is nil, then just erase the element
		if value.IsNil() {
			delete(*t.hash, v)
		} else {
			(*t.hash)[key] = value
		}
	} else {
		// If key is not existed and value is not nil, then insert it
		if !value.IsNil() {
			(*t.hash)[key] = value
		}
	}

}

// Get Value of key from array first,
// if key is number, then get the value from array when key number
// is fit with array as index, otherwise try search in hash table.
// Return value is 'nil' if 'key' is not existed.
func (t *Table) GetValue(key Value) Value {
	// Get from array first
	if key.Type == ValueTNumber && isInt(key.Num) {
		index := int64(key.Num)
		if index >= 1 && index <= t.ArraySize() {
			return (*t.array)[index-1]
		}
	}

	// Get from hash table
	if t.hash != nil {
		v, ok := (*t.hash)[key]
		if ok {
			return v
		}
	}

	// key not exist
	return Value{}
}

// Get first key-value pair of table, return true if table is not empty.
func (t *Table) FirstKeyValue(key, value *Value) bool {
	// array part
	if t.ArraySize() > 0 {
		key.Num = 1 // first element index
		key.Type = ValueTNumber
		*value = (*t.array)[0]
		return true
	}

	// hash part
	if t.hash != nil && len(*t.hash) == 0 {
		for k, v := range *t.hash {
			*key = k
			*value = v
			return true
		}
	}

	return false
}

// Get the next key-value pair by current 'key', return false if there
// is no key-value pair any more.
func (t *Table) NextKeyValue(key, nextKey, nextValue *Value) bool {
	// array part
	if key.Type == ValueTNumber && isInt(key.Num) {
		index := int64(key.Num) + 1
		if index >= 1 && index <= t.ArraySize() {
			nextValue.Num = float64(index)
			nextValue.Type = ValueTNumber
			*nextValue = (*t.array)[index-1]
			return true
		}
	}

	// hash part
	v, ok := (*t.hash)[*key]
	if ok {
		isV := false
		for nk, nv := range *t.hash {
			if nv.isEqual(&v) {
				isV = true
			}
			if isV == true {
				*nextKey = nk
				*nextValue = nv
				return true
			}
		}
	} else if !ok && len(*t.hash) != 0 {
		for k1, v1 := range *t.hash {
			*nextKey = k1
			*nextValue = v1
			return true
		}
	}

	return false
}

// Return the number of array part elements.
func (t *Table) ArraySize() int64 {
	if t.array != nil {
		return int64(len(*t.array))
	} else {
		return 0
	}
}
