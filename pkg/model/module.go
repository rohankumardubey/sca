package model

//Module represente a module collecting data //TODO add a Close method that close chan
type Module interface {
	ID() string
	//New(map[string]string) Module
	//Flags() *pflag.FlagSet
	Event() <-chan string
	GetData() interface{}
}
