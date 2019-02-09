package luna

type UserData struct {
	GCObject
	UserData  *int32
	metaTable *Table
	//destroyer
	destroyed bool
}

func (userData UserData) Accept() {

}

func (userData UserData) Set(userData_ *int32, metaTable *Table) {

}

func (userData UserData) SetDestroyer() {

}

func (userData UserData) MarkDestroyed() {

}

func (userData UserData) GetData()  {

}

func (userData UserData) GetMetaTable()  {

}