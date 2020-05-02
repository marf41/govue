GoVue = {
    install(Vue, options) {
    }
}

Vue.use(GoVue)
Vue.use(Vuetify)

Vue.component('v-footer-cmp', { template: '#tmp-footer', data() { return { text: '' } },
    methods: { set(s) { this.text = s } },
    computed: { value() { var y = new Date().getFullYear(); return this.text || ('&copy; ' + y) } } }) 
Vue.component('v-snack', { template: '#tmp-snack', data() { return { text: '', shw: false, color: '', timeout: 3000 } },
    methods: { set(s, c) { this.text = s; this.color = c || ''; this.shw = true },
                show() { this.shw = true } } })