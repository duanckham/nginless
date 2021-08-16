package nginless

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
)

// Nginless ...
type Nginless struct {
	version string
	ports   []string
	logger  *zap.Logger
	router  *Router
	actions string
}

// Options ...
type Options struct {
	Version string
	Logger  *zap.Logger
}

// New ...
func New(options Options) *Nginless {
	routerPath := flag.String("r", "", "Router config file path")
	actionPath := flag.String("a", "", "Action files path")
	ports := flag.String("p", "80", "Listening ports")

	flag.Parse()

	router := NewRouter(*routerPath)

	fmt.Printf("*     version: %s\n", options.Version)
	fmt.Printf("* router path: %s\n", *routerPath)
	fmt.Printf("* action path: %s\n", *actionPath)
	fmt.Printf("*       ports: %s\n", *ports)

	return &Nginless{
		version: options.Version,
		ports:   strings.Split(*ports, ","),
		logger:  options.Logger,
		router:  router,
		actions: *actionPath,
	}
}

// Run ...
func (n *Nginless) Run() {
	// Create listeners.
	listeners := NewListeners()
	defer listeners.Close()

	for _, port := range n.ports {
		// Listen local port.
		l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
		if err != nil {
			panic(err)
		}

		// Add listen into listeners.
		go listeners.Bind(l)
	}

	m := cmux.New(listeners.(net.Listener))

	httpListener := m.Match(cmux.HTTP1Fast())

	go n.startHTTP(httpListener)

	m.Serve()
}

func (n *Nginless) startHTTP(l net.Listener) {
	http.HandleFunc("/", n.handleTraffic)
	http.Serve(l, nil)
}

func (n *Nginless) startHTTPS(l net.Listener) {
	// go func() {
	// 	// Handle HTTPS
	// 	l, err := net.Listen("tcp", ":443")
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	for {
	// 		conn, err := l.Accept()
	// 		if err != nil {
	// 			fmt.Println("???", err)
	// 			continue
	// 		}

	// 		go n.handleConnection(conn)
	// 	}
	// }()
}

func (n *Nginless) handleTraffic(w http.ResponseWriter, req *http.Request) {
	matched, handler := n.router.Match(req)

	// Write nginless sign into header.
	w.Header().Set("x-nginless-version", n.version)

	d := &D{req, w, false}

	if !matched {
		d.returnInternalServerError()
		return
	}

	// Run steps.
	for i, step := range handler.Steps {
		start := time.Now()

		n.logger.Info(
			".handleTraffic",
			zap.Int("step", i),
			zap.String("uri", req.URL.String()),
			zap.String("rule", step.Source),
			zap.String("action", step.Action),
			zap.Any("parameters", step.Parameters),
			zap.Duration("took", time.Since(start)),
		)

		d = n.do(d, step)
	}
}
