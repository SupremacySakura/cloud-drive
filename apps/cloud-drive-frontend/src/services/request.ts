import axios from "axios";
import { useUserStore } from "../stores/user";

const request = axios.create({
    timeout: 10000,
})
request.interceptors.request.use(
    (config) => {
        const token = useUserStore().token
        if(token){
            config.headers.Authorization = `Bearer ${token}`
        }
        return config
    },
    (error) => {
        return Promise.reject(error)
    }
)

request.interceptors.response.use(
    (config) => {
        if(config?.data?.data?.token){
            useUserStore().setToken(config.data.data.token)
        }
        return config
    },
    (error) => {
        return Promise.reject(error)
    }
)

export default request