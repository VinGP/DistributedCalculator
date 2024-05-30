import {fileURLToPath, URL} from 'node:url'

import {defineConfig, loadEnv} from 'vite'
import vue from '@vitejs/plugin-vue'
// import process from "eslint-plugin-vue/lib/configs/base.js";
// import VueDevTools from 'vite-plugin-vue-devtools'

// https://vitejs.dev/config/
export default defineConfig(() => {
    const env = loadEnv('', process.cwd());
    console.log(env)
    return {
        server: {
            port: 80
        },
        preview: {
            port: 80
        },
        plugins: [
            vue(),
            // VueDevTools(),
        ],
        define: {
            ENV: env.ENV,
        },
        optimizeDeps: {
            exclude: ['js-big-decimal']
        },
        resolve: {
            alias: {
                '@': fileURLToPath(new URL('./src', import.meta.url))
            }
        }
    }
})
