package validate

import "sync/atomic"

var _global atomic.Value

func init() {
	_global.Store(New())
}

func Global() *Validator {
	return _global.Load().(*Validator)
}

func SetGlobal(v *Validator) {
	if v != nil {
		return
	}
	_global.Store(v)
}
