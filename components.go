package govue

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
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

type vueFieldType string

// VueEditorFields define `Editor` fields and their types
type VueEditorFields struct {
	Label  string
	Model  string
	Type   vueFieldType
	Cols   uint
	Items  string
	Suffix string
	Prefix string
}

// AddEditorFieldType adds new component for editing given field type
func (v *Vue) AddEditorFieldType(fieldType vueFieldType, template string, script ...string) {
	if v.hasEditorFieldType == nil {
		v.hasEditorFieldType = make(map[vueFieldType]bool)
	}
	v.hasEditorFieldType[fieldType] = true
	if len(script) == 0 {
		script = []string{"\tdata() { return {} },"}
	}
	s := append(script, "\tprops: [ 'item', 'field' ],")
	v.AddComponent("editor-field-"+string(fieldType),
		template, s...)
}

// AddEditor adds new Vue component for CRUD editing
func (v *Vue) AddEditor(name string, fields []VueEditorFields) {
	for _, f := range fields {
		if !v.hasEditorFieldType[f.Type] {
			log.Printf("No such editor field type: %q.\n", f)
		}
	}
	v.AddComponent("editor-"+name, `
	`)

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

	v.AddComponent("dialog-editor", `
v-dialog[v-model="show"][scrollable][persistent][max-width="50%"]
	v-card
		v-card-title[v-text="isNew ? addTitle : editTitle"]
		v-divider
		v-card-text
			v-form
				v-container
					v-row
						component
						[:is="'v-editor-field-' + (field.type || '')"]
						[:item="item"]
						[:field="field"]
						[v-for="(field, n) in fields"]
						[key="(field.type || '') + field.label + n"]
			v-card[v-if="selIcon"]
				v-card-title[v-text="$root.selectIconTitle || 'Select icon'"]
				v-card-text
					v-container
						v-row
							v-cols[cols="12"]
								v-text-field
								[icon-prepend="search"]
								[outlined]
								[v-model="selIconFilter"]
								[label="$root.selectIconSearch || 'Search...'"]
						v-row[no-gutters][align="center"][justify="center"]
							v-col
							[cols="1"]
							[v-for="i in $root.matFilterIcons"]
							[:key="i"]
								v-card[outlined][tile]
									v-icon
									[v-text="i"]
									[@click="$set(item, icon, i)"]
		v-card-actions
			template[v-if="icon"]
				v-btn
				[text]
				[@click="selIcon = !selIcon"]
				[v-text="selIcon ? ($root.closeIcons || 'Close icons') : ($root.changeIcon : 'Change icon')"]
				v-btn
				[text]
				[@click="item[icon] = null"]
				[v-text="$root.removeIcon || 'Remove icon'"]
				v-spacer
				v-btn[text][@click="cancel"][v-text="$root.closeButton || 'Close'"]
				v-btn
				[text]
				[@click="save"]
				[v-text="isNew ? ($root.addButton || 'Add') : ($root.saveButton || 'Save')"]
`, `
	data() { return { show: false, item: {}, isNew: true, selIcon: false, selIconFilter: '', was: {} } },
	props: [ 'type', 'fields', 'icon', 'default' ],
	methods: {
		edit(i) { this.was = {}; Object.assign(this.was, i); this.item = i;
			this.isNew = false; this.show = true },
		make() { this.item = {}
			if (this.default) { Object.assign(this.item, this.default) }
			this.isNew = true; this.show = true },
		save() { this.$emit('save', this.item); this.show = false },
		cancel() { this.show = false
			if (!this.isNew) { Object.assign(this.item, this.was) } },
	}, computed: {
		matFilterIcons() { var f = this.selIconFilter
			if (!f || !f.length) { return this.$root.maticons }
				return this.$root.materialIcons.filter(function(v) { return v.includes(f) }) },
	}, `)

	baseCol := `
	v-col[:cols="(field.cols || 12)"]`
	baseField := baseCol + `
		v-text-field
			[:type="field.type"]
			[:label="field.label || field.model"]
			[:prepend-icon="item[field.picon]"]
			[:suffix="field.suffix"]
			[:prefix="field.prefix"]`
	v.AddEditorFieldType("", baseField+`
			[v-model="item[field.model]"]
	`)
	v.AddEditorFieldType("number", baseField+`
			[type="number"]
			[v-model.number="item[field.model]"]
			[v-if="field.not ? (item[field.if] != true) : (item[field.if] != false)"]
	`)
	v.AddEditorFieldType("textarea", baseCol+`
		v-textarea[outlined][v-model="item[field.model]"] `)
	v.AddEditorFieldType("select", baseCol+`
		v-select
			[:items="$root[field.items]"]
			[v-model="item[field.model]"]
			[outlined]
			[:label="field.label || field.model"]`)
	v.AddEditorFieldType("slider", baseCol+`
		v-slider.pr-4
			[v-model="item[field.model]"]
			[:label="field.label || field.model"]
			template[v-slot:append]
				v-text-field.mt-0.pt-0
					[type="number"]
					[v-model="item[field.model]"]
					[:prepend-icon="item[field.picon]"]
					[:suffix="field.suffix"]`)
	v.AddEditorFieldType("switch", baseCol+`
		v-switch
			[v-model="item[field.model]"]
			[:label="field.label || field.model"]`)
	v.AddEditorFieldType("divider", baseCol+`
		v-divider
		span.overline[v-if="field.label"][v-text="field.label || field.model"]
		span.overline[v-if="field.flabel"][v-text="field.flabel(item, $root)"]
	`)
}
