import { useState } from 'react'
import { login } from './adminApi'
import { PageContainer } from '../components/PageContainer'
import { Spinner } from '../components/Spinner'

export function AdminLoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await login(email, password)
      window.history.replaceState(null, '', '/admin/clients')
      window.dispatchEvent(new PopStateEvent('popstate'))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Login failed')
      setLoading(false)
    }
  }

  return (
    <PageContainer>
      <h1>Admin Login</h1>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="input-label" htmlFor="email">Email</label>
          <input
            id="email"
            className="input"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="admin@playperu.com"
            autoFocus
            required
          />
        </div>
        <div>
          <label className="input-label" htmlFor="password">Password</label>
          <input
            id="password"
            className="input"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        {error && <p className="text-feedback-error">{error}</p>}
        <button type="submit" disabled={loading} className="btn w-full">
          {loading ? <Spinner /> : 'Log in'}
        </button>
      </form>
    </PageContainer>
  )
}
