package hst

// Group 路由分组
type Group struct {
	prefix string
	hst    *HST
}

// Group 路由分组
func (o *HST) Group(name string, handler ...HandlerFunc) *Group {
	if handler[0] != nil {
		handleFunc(o, name, handler...)
	}
	return &Group{
		hst:    o,
		prefix: name,
	}
}

// HandleFunc ...
// Example:
//		HandleFunc("/", func(c *hst.Context){}, func(c *hst.Context){})
func (o *Group) HandleFunc(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o.hst, o.prefix+pattern, handler...)
}
