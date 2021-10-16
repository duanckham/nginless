package nginless

import (
	"net/http"
)

// D ...
type D struct {
	req      *http.Request
	res      http.ResponseWriter
	finished bool
}

func (d *D) returnInternalServerError() *D {
	if !d.finished {
		d.res.WriteHeader(http.StatusInternalServerError)
		d.finished = true
	}

	return d
}

func (d *D) done() *D {
	d.finished = true
	return d
}

func (n *Nginless) do(d *D, step Step) *D {
	switch step.Action {
	// eg:
	// proxy($remote_address)
	case "proxy":
		return n.doProxy(d, step.Parameters)

	// eg:
	// balancing($remote_address, ...$remote_address)
	case "balancing":
		return n.doBalancing(d, step.Parameters)

	// eg:
	// call($tengo_script)
	case "call":
		return n.doCall(d, step.Parameters)

	// eg:
	// json({"a":"b"})
	case "json":
		return n.doJSON(d, step.Parameters)
	}

	return d
}
