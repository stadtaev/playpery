import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { ArrowLeft, Pencil, Timer, Users, Trophy, MapPin, Clock } from 'lucide-react'
import { getGameStatus } from './adminApi'
import type { GameStatus, TeamStatus } from './adminTypes'
import { navigate } from '@/lib/navigate'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardHeader, CardContent } from '@/components/ui/card'
import { Alert } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'

const statusBadgeVariant: Record<string, 'default' | 'success' | 'warning' | 'error'> = {
  draft: 'default',
  active: 'success',
  paused: 'warning',
  ended: 'error',
}

function ProgressBar({ completed, total }: { completed: number; total: number }) {
  const pct = total > 0 ? (completed / total) * 100 : 0
  return (
    <div className="w-full h-2 bg-base rounded-full overflow-hidden">
      <motion.div
        className="h-full bg-accent rounded-full"
        initial={{ width: 0 }}
        animate={{ width: `${pct}%` }}
        transition={{ duration: 0.6, ease: 'easeOut' }}
      />
    </div>
  )
}

function TeamCard({ team, totalStages, index }: { team: TeamStatus; totalStages: number; index: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.06, duration: 0.3 }}
    >
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="text-base font-semibold text-text-primary">{team.name}</span>
              {team.guideName && (
                <span className="text-xs text-text-muted">
                  Guide: {team.guideName}
                </span>
              )}
            </div>
            <div className="flex items-center gap-1.5">
              <Trophy size={14} className="text-accent" />
              <span className="text-sm font-medium text-accent">{team.completedStages} pts</span>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            <div>
              <div className="flex items-center justify-between text-xs text-text-secondary mb-1.5">
                <span>{team.completedStages}/{totalStages} stages</span>
                <span>{totalStages > 0 ? Math.round((team.completedStages / totalStages) * 100) : 0}%</span>
              </div>
              <ProgressBar completed={team.completedStages} total={totalStages} />
            </div>

            {team.players.length > 0 ? (
              <div className="space-y-1">
                <span className="text-xs font-medium text-text-muted uppercase tracking-wider">
                  Players ({team.players.length})
                </span>
                <div className="flex flex-wrap gap-1.5">
                  {team.players.map((p, i) => (
                    <span
                      key={i}
                      className="inline-flex items-center gap-1 rounded-md bg-base px-2 py-1 text-xs text-text-secondary"
                    >
                      {p.name}
                      {p.role === 'supervisor' && (
                        <Badge variant="warning" className="text-[10px] px-1 py-0">sup</Badge>
                      )}
                    </span>
                  ))}
                </div>
              </div>
            ) : (
              <p className="text-xs text-text-muted">No players yet</p>
            )}
          </div>
        </CardContent>
      </Card>
    </motion.div>
  )
}

export function AdminGameStatusPage({ client, id }: { client: string; id: string }) {
  const [game, setGame] = useState<GameStatus | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    let active = true

    function load() {
      getGameStatus(client, id)
        .then((g) => { if (active) { setGame(g); setError('') } })
        .catch((e) => { if (active) setError(e.message) })
    }

    load()
    const interval = setInterval(load, 5000)
    return () => { active = false; clearInterval(interval) }
  }, [client, id])

  if (error) {
    return (
      <div className="py-8">
        <Alert variant="error">{error}</Alert>
      </div>
    )
  }

  if (!game) {
    return (
      <div className="flex items-center justify-center py-20">
        <Spinner size={32} />
      </div>
    )
  }

  const totalPlayers = game.teams.reduce((sum, t) => sum + t.players.length, 0)
  const sortedTeams = [...game.teams].sort((a, b) => b.completedStages - a.completedStages)

  return (
    <div>
      <div className="flex items-center gap-3 mb-6">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => navigate(`/admin/clients/${client}/games`)}
          className="text-text-muted hover:text-text-primary"
        >
          <ArrowLeft size={18} />
        </Button>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-3 flex-wrap">
            <h2 className="text-xl font-semibold text-text-primary truncate">{game.scenarioName}</h2>
            <Badge variant={statusBadgeVariant[game.status] ?? 'default'}>{game.status}</Badge>
            {game.status === 'active' && (
              <div className="flex items-center gap-1">
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-success opacity-75" />
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-success" />
                </span>
                <span className="text-xs text-text-muted">Live</span>
              </div>
            )}
          </div>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => navigate(`/admin/clients/${client}/games/${id}/edit`)}
        >
          <Pencil size={14} />
          Edit
        </Button>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-6">
        <div className="rounded-lg border border-border bg-card p-3">
          <div className="flex items-center gap-2 text-text-muted mb-1">
            <Users size={14} />
            <span className="text-xs">Players</span>
          </div>
          <span className="text-lg font-semibold text-text-primary">{totalPlayers}</span>
        </div>
        <div className="rounded-lg border border-border bg-card p-3">
          <div className="flex items-center gap-2 text-text-muted mb-1">
            <Users size={14} />
            <span className="text-xs">Teams</span>
          </div>
          <span className="text-lg font-semibold text-text-primary">{game.teams.length}</span>
        </div>
        <div className="rounded-lg border border-border bg-card p-3">
          <div className="flex items-center gap-2 text-text-muted mb-1">
            <MapPin size={14} />
            <span className="text-xs">Stages</span>
          </div>
          <span className="text-lg font-semibold text-text-primary">{game.totalStages}</span>
        </div>
        <div className="rounded-lg border border-border bg-card p-3">
          <div className="flex items-center gap-2 text-text-muted mb-1">
            <Timer size={14} />
            <span className="text-xs">Timer</span>
          </div>
          <span className="text-lg font-semibold text-text-primary">
            {game.timerEnabled ? `${game.timerMinutes}m` : 'Off'}
          </span>
        </div>
      </div>

      {game.startedAt && (
        <div className="flex items-center gap-1.5 text-xs text-text-muted mb-6">
          <Clock size={12} />
          Started {new Date(game.startedAt).toLocaleString()}
        </div>
      )}

      {game.teams.length === 0 ? (
        <p className="text-text-secondary">No teams yet.</p>
      ) : (
        <>
          <h3 className="text-sm font-medium text-text-muted uppercase tracking-wider mb-4">
            Scoreboard
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {sortedTeams.map((team, i) => (
              <TeamCard key={team.id} team={team} totalStages={game.totalStages} index={i} />
            ))}
          </div>
        </>
      )}
    </div>
  )
}
