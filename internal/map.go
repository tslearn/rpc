package internal

var emptyRTMap = RTMap{}

type RTMap struct {
	rt    Runtime
	items []posRecord
}
