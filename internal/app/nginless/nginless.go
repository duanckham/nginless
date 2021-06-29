package nginless

import (
	"flag"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// Nginless ...
type Nginless struct {
	port    int32
	logger  *zap.Logger
	router  *Router
	actions string
}

// Options ...
type Options struct {
	Logger *zap.Logger
}

// New ...
func New(options Options) *Nginless {
	routerPath := flag.String("r", "", "Router config file path")
	actionPath := flag.String("a", "", "Action files path")
	port := flag.Int("p", 80, "Listening port")

	flag.Parse()

	router := NewRouter(*routerPath)

	fmt.Printf("* router path: %s\n", *routerPath)
	fmt.Printf("* action path: %s\n", *actionPath)
	fmt.Printf("* port       : %d\n", *port)

	return &Nginless{
		port:    int32(*port),
		logger:  options.Logger,
		router:  router,
		actions: *actionPath,
	}
}

// Run ...
func (n *Nginless) Run() {
	http.HandleFunc("/", n.handleTraffic)
	http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)
}

func (n *Nginless) handleTraffic(w http.ResponseWriter, req *http.Request) {
	uri := req.Host + req.URL.String()
	matched, handler := n.router.Match(uri)

	d := &D{req, w, false}

	if !matched {
		d.returnInternalServerError()
		return
	}

	// Run steps.
	for _, step := range handler.Steps {
		d = n.do(d, step)
	}
}
