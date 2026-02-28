import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { Users, UserCog } from 'lucide-react'
import { lookupTeam, joinTeam } from './api'
import type { TeamLookup } from './types'
import { Card, CardHeader, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { MotionButton } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Alert } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'

export function JoinPage({ client, joinToken }: { client: string; joinToken: string }) {
  const [team, setTeam] = useState<TeamLookup | null>(null)
  const [error, setError] = useState('')
  const [name, setName] = useState('')
  const [joining, setJoining] = useState(false)

  useEffect(() => {
    lookupTeam(client, joinToken)
      .then(setTeam)
      .catch((e) => setError(e.message))
  }, [client, joinToken])

  async function handleJoin(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    setJoining(true)
    setError('')
    try {
      const resp = await joinTeam(client, joinToken, name.trim())
      localStorage.setItem('session_token', resp.token)
      localStorage.setItem('team_name', resp.teamName)
      localStorage.setItem('player_role', resp.role)
      localStorage.setItem('client', client)
      window.history.replaceState(null, '', '/game')
      window.dispatchEvent(new PopStateEvent('popstate'))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to join')
      setJoining(false)
    }
  }

  if (error && !team) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background px-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, ease: 'easeOut' }}
          className="w-full max-w-md"
        >
          <Card>
            <CardHeader>
              <h1 className="text-xl font-semibold text-text-primary">CityQuest</h1>
            </CardHeader>
            <CardContent>
              <Alert variant="error">{error}</Alert>
            </CardContent>
          </Card>
        </motion.div>
      </div>
    )
  }

  if (!team) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background px-4">
        <Spinner />
      </div>
    )
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: 'easeOut' }}
        className="w-full max-w-md"
      >
        <Card>
          <CardHeader>
            <h1 className="text-xl font-semibold text-text-primary">CityQuest</h1>
            <div className="flex items-center gap-2">
              <h2 className="text-lg text-text-secondary">Join {team.name}</h2>
              {team.role === 'supervisor' && (
                <Badge variant="warning">
                  <UserCog size={12} className="mr-1" />
                  Supervisor
                </Badge>
              )}
            </div>
            <p className="text-sm text-text-muted">{team.gameName}</p>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleJoin} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="player-name">Your name</Label>
                <Input
                  id="player-name"
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="Enter your name"
                  autoFocus
                  required
                />
              </div>
              {error && <Alert variant="error">{error}</Alert>}
              <MotionButton type="submit" disabled={joining} className="w-full">
                {joining ? (
                  <>
                    <Spinner size={16} className="text-accent-foreground" />
                    Joining...
                  </>
                ) : (
                  <>
                    <Users size={16} />
                    Join Game
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
