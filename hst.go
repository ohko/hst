package hst

// HST ...
type HST interface {
	HandleFunc(string, ...HandlerFunc)
	Static(string, string)
	HandlePfx(string, string)
	Favicon()
	Listen() error
}

// HandlerFunc ...
type HandlerFunc func(*Context)
