package nginless

import (
	"net/http"
)

// D ...
type D struct {
	req  *http.Request
	res  http.ResponseWriter
	done bool
}

func (d *D) returnInternalServerError() *D {
	if !d.done {
		d.res.WriteHeader(http.StatusInternalServerError)
		d.done = true
	}

	return d
}

func (n *Nginless) do(d *D, step Step) *D {
	switch step.Action {
	// eg:
	// proxy($remote_address)
	case "proxy":
		return n.doProxy(d, step.Parameters)
	// eg:
	// call($tengo_script)
	case "call":
		return n.doCall(d, step.Parameters)
	}

	return d
}
