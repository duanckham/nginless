package nginless

import (
	"io/ioutil"
	"reflect"

	"github.com/duanckham/go-pcre"
	"gopkg.in/yaml.v2"
)

// Router ...
type Router struct {
	Rules    []Rule
	Handlers []Handler
}

// Rule ...
type Rule struct {
	Condition interface{} `yaml:"rule"`
	Do        interface{} `yaml:"do"`
}

// Handler ...
type Handler struct {
	Regex []pcre.Regexp
	Steps []Step
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
func (r *Router) Match(uri string) (bool, Handler) {
	for _, v := range r.Handlers {
		for _, regex := range v.Regex {
			if regex.MatchString(uri, pcre.ANCHORED) {
				return true, v
			}
		}
	}

	return false, Handler{}
}

// LoadConfig ...
func (r *Router) loadConfig(filePath string) {
	// Read router config file.
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic("router config file do not exist.")
	}

	err = yaml.Unmarshal(bytes, &r.Rules)
	if err != nil {
		panic("parse router config file failed.")
	}
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
			handler.Regex = append(handler.Regex, pcre.MustCompile(v.Condition.(string), pcre.ANCHORED))
		case reflect.Slice:
			for _, c := range v.Condition.([]interface{}) {
				handler.Regex = append(handler.Regex, pcre.MustCompile(c.(string), pcre.ANCHORED))
			}
		default:
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

var (
	lb    = "("[0]
	rb    = ")"[0]
	comma = ","[0]
	space = " "[0]
)

// parseDoString ...
func parseDoString(s string) Step {
	bracketsFound := false
	action := ""
	parameters := []interface{}{}
	t := ""

	for i := 0; i < len(s); i++ {
		if s[i] == space {
			continue
		}

		if s[i] != lb && s[i] != rb {
			if bracketsFound {
				if s[i] == comma && action != "json" {
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
