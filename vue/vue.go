package vue
//go:generate go run github.com/tv42/becky vue.js vuetify.js materialdesignicons.min.css
//go:generate go run github.com/tv42/becky vuetifymin.css mdi.css mdif.woff2
//go:generate go run github.com/tv42/becky MarqueeTextumd.js MarqueeText.css
//go:generate go run github.com/tv42/becky materialdesigniconsfont.woff2
import "net/http"
import "github.com/gorilla/mux"

func js(a asset) http.Handler { return a }
func css(a asset) http.Handler { return a }
func woff2(a asset) http.Handler { return a }

func Handlers(r *mux.Router) {
    r.Handle("/vue.js", vue)
    r.Handle("/mdi.css", mdi)
    r.Handle("/mdif.woff2", mdif)
    r.Handle("/vuetify.js", vuetify)
    r.Handle("/vuetify.min.css", vuetifymin)
    r.Handle("/materialdesignicons.min.css", materialdesignicons)
    r.Handle("/MarqueeText.css", MarqueeText)
    r.Handle("/MarqueeTextumd.js", MarqueeTextumd)
    r.Handle("/fonts/materialdesignicons-webfont.woff2", materialdesigniconsfont)
}
