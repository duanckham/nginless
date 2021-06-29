package nginless

import (
	"net/http"

	"go.uber.org/zap"
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
	n.logger.Info(".do", zap.String("rule", step.Source), zap.String("action", step.Action), zap.Any("parameters", step.Parameters))

	switch step.Action {
	// proxy($remote_address)
	case "proxy":
		return n.doProxy(d, step.Parameters)
	// call($tengo_script)
	case "call":
		return n.doCall(d, step.Parameters)
	}

	return d
}
