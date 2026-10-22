import { createApp } from "vue";
import { createPinia } from "pinia";
import "@arco-design/web-vue/dist/arco.css";
import "./styles/global.css";
import App from "./App.vue";
import router from "./router/index.js";

createApp(App).use(createPinia()).use(router).mount("#app");
