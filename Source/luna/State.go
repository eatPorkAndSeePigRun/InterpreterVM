package luna

type State struct {
}

func (state State) DoModule(moduleName string) {

}

func (state State) DoString(str, name string) {

}

func (state State) GetString(str string) *String {
	return &String{}
}
