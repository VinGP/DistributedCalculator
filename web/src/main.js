import './assets/main.css'


import {createApp} from 'vue'
import {createPinia} from 'pinia'

import App from './App.vue'
import router from './router'
import VueAxios from "vue-axios";
import axios from "axios";




const app = createApp(App)
app.use(VueAxios, axios)

// app.use(Skeleton)



app.use(createPinia())
app.use(router)

app.mount('#app')
