import client from './client'

export interface LogintInput {
    email: string
    password: string
}

export function login(input: LogintInput) {
    return client.post('/auth/login', input).then(response => response.data)
}