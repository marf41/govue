// Package govue implements handlers, templates, and schemas for
// creating SPAs in Vue.js, using ace templates and REST layer
package govue

//go:generate go build github.com/tv42/becky
//go:generate sh -c "./becky -wrap IntWrap -var _ *.css *.js *.amber"

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/rest-layer/resource"

	// "github.com/yosssi/ace"
	"github.com/eknkc/amber"

	"github.com/spf13/afero"

	"github.com/marf41/GoVue/roboto"
	"github.com/marf41/GoVue/vue"
)

const uid = "216672ef-bf07-400f-9eb7-9dcac1a2de0d"

type tmpContents struct {
	Head       []string
	Body       []string
	App        []string
	Components []string
	Scripts    []string
}

type tmpAssets struct {
	js      map[string]http.Handler
	css     map[string]http.Handler
	tmp     map[string]string
	json    map[string]http.Handler
	init    bool
	ModTime time.Time
}

var assets tmpAssets

// Wrap adds becky assets to `assets` struct
func Wrap(name, content, etag string) bool {
	prefixWrap("", asset{name, content, etag})
	return true
}

// IntWrap adds internal becky assets
func IntWrap(a asset) bool { prefixWrap(uid, a); return true }

func prefixWrap(prefix string, a asset) {
	wrap := filepath.Ext(a.Name)
	wrap = wrap[1:]
	name := prefix + a.Name
	assets.ModTime = time.Now()
	// log.Printf("New asset: %q %q %q.\n", wrap, a.Name, prefix)
	switch wrap {
	case "js":
		if assets.js == nil {
			assets.js = make(map[string]http.Handler)
		}
		assets.js[name] = a
	case "css":
		if assets.css == nil {
			assets.css = make(map[string]http.Handler)
		}
		assets.css[name] = a
	case "jsonf":
		fallthrough
	case "json":
		if assets.json == nil {
			assets.json = make(map[string]http.Handler)
		}
		assets.json[name] = a
	case "amb":
		fallthrough
	case "amber":
		if assets.tmp == nil {
			assets.tmp = make(map[string]string)
		}
		if a.Name == "base.amber" {
			assets.tmp[a.Name] = a.Content
		} else {
			assets.tmp[name] = a.Content
		}
	default:
		log.Printf("Unknown asset type - file %q.\n", a.Name)
	}
}

// Component struct for parsed Vue component
type vueParsedComponent struct {
	Name     string
	Template string
	Script   string
}

type vueComponent struct {
	Name     string
	Template string
	Script   string
}

// Vue data struct
type vueData struct {
	Lang       string
	Scripts    []string
	Styles     []string
	User       string
	Title      string
	Includes   []string
	UserData   interface{}
	Components []vueParsedComponent
}

// Vue settings struct
type Vue struct {
	Lang               string
	Title              string
	User               string
	Auth               func(w http.ResponseWriter, r *http.Request) bool
	Scripts            []string
	Styles             []string
	Data               interface{}
	Router             *mux.Router
	Index              resource.Index
	Schemas            map[string]Schema
	DB                 string
	Debug              bool
	FS                 http.FileSystem
	Components         map[string]*vueComponent
	parsedComponents   []vueParsedComponent
	hasEditorFieldType map[vueFieldType]bool
}

// NewVue creates new `Vue` instance
func NewVue() Vue {
	v := Vue{}
	return v
}

// GetRouter returns `mux.Router` that will be used with `Start`
func (v *Vue) GetRouter() *mux.Router {
	if v.Router == nil {
		v.Router = mux.NewRouter()
	}
	return v.Router
}

// Start web server, with logging handler, on supplied port
func (v Vue) Start(port string) {
	r := v.GetRouter()
	v.Handlers(r)
	v.SetAPI()
	go http.ListenAndServe(port, handlers.LoggingHandler(os.Stdout, r))
	log.Printf("Ready at %s.\n", port)

	for {
		runtime.Gosched()
	}
}

// Handlers sets handlers in passed gorilla mux router
func (v *Vue) Handlers(r *mux.Router) {
	vue.Handlers(r)
	roboto.Handlers(r)
	for name, asset := range assets.js {
		r.Handle("/"+name, asset)
	}
	for name, asset := range assets.css {
		r.Handle("/"+name, asset)
	}
	for name, asset := range assets.json {
		r.Handle("/"+name, asset)
	}
	log.Printf("Creating filesystem for %d files.\n", len(assets.tmp))
	fs := afero.NewMemMapFs()
	fs.Mkdir("/tmp", 0755)
	for name, asset := range assets.tmp {
		err := afero.WriteFile(fs, "/tmp/"+name, []byte(asset), 0644)
		if err != nil {
			log.Println(err)
		}
	}
	v.FS = afero.NewHttpFs(afero.NewReadOnlyFs(fs))
	if v.Debug {
		afero.Walk(fs, "/", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Println(err)
				return err
			}
			if path == "/" {
				return nil
			}
			if info.IsDir() {
				log.Printf("\t %s:\n", path)
			} else {
				log.Printf("\t \t %s\n", path)
			}
			return nil
		})
	}

	v.addComponents()

	r.HandleFunc("/", v.handler())
}

// AddComponent adds new Vue component
func (v *Vue) AddComponent(name, template string, script ...string) {
	// log.Printf("Adding component %q.\n", name)
	if v.Components == nil {
		v.Components = make(map[string]*vueComponent)
	}
	v.Components[name] = &vueComponent{name, template, strings.Join(script, "\n")}
	// v.CheckComponents()
}

// CheckComponents returns number of registered Vue components
func (v Vue) CheckComponents() int {
	if v.Components == nil {
		log.Println("Uninitialized component map!")
		return 0
	}
	list := []string{}
	for name := range v.Components {
		list = append(list, name)
	}
	if v.Debug {
		log.Printf("%d registered components: %v.\n", len(list), list)
	}
	return len(list)
}

// Main handler for root
func (v Vue) handler() http.HandlerFunc {
	v.CheckComponents()
	if v.Debug {
		if v.FS != nil {
			f, err := v.FS.Open("/tmp/" + uid + "first-base.amber")
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("Base OK.\n")
				f.Close()
			}
		} else {
			log.Println("Virtual filesystem not set!")
		}
	}
	c := amber.New()
	c.VirtualFilesystem = v.FS
	err := c.ParseFile("/tmp/main.amber")
	if ferr("TP", err) {
		return nil
	}
	tpl, err := c.Compile()
	if ferr("TC", err) {
		return nil
	}

	v.parsedComponents = []vueParsedComponent{}
	sb := "\n\t<script id=%q type=%q>\n"
	se := "\t</script>"
	for name, component := range v.Components {
		if v.Debug {
			log.Printf("Registering component: %q.\n", name)
		}
		n := "tmp-" + name
		sbx := fmt.Sprintf(sb, n, "text/x-template")
		sbs := fmt.Sprintf(sb, "js-"+n, "application/javascript")
		cc := amber.New()
		cc.Parse(strings.TrimSpace(component.Template))
		ct, err := cc.CompileString()
		if !logerr("VUEC", err) {
			// log.Printf("Component %q:\n%q\n.", name, component.Template)
			ct = strings.Join(strings.Split(ct, "\n"), "\t\t\n")
			cs := fmt.Sprintf("\tVue.component(%q, { template: %q, %s })\n",
				"v-"+name, "#"+n, component.Script)
			cmp := vueParsedComponent{n, sbx + ct + se, sbs + cs + se}
			v.parsedComponents = append(v.parsedComponents, cmp)
		}
	}
	// log.Printf("%d components registered.", len(data.Components))
	return func(w http.ResponseWriter, r *http.Request) {
		if v.Auth != nil {
			log.Println("Checking auth...")
			if v.Auth != nil && !v.Auth(w, r) {
				return
			}
		}
		data := vueData{}
		data.User = v.User
		data.Title = v.Title
		data.Lang = v.Lang
		data.Scripts = []string{}
		data.Styles = append(v.Styles, "style.css")
		hasScript := false
		for name := range assets.js {
			if name == "script.js" {
				hasScript = true
			} else {
				if !strings.Contains(name, uid) ||
					(strings.Contains(name, uid) && strings.Contains(name, "axios")) {
					data.Scripts = append(data.Scripts, name)
				}
			}
		}
		data.Scripts = append(data.Scripts, v.Scripts...)
		data.Scripts = append(data.Scripts, uid+"components.js")
		if hasScript {
			data.Scripts = append(data.Scripts, "script.js")
		}
		data.Scripts = append(data.Scripts, uid+"vuescript.js")

		data.Components = v.parsedComponents
		err = tpl.Execute(w, data)
		if werr(err, w) {
			return
		}
	}
}
