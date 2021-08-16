package nginless

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
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

const socking = "/tmp/nginless.sock"

// Run ...
func (n *Nginless) Run() {

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

	// http.HandleFunc("/", n.handleTraffic)
	// http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)

	var (
		listener net.Listener
		err      error
	)

	// Create listeners.
	listeners := NewListeners()

	err = os.RemoveAll(socking)
	if err != nil {
		panic(err)
	}

	listener, err = net.Listen("unix", socking)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for _, port := range n.ports {
		go func(port string) {

			fmt.Println("#1", port)

			// Listen local port.
			l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
			if err != nil {
				panic(err)
			}

			// Add listen into listeners.
			listeners.Add(l)

			// for {
			// 	conn, err := l.Accept()

			// 	fmt.Printf("new connection from %s (%s)\n", conn.RemoteAddr(), port)

			// 	if err != nil {
			// 		fmt.Println("create connection failed from", conn.RemoteAddr())
			// 		continue
			// 	}

			// 	go func() {
			// 		defer conn.Close()

			// 		entry, err := net.Dial("unix", socking)
			// 		if err != nil {
			// 			fmt.Println("error dialing remote", err)
			// 			return
			// 		}
			// 		defer entry.Close()

			// 		fmt.Println("#3")

			// 		closer := make(chan struct{}, 2)

			// 		go copy(closer, entry, conn)
			// 		go copy(closer, conn, entry)

			// 		<-closer

			// 		fmt.Println("connection complete", conn.RemoteAddr())
			// 	}()
			// }

		}(port)

	}

	fmt.Println("#0")

	m := cmux.New(listeners.(net.Listener))

	httpListener := m.Match(cmux.Any())

	go func() {
		http.HandleFunc("/", n.handleTraffic)
		http.Serve(httpListener, nil)
	}()

	m.Serve()
}

func copy(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{} // connection is closed, send signal to stop proxy
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
