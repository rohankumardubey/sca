package model

import "github.com/spf13/pflag"

//Module represente a module collecting data //TODO add a Close method tha close chan
type Module interface {
	ID() string
	New() Module
	Flags() *pflag.FlagSet
	Event() <-chan string
	GetData() interface{}
}
