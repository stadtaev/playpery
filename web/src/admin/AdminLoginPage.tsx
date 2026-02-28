import { useState } from 'react'
import { motion } from 'framer-motion'
import { LogIn } from 'lucide-react'
import { login } from './adminApi'
import { Card, CardHeader, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { MotionButton } from '@/components/ui/button'
import { Alert } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'

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
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: 'easeOut' }}
        className="w-full max-w-sm"
      >
        <Card>
          <CardHeader>
            <h1 className="text-xl font-semibold text-text-primary">Admin Login</h1>
            <p className="text-sm text-text-secondary">Sign in to manage your quests</p>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="admin@playperu.com"
                  autoFocus
                  required
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="password">Password</Label>
                <Input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </div>
              {error && <Alert variant="error">{error}</Alert>}
              <MotionButton type="submit" disabled={loading} className="w-full">
                {loading ? (
                  <>
                    <Spinner size={16} className="text-accent-foreground" />
                    Logging in...
                  </>
                ) : (
                  <>
                    <LogIn size={16} />
                    Log in
                  </>
                )}
              </MotionButton>
            </form>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  )
}
