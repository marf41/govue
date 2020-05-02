package roboto
//go:generate go run github.com/tv42/becky RobotoRegular.woff2 roboto.css
import "net/http"
import "github.com/gorilla/mux"

func css(a asset) http.Handler { return a }
func woff2(a asset) http.Handler { return a }

func Handlers(r *mux.Router) {
    r.Handle("/roboto.css", roboto)
    r.Handle("/fonts/Regular/Roboto-Regular.woff2", RobotoRegular)
}
