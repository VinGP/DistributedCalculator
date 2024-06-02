import {fileURLToPath, URL} from 'node:url'

import {defineConfig, loadEnv} from 'vite'
import vue from '@vitejs/plugin-vue'


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
        ],
        define: {
            VITE_API_ENDPOINT: env.ENV,
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
