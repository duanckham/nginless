package nginless

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

// doProxy forward the request to the specified address.
// eg:
// proxy(https://www.google.com)
// proxy(http://1.2.3.4:8000)
func (n *Nginless) doProxy(d *D, parameters []interface{}) *D {
	if len(parameters) == 0 {
		return d.returnInternalServerError()
	}

	remote, err := url.Parse(parameters[0].(string))
	if err != nil {
		return d.returnInternalServerError()
	}

	client := &http.Client{}

	uri := fmt.Sprintf("%s://%s%s", remote.Scheme, remote.Host, d.req.RequestURI)
	req, err := http.NewRequest(d.req.Method, uri, d.req.Body)
	if err != nil {
		n.logger.Error(".doProxy create new request failed", zap.Error(err))
	}

	// Copy request headers.
	for k, headers := range d.req.Header {
		for _, item := range headers {
			req.Header.Add(k, item)
		}
	}

	// Request.
	res, err := client.Do(req)
	if err != nil {
		return d.returnInternalServerError()
	}

	// Copy response headers.
	for k, headers := range res.Header {
		for _, item := range headers {
			d.res.Header().Add(k, item)
		}
	}

	// Copy response body.
	_, err = io.Copy(d.res, res.Body)
	if err != nil {
		return d.returnInternalServerError()
	}

	d.done = true

	return d
}