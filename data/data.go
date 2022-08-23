package data

type Data interface {
	data()
}

type dataType struct{}

func (*dataType) data() {}

type String struct {
	dataType
	Value string
}

type Error struct {
	dataType
	Message string
}

type Array struct {
	dataType
	Values []Data
}

type Integer struct {
	dataType
	Value int64
}

type Null struct {
	dataType
}

var NULL = &Null{}

var Test Data = &String{}
