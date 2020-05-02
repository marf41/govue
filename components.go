package govue

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

/*
Vue.component('v-footer-cmp', { template: '#tmp-footer', data() { return { text: '' } },
    methods: { set(s) { this.text = s } },
    computed: { value() { var y = new Date().getFullYear(); return this.text || ('&copy; ' + y) } } })
Vue.component('v-snack', { template: '#tmp-snack', data() { return { text: '', shw: false, color: '', timeout: 3000 } },
    methods: { set(s, c) { this.text = s; this.color = c || ''; this.shw = true },
				show() { this.shw = true } } })
*/

type vcsMethod struct {
	Arguments  int
	Body       string
	Parameters []string
}

type VueComponentScript struct {
	Data    map[string]string
	Methods map[string]vcsMethod
}

func (vcs *VueComponentScript) checkData() {

	if vcs.Data == nil {
		vcs.Data = make(map[string]string)
	}
}

func (vcs *VueComponentScript) checkMethods() {
	if vcs.Methods == nil {
		vcs.Methods = make(map[string]vcsMethod)
	}
}

func (vcs *VueComponentScript) NewData(name, def string, opts ...string) {
	vcs.checkData()
	vcs.checkMethods()
	vcs.Data[name] = def
	switch len(opts) {
	case 1: // setter
		vcs.Methods[opts[0]] = vcsMethod{1, "this.%s = %s", []string{name, "arg1"}}
		fallthrough
	default:
	}
}

func (vcs VueComponentScript) String() string {
	tdm := `
	data() { return { {{ range .Data }}{{ . }}{{ end }} } }, {{ if len .Methods }}
	methods: { {{ range .Methods }}
		{{ . }}{{ end }}
	}, {{ end }}`
	data := make(map[string][]template.HTML)
	data["Data"] = []template.HTML{}
	data["Methods"] = []template.HTML{}
	for name, def := range vcs.Data {
		data["Data"] = append(data["Data"], template.HTML(fmt.Sprintf("%s: %s,", name, def)))
	}
	for name, m := range vcs.Methods {
		var ms strings.Builder
		ms.WriteString(name)
		ms.WriteString("(")
		for i := 1; i <= m.Arguments; i++ {
			if i > 1 {
				ms.WriteString(", ")
			}
			ms.WriteString(fmt.Sprintf("arg%d", i))
		}
		ms.WriteString(") { ")
		params := make([]interface{}, len(m.Parameters))
		for i, v := range m.Parameters {
			params[i] = v
		}
		ms.WriteString(fmt.Sprintf(m.Body, params...))
		ms.WriteString("; }, ")
		data["Methods"] = append(data["Methods"], template.HTML(ms.String()))
	}
	tpl, err := template.New("vcs").Parse(tdm)
	ferr("VCS", err)
	var b bytes.Buffer
	err = tpl.Execute(&b, data)
	return b.String()
}

// AddComponents adds new component to provided Vue instance,
// injecting VueComponentScript in added component's script
func (vcs VueComponentScript) AddComponent(v *Vue, name, template string, script ...string) {
	s := append([]string{vcs.String()}, script...)
	v.AddComponent(name, template, s...)
}

func (v *Vue) addComponents() {
	// snackCmp := VueComponentScript{}
	// snackCmp.newData("")
	v.AddComponent("snack", `
v-snackbar[v-model="shw"][:timeout="timeout"][:color="color"]
	span[v-html="text"]
	v-btn[@click="shw = false"][v-text="close"]`, `
	data() { return { text: '', shw: false, color: '', timeout: 3000, close: 'X' } },
	methods: {
		set(s,c) { this.text = s; this.color = c || ''; this.shw = true },
		show() { this.shw = true }
	}`)

	footCmp := VueComponentScript{}
	footCmp.NewData("text", "''", "set")
	/*
			v.AddComponent("footer-cmp", `span.text-xs-right[:v-html="value"]`, `
		data() { return { text: '' } },
		methods: { set(s) { this.text = s } },
		computed: { value() { var y = new Date().getFullYear(); return this.text || ('&copy; ' + y) } }
			`)
	*/
	footCmp.AddComponent(v, "footer-cmp", `span.text-xs-right[:v-html="value"]`,
		"\t"+`computed: { value() { var y = new Date().getFullYear(); return this.text || ('&copy; ' + y) } }`)

}
