package nginless

import "net"

// Listeners ...
type Listeners interface {
	Accept() (net.Conn, error)
	Close() error
	Addr() net.Addr
	Add(listener net.Listener)
}

// ListenersImpl ...
type ListenersImpl struct {
	// listeners []net.Listener
}

// NewListeners ...
func NewListeners() Listeners {
	return ListenersImpl{
		// listeners: []net.Listener{},
	}
}

// Add ...
func (l ListenersImpl) Add(listener net.Listener) {

}

// Accept ...
func (l ListenersImpl) Accept() (net.Conn, error) {

	return nil, nil
}

// Close ...
func (l ListenersImpl) Close() error {
	return nil
}

// Addr ...
func (l ListenersImpl) Addr() net.Addr {
	return nil
}
