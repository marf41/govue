if (!vm) {
vm = new Vue({
    el: '#app', vuetify: new Vuetify(),
    // delimiters: ['${', '}'],
    data() { return {
    } },
    filters: {
    },
    methods: {
    },
    watch: {
    },
    computed: {
    },
    created() {
        var self = this
        axios.defaults.headers.patch['Allow-Patch'] = 'application/json'
        this.$vuetify.theme.dark = true
        if (window.maticons) { this.maticons = Object.keys(window.maticons) }
    },
    mounted() { if (this.refresh) { this.refresh() } },
})
}
