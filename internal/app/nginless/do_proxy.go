package nginless

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/valyala/bytebufferpool"
	"go.uber.org/zap"
)

// doProxy forward the request to the specified address.
// eg:
// proxy(https://www.google.com)
// proxy(http://1.2.3.4:8000)
// refs:
// https://sourcegraph.com/github.com/golang/go/-/blob/src/net/http/httputil/reverseproxy.go?L214
func (n *Nginless) doProxy(d *D, parameters []interface{}) *D {
	if len(parameters) == 0 {
		return d.returnInternalServerError()
	}

	remote, err := url.Parse(parameters[0].(string))
	if err != nil {
		return d.returnInternalServerError()
	}

	// Create request client.
	client := &http.Client{}

	// Build up URI.
	uri := fmt.Sprintf("%s://%s%s", remote.Scheme, remote.Host, d.req.RequestURI)

	// Build request.
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

	// Send to remote server.
	res, err := client.Do(req)
	if err != nil {
		n.logger.Error(".doProxy send request to remote failed", zap.Error(err))
		return d.returnInternalServerError()
	}

	// Copy response headers.
	for k, headers := range res.Header {
		if strings.ToLower(k) == "x-nginless-version" {
			continue
		}

		for _, item := range headers {
			d.res.Header().Set(k, item)
		}
	}

	// Write status code.
	// refs:
	// https://stackoverflow.com/a/26097384
	d.res.WriteHeader(res.StatusCode)

	// Copy response body.
	bb := bytebufferpool.Get()
	n.copyBuffer(d.res, res.Body, bb.B)

	defer res.Body.Close()
	defer bytebufferpool.Put(bb)

	if err != nil {
		n.logger.Error(".doProxy copy response failed", zap.Int64("res.ContentLength", res.ContentLength), zap.Error(err))
		return d.returnInternalServerError()
	}

	return d.done()
}

func (n *Nginless) copyBuffer(dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	if len(buf) == 0 {
		buf = make([]byte, 32*1024)
	}

	var written int64

	for {
		nr, rerr := src.Read(buf)
		if rerr != nil && rerr != io.EOF && rerr != context.Canceled {
			n.logger.Error(".doProxy read error during body copy", zap.Error(rerr))
		}

		if nr > 0 {
			nw, werr := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if werr != nil {
				return written, werr
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}

		if rerr != nil {
			if rerr == io.EOF {
				rerr = nil
			}

			return written, rerr
		}
	}
}
