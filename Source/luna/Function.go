package luna

type Function struct {
	GCObject
}

type Closure struct {
	GCObject
	prototype *Function
	upValues  []*UpValue
}

func (closure Closure) Accept(visitor GCObjectVisitor) {

}

func (closure Closure) GetPrototype() {

}

func (closure Closure) SetPrototype() {

}

func (closure Closure) AddUpValue() {

}

func (closure Closure) GetUpValue() {

}
