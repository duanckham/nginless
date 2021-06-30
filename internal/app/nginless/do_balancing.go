package nginless

import (
	"math/rand"
)

// doBalancing forward the request to the random address.
// eg:
// balancing(https://www.google.com, https://www.youtube.com)
func (n *Nginless) doBalancing(d *D, parameters []interface{}) *D {
	if len(parameters) == 0 {
		return d.returnInternalServerError()
	}

	return n.doProxy(d, []interface{}{
		parameters[rand.Intn(len(parameters))],
	})
}
