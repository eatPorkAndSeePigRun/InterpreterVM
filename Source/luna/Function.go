package luna

type Function struct {
}

type Closure struct {
	prototype *Function
	upValues  []*UpValue
}

func (closure Closure) Accept()  {

}

func (closure Closure) GetPrototype()  {

}

func (closure Closure) SetPrototype()  {

}

func (closure Closure) AddUpValue()  {

}

func (closure Closure) GetUpValue()  {

}
