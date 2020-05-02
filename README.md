# WIP

# GoVue

Framework for embedding Vue.js SPA in single Go binary.

## Uses

- [becky](https://github.com/tv42/becky) for assets embedding,
- [afero](https://github.com/spf13/afero) for virtual filesystem,
- [amber](https://github.com/eknkc/amber) for templating,
- [gorilla's mux](https://github.com/gorilla/mux) for routing,
- [rest-layer](https://github.com/rs/rest-layer) for REST queries (WIP),

## Usage

Add to your `.go` file (adapt `sh` command when running in Windows):

```golang
//go:generate go build github.com/tv42/becky↵
//go:generate sh -c "./becky -wrap Wrap -var _ *.js *.css *.amber"↵
```

Import library:

```golang
import "github.com/marf41/GoVue"
```

Create new instance:

```golang
var vue GoVue.Vue
```

Minimal `main`:

```golang
func main() {
    vue.Title = "Test page"
    vue.Lang = "en" // set in "html" tag
    vue.Start(":8080")
}
```

Run it:

```sh
go build -o build . && ./build
```

# Known issues

- becky's, when using `dev` tag, has wrong paths for files

---

To be continued...
