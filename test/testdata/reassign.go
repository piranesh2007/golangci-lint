//golangcitest:args -Ereassign
package testdata

import (
	"io"
	"net/http"
)

func reassignTest() {
	http.DefaultClient = nil
	io.EOF = nil // want `reassigning variable EOF in other package io`
}
