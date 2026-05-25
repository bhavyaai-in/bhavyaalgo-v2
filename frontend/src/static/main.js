import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from '../App.vue'
import router from '../router'
import { setRouter } from '../utils/api.js'
import './style.css'
import './broker-common.css'
import './common.css'

setRouter(router)
createApp(App).use(createPinia()).use(router).mount('#app')
