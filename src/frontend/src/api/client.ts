import axios from 'axios'

const client = axios.create({
    baseURL: import.meta.env.VITE_API_URL ?? 'http://localhost:8088/api/v1',
    withCredentials: true,
})

export default client