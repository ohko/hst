package main

import "github.com/ohko/hst"

// Auto ...
type Auto struct{}

// Hello ...
func (o *Auto) Hello(ctx *hst.Context) {
	ctx.JSON(200, "auto::hello")
}
