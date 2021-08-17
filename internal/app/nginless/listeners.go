package nginless

import (
	"net"
)

// ListenersAddr ...
type ListenersAddr interface {
	net.Addr
}

// listenersAddrImpl ...
type listenersAddrImpl struct {
}

// NewListenersAddr ...
func NewListenersAddr() ListenersAddr {
	return &listenersAddrImpl{}
}

// Network ...
func (l *listenersAddrImpl) Network() string {
	return "memory"
}

// String ...
func (l *listenersAddrImpl) String() string {
	return "[::]:0"
}

// Listeners ...
type Listeners interface {
	Accept() (net.Conn, error)
	Close() error
	Addr() net.Addr
	Bind(listener net.Listener)
}

// listenersImpl ...
type listenersImpl struct {
	listeners []net.Listener
	c         chan accept
}

type accept struct {
	conn net.Conn
	err  error
}

// NewListeners ...
func NewListeners() Listeners {
	return &listenersImpl{
		c: make(chan accept),
	}
}

// Accept ...
func (l *listenersImpl) Accept() (net.Conn, error) {
	accept := <-l.c
	return accept.conn, accept.err
}

// Close ...
func (l *listenersImpl) Close() error {
	var e error

	for _, listener := range l.listeners {
		err := listener.Close()
		if err != nil {
			e = err
		}
	}

	return e
}

// Addr ...
func (l *listenersImpl) Addr() net.Addr {
	return NewListenersAddr()
}

// Bind ...
func (l *listenersImpl) Bind(listener net.Listener) {
	// Insert into listeners array.
	l.listeners = append(l.listeners, listener)

	for {
		conn, err := listener.Accept()
		l.c <- accept{conn, err}
	}
}
