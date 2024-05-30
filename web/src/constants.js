const env = import.meta.env
// console.log(env)
// console.log(process.env.ENV)
let API_URL = ""

if (env.MODE === "development") {
    API_URL = "http://localhost:8080"
} else {
    API_URL = "https://api.calculator.vingp.dev"
}

export {API_URL}