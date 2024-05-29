const env = import.meta.env.ENV;

let API_URL = ""

if (env.length === 0 || env === "local") {
    API_URL = "http://localhost:80"
} else {
    API_URL = "https://api.calculator.vingp.dev"
}

export {API_URL}