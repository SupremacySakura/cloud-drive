import axios from "axios";

const request = axios.create({
    timeout: 10000,
})
request.interceptors.request.use(
    (config) => {
        return config
    },
    (error) => {
        return error
    }
)

request.interceptors.response.use(
    (config) => {
        return config
    },
    (error) => {
        return error
    }
)

export default request