package hst

import "time"

// HST ...
type HST interface {
	Shutdown(waitTime time.Duration)
	HandleFunc(pattern string, handler ...HandlerFunc)
	Static(partten, path string)
	HandlePfx(partten, pfxPath string)
	Favicon()
	Listen() error
}

// HandlerFunc ...
type HandlerFunc func(*Context)
