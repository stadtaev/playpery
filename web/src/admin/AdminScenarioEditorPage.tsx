import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Plus, ChevronDown, ArrowUp, ArrowDown, Trash2, Save, X } from 'lucide-react'
import { getScenario, createScenario, updateScenario } from './adminApi'
import type { Stage, ScenarioRequest } from './adminTypes'
import { navigate } from '@/lib/navigate'
import { Button, MotionButton } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Alert } from '@/components/ui/alert'
import { Spinner } from '@/components/ui/spinner'

const modeLabels: Record<string, string> = {
  classic: 'Classic',
  qr_quiz: 'QR Quiz',
  qr_hunt: 'QR Hunt',
  math_puzzle: 'Math Puzzle',
  guided: 'Guided',
}

function modeNeedsQuestion(mode: string, hasQuestions: boolean): boolean {
  return mode === 'classic' || mode === 'qr_quiz' || (mode === 'guided' && hasQuestions)
}

type StageWithKey = Stage & { _key: string }
let stageKeyCounter = 0

function emptyStage(): StageWithKey {
  return { _key: `s${++stageKeyCounter}`, stageNumber: 0, location: '', clue: '', question: '', correctAnswer: '', lat: 0, lng: 0 }
}

function withKeys(stages: Stage[]): StageWithKey[] {
  return stages.map((s) => ({ ...s, _key: `s${++stageKeyCounter}` }))
}

function stagePreview(stage: Stage): string {
  return stage.clue ? stage.clue.slice(0, 60) + (stage.clue.length > 60 ? '...' : '') : 'No clue set'
}

export function AdminScenarioEditorPage({ id }: { id?: string }) {
  const [name, setName] = useState('')
  const [city, setCity] = useState('')
  const [description, setDescription] = useState('')
  const [mode, setMode] = useState('classic')
  const [hasQuestions, setHasQuestions] = useState(false)
  const [stages, setStages] = useState<StageWithKey[]>([emptyStage()])
  const [expandedStage, setExpandedStage] = useState<number | null>(0)
  const [loading, setLoading] = useState(!!id)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    getScenario(id)
      .then((s) => {
        setName(s.name)
        setCity(s.city)
        setDescription(s.description)
        setMode(s.mode || 'classic')
        setHasQuestions(s.hasQuestions || false)
        setStages(s.stages.length > 0 ? withKeys(s.stages) : [emptyStage()])
        setExpandedStage(null)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id])

  function updateStage(index: number, field: keyof Stage, value: string | number) {
    setStages((prev) => prev.map((s, i) => (i === index ? { ...s, [field]: value } : s)))
  }

  function addStage() {
    setStages((prev) => {
      setExpandedStage(prev.length)
      return [...prev, emptyStage()]
    })
  }

  function removeStage(index: number) {
    setStages((prev) => prev.filter((_, i) => i !== index))
    if (expandedStage === index) setExpandedStage(null)
    else if (expandedStage !== null && expandedStage > index) setExpandedStage(expandedStage - 1)
  }

  function moveStage(index: number, direction: -1 | 1) {
    const target = index + direction
    if (target < 0 || target >= stages.length) return
    setStages((prev) => {
      const next = [...prev]
      ;[next[index], next[target]] = [next[target], next[index]]
      return next
    })
    if (expandedStage === index) setExpandedStage(target)
    else if (expandedStage === target) setExpandedStage(index)
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError('')

    const data: ScenarioRequest = {
      name,
      city,
      description,
      mode,
      hasQuestions: mode === 'guided' ? hasQuestions : undefined,
      stages: stages.map(({ _key, ...s }, i) => ({ ...s, stageNumber: i + 1 })),
    }

    try {
      if (id) {
        await updateScenario(id, data)
      } else {
        await createScenario(data)
      }
      navigate('/admin/scenarios')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Save failed')
      setSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Spinner size={32} />
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold text-text-primary">
          {id ? 'Edit Scenario' : 'New Scenario'}
        </h2>
      </div>

      {error && (
        <div className="mb-4">
          <Alert variant="error">{error}</Alert>
        </div>
      )}

      <form onSubmit={handleSubmit}>
        {/* Scenario metadata */}
        <div className="rounded-lg border border-border bg-card p-5 mb-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div className="space-y-1.5">
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)} required />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="city">City</Label>
              <Input id="city" value={city} onChange={(e) => setCity(e.target.value)} required />
            </div>
          </div>

          <div className="space-y-1.5 mb-4">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1.5">
              <Label htmlFor="mode">Mode</Label>
              <Select id="mode" value={mode} onChange={(e) => setMode(e.target.value)}>
                {Object.entries(modeLabels).map(([value, label]) => (
                  <option key={value} value={value}>{label}</option>
                ))}
              </Select>
            </div>
            {mode === 'guided' && (
              <div className="flex items-center gap-3 pt-6">
                <Switch checked={hasQuestions} onCheckedChange={setHasQuestions} />
                <Label className="cursor-pointer" onClick={() => setHasQuestions(!hasQuestions)}>
                  Include questions at each stage
                </Label>
              </div>
            )}
          </div>
        </div>

        {/* Stages accordion */}
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-base font-medium text-text-primary">Stages</h3>
          <Button type="button" variant="outline" size="sm" onClick={addStage}>
            <Plus size={14} />
            Add Stage
          </Button>
        </div>

        <div className="space-y-2 mb-6">
          {stages.map((stage, i) => {
            const isExpanded = expandedStage === i
            return (
              <div
                key={stage._key}
                className="rounded-lg border border-border bg-card overflow-hidden"
              >
                {/* Collapsed header row */}
                <button
                  type="button"
                  className="w-full flex items-center gap-3 px-4 py-3 bg-transparent border-none cursor-pointer text-left hover:bg-popover transition-colors"
                  onClick={() => setExpandedStage(isExpanded ? null : i)}
                >
                  <Badge variant="default" className="shrink-0 font-mono">{i + 1}</Badge>
                  <span className="font-medium text-text-primary truncate min-w-0">
                    {stage.location || 'Untitled stage'}
                  </span>
                  <span className="text-text-muted text-xs truncate min-w-0 hidden sm:block">
                    {stagePreview(stage)}
                  </span>
                  <div className="ml-auto flex items-center gap-1 shrink-0">
                    <span
                      className="inline-flex items-center justify-center h-7 w-7 rounded text-text-muted hover:text-text-primary hover:bg-card transition-colors"
                      role="button"
                      tabIndex={0}
                      onClick={(e) => { e.stopPropagation(); moveStage(i, -1) }}
                      onKeyDown={(e) => { if (e.key === 'Enter') { e.stopPropagation(); moveStage(i, -1) } }}
                      aria-label="Move up"
                      aria-disabled={i === 0}
                      style={i === 0 ? { opacity: 0.3, pointerEvents: 'none' } : undefined}
                    >
                      <ArrowUp size={14} />
                    </span>
                    <span
                      className="inline-flex items-center justify-center h-7 w-7 rounded text-text-muted hover:text-text-primary hover:bg-card transition-colors"
                      role="button"
                      tabIndex={0}
                      onClick={(e) => { e.stopPropagation(); moveStage(i, 1) }}
                      onKeyDown={(e) => { if (e.key === 'Enter') { e.stopPropagation(); moveStage(i, 1) } }}
                      aria-label="Move down"
                      aria-disabled={i === stages.length - 1}
                      style={i === stages.length - 1 ? { opacity: 0.3, pointerEvents: 'none' } : undefined}
                    >
                      <ArrowDown size={14} />
                    </span>
                    <span
                      className="inline-flex items-center justify-center h-7 w-7 rounded text-text-muted hover:text-error hover:bg-card transition-colors"
                      role="button"
                      tabIndex={0}
                      onClick={(e) => { e.stopPropagation(); removeStage(i) }}
                      onKeyDown={(e) => { if (e.key === 'Enter') { e.stopPropagation(); removeStage(i) } }}
                      aria-label="Remove stage"
                      aria-disabled={stages.length <= 1}
                      style={stages.length <= 1 ? { opacity: 0.3, pointerEvents: 'none' } : undefined}
                    >
                      <Trash2 size={14} />
                    </span>
                    <motion.span
                      className="inline-flex items-center justify-center h-7 w-7 text-text-muted"
                      animate={{ rotate: isExpanded ? 180 : 0 }}
                      transition={{ duration: 0.2 }}
                    >
                      <ChevronDown size={16} />
                    </motion.span>
                  </div>
                </button>

                {/* Expanded form */}
                <AnimatePresence initial={false}>
                  {isExpanded && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: 'auto', opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      transition={{ duration: 0.25, ease: 'easeInOut' }}
                      className="overflow-hidden"
                    >
                      <div className="px-4 pb-4 pt-1 border-t border-border">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-3">
                          <div className="space-y-1.5">
                            <Label>Location</Label>
                            <Input
                              value={stage.location}
                              onChange={(e) => updateStage(i, 'location', e.target.value)}
                              required
                            />
                          </div>
                          <div className="space-y-1.5">
                            <Label>Clue</Label>
                            <Textarea
                              value={stage.clue}
                              onChange={(e) => updateStage(i, 'clue', e.target.value)}
                              rows={2}
                            />
                          </div>
                        </div>

                        {/* Mode-specific fields */}
                        {(mode === 'qr_quiz' || mode === 'qr_hunt') && (
                          <div className="mt-4 space-y-1.5">
                            <Label>Unlock Code</Label>
                            <Input
                              value={stage.unlockCode || ''}
                              onChange={(e) => updateStage(i, 'unlockCode', e.target.value)}
                              placeholder="Auto-generated if empty"
                            />
                          </div>
                        )}

                        {mode === 'math_puzzle' && (
                          <div className="mt-4 space-y-1.5 max-w-xs">
                            <Label>Location Number</Label>
                            <Input
                              type="number"
                              value={stage.locationNumber || ''}
                              onChange={(e) => updateStage(i, 'locationNumber', parseInt(e.target.value) || 0)}
                              required
                            />
                          </div>
                        )}

                        {modeNeedsQuestion(mode, hasQuestions) && (
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
                            <div className="space-y-1.5">
                              <Label>Question</Label>
                              <Input
                                value={stage.question}
                                onChange={(e) => updateStage(i, 'question', e.target.value)}
                                required
                              />
                            </div>
                            <div className="space-y-1.5">
                              <Label>Correct Answer</Label>
                              <Input
                                value={stage.correctAnswer}
                                onChange={(e) => updateStage(i, 'correctAnswer', e.target.value)}
                                required
                              />
                            </div>
                          </div>
                        )}

                        {/* Lat/Lng row */}
                        <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mt-4">
                          <div className="space-y-1.5">
                            <Label>Latitude</Label>
                            <Input
                              type="number"
                              step="any"
                              value={stage.lat || ''}
                              onChange={(e) => updateStage(i, 'lat', parseFloat(e.target.value) || 0)}
                            />
                          </div>
                          <div className="space-y-1.5">
                            <Label>Longitude</Label>
                            <Input
                              type="number"
                              step="any"
                              value={stage.lng || ''}
                              onChange={(e) => updateStage(i, 'lng', parseFloat(e.target.value) || 0)}
                            />
                          </div>
                        </div>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            )
          })}
        </div>

        {/* Actions */}
        <div className="flex gap-3">
          <MotionButton type="submit" disabled={saving}>
            {saving ? (
              <>
                <Spinner size={16} />
                Saving...
              </>
            ) : (
              <>
                <Save size={16} />
                {id ? 'Update Scenario' : 'Create Scenario'}
              </>
            )}
          </MotionButton>
          <Button type="button" variant="outline" onClick={() => navigate('/admin/scenarios')}>
            <X size={16} />
            Cancel
          </Button>
        </div>
      </form>
    </div>
  )
}
