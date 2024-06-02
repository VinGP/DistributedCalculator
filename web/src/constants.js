const env = import.meta.env

let API_URL = ""

console.log(env.VITE_API_ENDPOINT)

if (!env.VITE_API_ENDPOINT) {
    API_URL = "http://localhost:8080"
} else {
    API_URL = env.VITE_API_ENDPOINT
}

export {API_URL}