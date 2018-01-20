// Example:
// 		type myhandler struct{}
// 		func (h *myhandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 			fmt.Fprintln(w, "Hello world!")
// 		}
// Example:
// 		handle := http.NewServeMux()
// 		handle.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 			fmt.Fprintln(w, "Hello world!")
// 		})

package hst

import (
	"net/http"
)

// HST ...
type HST interface {
	Listen() error
	HandleFunc(string, http.HandlerFunc)
}
