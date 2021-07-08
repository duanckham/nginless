package nginless

// doCall ...
func (n *Nginless) doJSON(d *D, parameters []interface{}) *D {
	s := parameters[0].(string)

	d.res.Header().Add("Content-Type", "application/json")
	d.res.Write([]byte(s))

	return d
}
