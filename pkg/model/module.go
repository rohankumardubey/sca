package model

//Module represente a module collecting data //TODO add a Close method tha close chan
type Module interface {
	ID() string
	Event() <-chan string
	GetData() interface{}
}
