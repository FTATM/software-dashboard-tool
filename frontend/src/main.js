import { createApp } from 'vue'
import '@vuepic/vue-datepicker/dist/main.css';
import './style.css'
import App from './App.vue'
import router from './router';
import 'vue3-toastify/dist/index.css';
import { createPinia } from 'pinia';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate';
import { i18n } from './i18n'

const app = createApp(App);
const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);
app.use(i18n)
app.use(pinia);
app.use(router);
app.mount('#app');
