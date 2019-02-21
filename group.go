package hst

// Group 路由分组
type Group struct {
	prefix string
	hst    *HST
}

// Group 路由分组
func (o *HST) Group(name string, handler ...HandlerFunc) *Group {
	if handler[0] != nil {
		handleFunc(o, "", name, handler...)
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
	return handleFunc(o.hst, "", o.prefix+pattern, handler...)
}

// GET ...
func (o *Group) GET(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o.hst, "GET", pattern, handler...)
}

// POST ...
func (o *Group) POST(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o.hst, "POST", pattern, handler...)
}

// PUT ...
func (o *Group) PUT(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o.hst, "PUT", pattern, handler...)
}

// PATCH ...
func (o *Group) PATCH(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o.hst, "PATCH", pattern, handler...)
}

// DELETE ...
func (o *Group) DELETE(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o.hst, "DELETE", pattern, handler...)
}

// OPTIONS ...
func (o *Group) OPTIONS(pattern string, handler ...HandlerFunc) *HST {
	return handleFunc(o.hst, "OPTIONS", pattern, handler...)
}
