package hst

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strings"
)

// BasicAuth http验证
func BasicAuth(user, pass string) HandlerFunc {
	return func(c *Context) {
		success := false
		defer func() {
			if !success {
				c.W.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				c.W.WriteHeader(http.StatusUnauthorized)
				c.Close()
			}
		}()

		basicAuthPrefix := "Basic "
		auth := c.R.Header.Get("Authorization")
		if strings.HasPrefix(auth, basicAuthPrefix) {
			if payload, err := base64.StdEncoding.DecodeString(auth[len(basicAuthPrefix):]); err == nil {
				pair := bytes.SplitN(payload, []byte(":"), 2)
				if len(pair) == 2 && bytes.Equal(pair[0], []byte(user)) && bytes.Equal(pair[1], []byte(pass)) {
					success = true
				}
			}
		}
	}
}
