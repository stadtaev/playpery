import { useState, useEffect } from 'react'
import { getScenario, createScenario, updateScenario } from './adminApi'
import type { Stage, ScenarioRequest } from './adminTypes'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

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

function emptyStage(): Stage {
  return { stageNumber: 0, location: '', clue: '', question: '', correctAnswer: '', lat: 0, lng: 0 }
}

export function AdminScenarioEditorPage({ id }: { id?: string }) {
  const [name, setName] = useState('')
  const [city, setCity] = useState('')
  const [description, setDescription] = useState('')
  const [mode, setMode] = useState('classic')
  const [hasQuestions, setHasQuestions] = useState(false)
  const [stages, setStages] = useState<Stage[]>([emptyStage()])
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
        setStages(s.stages.length > 0 ? s.stages : [emptyStage()])
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id])

  function updateStage(index: number, field: keyof Stage, value: string | number) {
    setStages((prev) => prev.map((s, i) => (i === index ? { ...s, [field]: value } : s)))
  }

  function addStage() {
    setStages((prev) => [...prev, emptyStage()])
  }

  function removeStage(index: number) {
    setStages((prev) => prev.filter((_, i) => i !== index))
  }

  function moveStage(index: number, direction: -1 | 1) {
    const target = index + direction
    if (target < 0 || target >= stages.length) return
    setStages((prev) => {
      const next = [...prev]
      ;[next[index], next[target]] = [next[target], next[index]]
      return next
    })
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
      stages: stages.map((s, i) => ({ ...s, stageNumber: i + 1 })),
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
    return <p aria-busy="true">Loading scenario...</p>
  }

  return (
    <>
      <h2>{id ? 'Edit Scenario' : 'New Scenario'}</h2>
      {error && <p role="alert" style={{ color: 'var(--pico-color-red-500)' }}>{error}</p>}
      <form onSubmit={handleSubmit}>
        <div className="grid">
          <label>
            Name
            <input type="text" value={name} onChange={(e) => setName(e.target.value)} required />
          </label>
          <label>
            City
            <input type="text" value={city} onChange={(e) => setCity(e.target.value)} required />
          </label>
        </div>
        <label>
          Description
          <textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={2} />
        </label>

        <div className="grid">
          <label>
            Mode
            <select value={mode} onChange={(e) => setMode(e.target.value)}>
              {Object.entries(modeLabels).map(([value, label]) => (
                <option key={value} value={value}>{label}</option>
              ))}
            </select>
          </label>
          {mode === 'guided' && (
            <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', paddingTop: '1.75rem' }}>
              <input type="checkbox" checked={hasQuestions} onChange={(e) => setHasQuestions(e.target.checked)} role="switch" />
              Include questions at each stage
            </label>
          )}
        </div>

        <h3>Stages</h3>
        {stages.map((stage, i) => (
          <article key={i} style={{ padding: '1rem', marginBottom: '1rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.5rem' }}>
              <strong>Stage {i + 1}</strong>
              <div>
                <button type="button" className="outline secondary" onClick={() => moveStage(i, -1)} disabled={i === 0} style={{ width: 'auto', padding: '0.15rem 0.4rem', marginRight: '0.25rem' }}>
                  &uarr;
                </button>
                <button type="button" className="outline secondary" onClick={() => moveStage(i, 1)} disabled={i === stages.length - 1} style={{ width: 'auto', padding: '0.15rem 0.4rem', marginRight: '0.25rem' }}>
                  &darr;
                </button>
                <button type="button" className="outline secondary" onClick={() => removeStage(i)} disabled={stages.length <= 1} style={{ width: 'auto', padding: '0.15rem 0.4rem' }}>
                  &times;
                </button>
              </div>
            </div>
            <label>
              Location
              <input type="text" value={stage.location} onChange={(e) => updateStage(i, 'location', e.target.value)} required />
            </label>
            <label>
              Clue
              <textarea value={stage.clue} onChange={(e) => updateStage(i, 'clue', e.target.value)} rows={2} />
            </label>
            {(mode === 'qr_quiz' || mode === 'qr_hunt') && (
              <label>
                Unlock Code
                <input type="text" value={stage.unlockCode || ''} onChange={(e) => updateStage(i, 'unlockCode', e.target.value)} placeholder="Auto-generated if empty" />
              </label>
            )}
            {mode === 'math_puzzle' && (
              <label>
                Location Number
                <input type="number" value={stage.locationNumber || ''} onChange={(e) => updateStage(i, 'locationNumber', parseInt(e.target.value) || 0)} required />
              </label>
            )}
            {modeNeedsQuestion(mode, hasQuestions) && (
              <>
                <label>
                  Question
                  <input type="text" value={stage.question} onChange={(e) => updateStage(i, 'question', e.target.value)} required />
                </label>
                <label>
                  Correct Answer
                  <input type="text" value={stage.correctAnswer} onChange={(e) => updateStage(i, 'correctAnswer', e.target.value)} required />
                </label>
              </>
            )}
            <div className="grid">
              <label>
                Latitude
                <input type="number" step="any" value={stage.lat || ''} onChange={(e) => updateStage(i, 'lat', parseFloat(e.target.value) || 0)} />
              </label>
              <label>
                Longitude
                <input type="number" step="any" value={stage.lng || ''} onChange={(e) => updateStage(i, 'lng', parseFloat(e.target.value) || 0)} />
              </label>
            </div>
          </article>
        ))}

        <button type="button" className="outline" onClick={addStage} style={{ marginBottom: '1rem' }}>
          Add Stage
        </button>

        <div style={{ display: 'flex', gap: '1rem' }}>
          <button type="submit" disabled={saving} aria-busy={saving}>
            {saving ? 'Saving...' : id ? 'Update Scenario' : 'Create Scenario'}
          </button>
          <button type="button" className="secondary" onClick={() => navigate('/admin/scenarios')}>
            Cancel
          </button>
        </div>
      </form>
    </>
  )
}
