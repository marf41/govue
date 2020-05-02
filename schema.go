package govue

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"

	"github.com/rs/rest-layer/resource"
	"github.com/rs/rest-layer/rest"
	"github.com/rs/rest-layer/schema"
	"github.com/rs/rest-layer/schema/query"

	sqlStorage "github.com/marf41/rest-layer-sql"
	_ "github.com/mattn/go-sqlite3" // registering the sqlite3 driver as a database driver
)

// Schema struct for rest-layer `Schema`
type Schema struct {
	Name     string
	Fields   schema.Fields
	Resource *resource.Resource
	Storer   resource.Storer
}

// AddSchema adds rest-layer's `Schema` to GoVue's `Index` and handlers
func (v Vue) AddSchema(name string, fields schema.Fields, modes ...resource.Mode) {
	conf := resource.DefaultConf
	if len(modes) > 0 {
		conf = resource.Conf{
			AllowedModes: modes,
		}
	}
	sch := schema.Schema{Fields: fields}
	r, s := addBind(name, v.DB, sch, v.GetIndex(), conf)
	schema := Schema{name, fields, r, s}
	if v.Schemas == nil {
		v.Schemas = make(map[string]Schema)
	}
	v.Schemas[name] = schema
}

// SetDB sets path to sqlite database file
func (v Vue) SetDB(path string) {
	v.DB = "file:" + path
}

// GetIndex returns `resource.Index` for `rest-layer`
func (v Vue) GetIndex() resource.Index {
	if v.Index == nil {
		v.Index = resource.NewIndex()
	}
	return v.Index
}

// SetAPI sets API endpoint at path (default: `/api/`)
func (v Vue) SetAPI(path ...string) {
	p := "/api/"
	if len(path) > 0 {
		p = path[0]
	}
	if v.Schemas == nil || len(v.Schemas) == 0 {
		return
	}
	api, err := rest.NewHandler(v.GetIndex())
	if err != nil {
		log.Fatalf("Invalid API configuration: %s", err)
	}
	v.GetRouter().PathPrefix(p).Handler(http.StripPrefix(p, api))
}

// SetBasicAuth sets basic authentication before root page,
// checking login against rest-layer store
func (v Vue) SetBasicAuth(realm, schema, userField, passField string) {
	v.Auth = func(w http.ResponseWriter, r *http.Request) bool {
		user, pass, ok := r.BasicAuth()
		log.Printf("Checking login for: %q / %q.\n", user, pass)
		any := false
		q := new(query.Query)
		us, err := v.Schemas[schema].Storer.Find(context.TODO(), q)
		if err != nil {
			werr(err, w)
			return false
		}
		for _, iu := range us.Items {
			u, ok := iu.GetField(userField).(string)
			p, pok := iu.GetField(passField).(string)
			uid, nok := iu.GetField("id").(string)
			if ok && pok &&
				subtle.ConstantTimeCompare([]byte(u), []byte(user)) == 1 &&
				subtle.ConstantTimeCompare([]byte(p), []byte(pass)) == 1 {
				log.Printf("Login as user: %q (%q).\n", user, uid)
				any = true
				if nok {
					v.User = uid
				}
			}
		}
		if !ok || !any {
			if ok {
				log.Printf("Invalid login: %q.\n", user)
			}
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%q", realm))
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return false
		}
		return true
	}
}

func addBind(name, db string, sch schema.Schema, index resource.Index, conf resource.Conf) (*resource.Resource, resource.Storer) {
	cfg := sqlStorage.Config{5, map[string]string{}}
	s, err := sqlStorage.NewHandler("sqlite3", db, name, &cfg)
	if err != nil {
		log.Fatalf("[%s] Error connecting database: %s", name, err)
	}
	err = s.Create(context.TODO(), &sch)
	if err != nil {
		log.Fatalf("[%s] Error creating table: %s", name, err)
	}
	return index.Bind(name, sch, s, conf), s
}
