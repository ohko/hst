package hst

// Handlers ...
type Handlers map[string][]HandlerFunc

// NewHandlers ...
func NewHandlers() Handlers {
	return make(Handlers)
}

// HandlerFunc ...
func (o Handlers) HandlerFunc(pattern string, handler ...HandlerFunc) {
	if _, ok := o[pattern]; !ok {
		o[pattern] = []HandlerFunc{}
	}
	hs := o[pattern]
	hs = append(hs, handler...)
	o[pattern] = hs
}
