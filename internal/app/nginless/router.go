package nginless

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/duanckham/go-pcre"
	"gopkg.in/yaml.v2"
)

var (
	lb    = "("[0]
	rb    = ")"[0]
	comma = ","[0]
	space = " "[0]
)

var singleParameterDoes = []string{"json"}

var reHeaderTest = regexp.MustCompile(`^header\.`)

// Config is the config yaml file structure.
//
// <example: something.yaml>
// rules:
//   - rule: testing.test:.*/https
//     do: call(reverse_proxy)
//
//   - rule: testing.test:.*/hi
//     do: call(reverse_proxy)
//
//   - rule: testing.test:.*/balancing
//     do: balancing(https://www.google.com, https://www.twitter.com)
//
//   - rule: testing.test:.*/proxy
//     do: proxy(https://www.google.com, https://www.twitter.com)
//
//   - rule: AppleWebKit\/\d+\.\d+\s
//     test: header.user-agent
//     do: json({"success":true,"message":"matching header.user-agent"})
//
// certificates:
//   - certificate: ./examples/testing.test.crt
//     key: ./examples/testing.test.key
type Config struct {
	Rules        []Rule        `yaml:"rules"`
	Certificates []Certificate `yaml:"certificates"`
}

// Router ...
type Router struct {
	Rules        []Rule
	Certificates []Certificate
	Handlers     []Handler
}

// Certificate ...
type Certificate struct {
	Certificate string `yaml:"certificate"`
	Key         string `yaml:"key"`
}

// Rule ...
type Rule struct {
	Condition interface{} `yaml:"rule"`
	Test      string      `yaml:"test"`
	Do        interface{} `yaml:"do"`
}

// Target $A.$B, eg: header.user-agent.
type Target struct {
	A string
	B string
}

// Handler ...
type Handler struct {
	Regex  []pcre.Regexp
	Steps  []Step
	Target Target
}

// Step ...
type Step struct {
	Source     string
	Action     string
	Parameters []interface{}
}

// NewRouter ...
func NewRouter(filePath string) *Router {
	r := &Router{
		Rules:    []Rule{},
		Handlers: []Handler{},
	}

	r.loadConfig(filePath)
	r.parse()

	return r
}

// Match ...
func (r *Router) Match(req *http.Request) (bool, Handler) {
	s := ""

	for _, v := range r.Handlers {
		for _, regex := range v.Regex {
			switch v.Target.A {
			case "url":
				s = req.Host + req.URL.String()
			case "header":
				s = req.Header.Get(v.Target.B)
			}

			if regex.MatchString(s, 0) {
				return true, v
			}
		}
	}

	return false, Handler{}
}

// loadConfig ...
func (r *Router) loadConfig(filePath string) {
	// Read router config file.
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic("router config file do not exist.")
	}

	config := Config{}

	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		panic("parse router config file failed.")
	}

	r.Rules = config.Rules
	r.Certificates = config.Certificates
}

// parse ...
func (r *Router) parse() {
	for _, v := range r.Rules {
		handler := Handler{
			Regex: []pcre.Regexp{},
			Steps: []Step{},
		}

		// Process condition.
		switch reflect.ValueOf(v.Condition).Kind() {
		case reflect.String:
			handler.Regex = append(handler.Regex, pcre.MustCompile(v.Condition.(string), 0))
		case reflect.Slice:
			for _, c := range v.Condition.([]interface{}) {
				handler.Regex = append(handler.Regex, pcre.MustCompile(c.(string), 0))
			}
		default:
		}

		// Process target.
		switch {
		case reHeaderTest.MatchString(v.Test):
			t := strings.Split(v.Test, ".")
			if len(t) != 2 {
				panic("the matching condition for header is invalid, the correct `test` should be `header.$something`")
			}

			handler.Target = Target{"header", t[1]}

		default:
			handler.Target = Target{"url", ""}
		}

		// Process action.
		switch reflect.ValueOf(v.Do).Kind() {
		case reflect.String:
			handler.Steps = parseSteps([]interface{}{v.Do})
		case reflect.Slice:
			handler.Steps = parseSteps(v.Do.([]interface{}))
		}

		r.Handlers = append(r.Handlers, handler)
	}
}

// parseSteps ...
func parseSteps(does []interface{}) []Step {
	steps := []Step{}

	for _, v := range does {
		steps = append(steps, parseDoString(v.(string)))
	}

	return steps
}

// parseDoString ...
func parseDoString(s string) Step {
	bracketsFound := false
	action := ""
	parameters := []interface{}{}
	t := ""

	for i := 0; i < len(s); i++ {
		if s[i] == space && !isSingleParameter(action) {
			continue
		}

		if s[i] != lb && s[i] != rb {
			if bracketsFound {
				if s[i] == comma && !isSingleParameter(action) {
					parameters = append(parameters, t)
					t = ""
				} else {
					t += string(s[i])
				}
			} else {
				action += string(s[i])
			}

			continue
		}

		// Process last parameter.
		if s[i] == rb && len(t) > 0 {
			parameters = append(parameters, t)
		}

		bracketsFound = true
	}

	return Step{
		Source:     s,
		Action:     action,
		Parameters: parameters,
	}
}

func isSingleParameter(action string) bool {
	for _, v := range singleParameterDoes {
		if v == action {
			return true
		}
	}

	return false
}
