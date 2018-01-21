package hst

import "time"

// HST ...
type HST interface {
	Shutdown(time.Duration)
	HandleFunc(string, ...HandlerFunc)
	Static(string, string)
	HandlePfx(string, string)
	Favicon()
	Listen() error
}

// HandlerFunc ...
type HandlerFunc func(*Context)
