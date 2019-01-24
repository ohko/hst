package main

import (
	"fmt"
	"time"

	"github.com/ohko/hst"
)

func main() {
	s := hst.New(nil)
	s.HandleFunc("/", func(c *hst.Context) {
		fmt.Fprintln(c.W, "hello")
	})
	s.HandleFunc("/ip", func(c *hst.Context) {
		fmt.Fprintln(c.W, c.R.RemoteAddr)
	})
	s.HandleFunc("/time", func(c *hst.Context) {
		fmt.Fprintln(c.W, time.Now().Format("2006-01-02 15:04:05"))
	})
	// s.ListenAutoCert(".https", "xx.example.com")
	s.ListenHTTP(":8080")
}
