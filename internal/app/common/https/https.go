package https

import (
	"crypto/tls"
	"net"
)

// Listener ...
type Listener interface {
	Accept() (net.Conn, error)
	Close() error
	Addr() net.Addr
	Bind(listener net.Listener)
	LoadPairs(pairs [][2]string)
}

type listenerImpl struct {
	config   *tls.Config
	listener net.Listener
}

// New ...
func New() Listener {
	return &listenerImpl{}
}

// Accept ...
func (l *listenerImpl) Accept() (net.Conn, error) {
	if l.config == nil {
		panic("https listener config missing")
	}

	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}

	return tls.Server(conn, l.config), nil
}

// Close ...
func (l *listenerImpl) Close() error {
	return nil
}

// Addr ...
func (l *listenerImpl) Addr() net.Addr {
	return nil
}

// Bind ...
func (l *listenerImpl) Bind(listener net.Listener) {
	l.listener = listener
}

// LoadPairs ...
func (l *listenerImpl) LoadPairs(pairs [][2]string) {
	var err error

	l.config = &tls.Config{}
	l.config.Certificates = make([]tls.Certificate, len(pairs))

	for i, pair := range pairs {
		if len(pair) < 2 {
			panic("certificate pair missing")
		}

		l.config.Certificates[i], err = tls.LoadX509KeyPair(pair[0], pair[1])
		if err != nil {
			panic("load certificate pairs failed")
		}
	}
}
