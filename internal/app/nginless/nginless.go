package nginless

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/duanckham/nginless/internal/app/common/https"
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
	httpsListener := m.Match(cmux.Any())

	// Start HTTP service.
	go n.startHTTP(httpListener)
	// Start HTTPS service.
	go n.startHTTPS(httpsListener)

	m.Serve()
}

func (n *Nginless) startHTTP(l net.Listener) {
	http.HandleFunc("/", n.handleTraffic)
	http.Serve(l, nil)
}

func (n *Nginless) startHTTPS(l net.Listener) {
	// Pick certificate pairs from all rules.
	pairs := make([][2]string, len(n.router.Certificates))

	for i, v := range n.router.Certificates {
		if v.Certificate != "" && v.Key != "" {
			pairs[i] = [2]string{v.Certificate, v.Key}
		}
	}

	// Create HTTPS listener.
	listener := https.New()

	// Bind original listener.
	listener.Bind(l)

	// Load all certificate pairs.
	listener.LoadPairs(pairs)

	http.Serve(listener, nil)
}

func (n *Nginless) handleTraffic(w http.ResponseWriter, req *http.Request) {
	matched, handler := n.router.Match(req)

	// Write nginless sign into header.
	w.Header().Del("x-nginless-version")
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
