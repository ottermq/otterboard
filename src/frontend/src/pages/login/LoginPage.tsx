import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { login } from '../../api/auth'
import { listWorkspaces } from '../../api/workspaces'

export default function LoginPage() {
    const navigate = useNavigate()

    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(false)

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        setError('')
        setLoading(true)
        try {
            await login({ email, password })
            const workspaces = await listWorkspaces()
            if (workspaces.length > 0) {
                navigate(`/workspaces/${workspaces[0].id}/issues`)
            }
        } catch {
            setError('Invalid email or password')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50">
            <form onSubmit={handleSubmit} className="flex flex-col gap-4 w-80 bg-white p-8 rounded-lg shadow">
                <h1 className="text-2xl font-semibold">Sign in</h1>
                {error && <p className="text-red-500 text-sm">{error}</p>}
                <input
                    type="email"
                    placeholder="Email"
                    value={email}
                    onChange={e => setEmail(e.target.value)}
                    className="border rounded px-3 py-2 text-sm"
                    required
                />
                <input
                    type="password"
                    placeholder="Password"
                    value={password}
                    onChange={e => setPassword(e.target.value)}
                    className="border rounded px-3 py-2 text-sm"
                    required
                />
                <button
                    type="submit"
                    disabled={loading}
                    className="bg-blue-600 text-white rounded px-3 py-2 text-sm disabled:opacity-50"
                >
                    {loading ? 'Signing in...' : 'Sign in'}
                </button>
            </form>
        </div>
    )
}