import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { getScenario, createScenario, updateScenario } from './adminApi'
import type { Stage, ScenarioRequest } from './adminTypes'
import { LoadingPage, Spinner } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

function modeNeedsQuestion(mode: string): boolean {
  return mode === 'classic' || mode === 'qr_quiz' || mode === 'supervised'
}

function emptyStage(): Stage {
  return { stageNumber: 0, location: '', clue: '', question: '', correctAnswer: '', funFacts: [], lat: 0, lng: 0 }
}

export function AdminScenarioEditorPage({ id }: { id?: string }) {
  const { t } = useTranslation('admin')
  const [name, setName] = useState('')
  const [city, setCity] = useState('')
  const [description, setDescription] = useState('')
  const [mode, setMode] = useState('supervised')
  const [stages, setStages] = useState<Stage[]>([emptyStage()])
  const [loading, setLoading] = useState(!!id)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const modeLabels: Record<string, string> = {
    supervised: t('mode_supervised'),
    classic: t('mode_classic'),
    qr_quiz: t('mode_qr_quiz'),
    qr_hunt: t('mode_qr_hunt'),
    math_puzzle: t('mode_math_puzzle'),
  }

  useEffect(() => {
    if (!id) return
    getScenario(id)
      .then((s) => {
        setName(s.name)
        setCity(s.city)
        setDescription(s.description)
        setMode(s.mode || 'supervised')
        setStages(s.stages.length > 0 ? s.stages : [emptyStage()])
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id])

  function updateStage(index: number, field: keyof Stage, value: string | number | string[]) {
    setStages((prev) => prev.map((s, i) => (i === index ? { ...s, [field]: value } : s)))
  }

  function updateFunFact(stageIndex: number, factIndex: number, value: string) {
    setStages((prev) => prev.map((s, i) => {
      if (i !== stageIndex) return s
      const facts = [...(s.funFacts || [])]
      facts[factIndex] = value
      return { ...s, funFacts: facts }
    }))
  }

  function addFunFact(stageIndex: number) {
    setStages((prev) => prev.map((s, i) => {
      if (i !== stageIndex) return s
      return { ...s, funFacts: [...(s.funFacts || []), ''] }
    }))
  }

  function removeFunFact(stageIndex: number, factIndex: number) {
    setStages((prev) => prev.map((s, i) => {
      if (i !== stageIndex) return s
      return { ...s, funFacts: (s.funFacts || []).filter((_, fi) => fi !== factIndex) }
    }))
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
      setError(e instanceof Error ? e.message : t('scenario_save_failed'))
      setSaving(false)
    }
  }

  if (loading) {
    return <LoadingPage message={t('scenarios_loading')} />
  }

  return (
    <>
      <h2>{id ? t('scenario_edit_title') : t('scenario_new_title')}</h2>
      {error && <ErrorMessage message={error} />}
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <label className="input-label" htmlFor="sc-name">{t('scenario_name')}</label>
            <input id="sc-name" className="input" type="text" value={name} onChange={(e) => setName(e.target.value)} required />
          </div>
          <div>
            <label className="input-label" htmlFor="sc-city">{t('scenario_city')}</label>
            <input id="sc-city" className="input" type="text" value={city} onChange={(e) => setCity(e.target.value)} required />
          </div>
        </div>
        <div>
          <label className="input-label" htmlFor="sc-desc">{t('scenario_description')}</label>
          <textarea id="sc-desc" className="input" value={description} onChange={(e) => setDescription(e.target.value)} rows={2} />
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 items-end">
          <div>
            <label className="input-label" htmlFor="sc-mode">{t('scenario_mode')}</label>
            <select id="sc-mode" className="input" value={mode} onChange={(e) => setMode(e.target.value)}>
              {Object.entries(modeLabels).map(([value, label]) => (
                <option key={value} value={value}>{label}</option>
              ))}
            </select>
          </div>
        </div>

        <h3 className="mt-8">{t('scenario_stages')}</h3>
        {stages.map((stage, i) => (
          <div key={i} className="card">
            <div className="flex justify-between items-center mb-4">
              <strong>{t('scenario_stage_n', { n: i + 1 })}</strong>
              <div className="flex gap-1">
                <button type="button" className="btn-ghost btn-sm" onClick={() => moveStage(i, -1)} disabled={i === 0}>
                  &uarr;
                </button>
                <button type="button" className="btn-ghost btn-sm" onClick={() => moveStage(i, 1)} disabled={i === stages.length - 1}>
                  &darr;
                </button>
                <button type="button" className="btn-danger btn-sm" onClick={() => removeStage(i)} disabled={stages.length <= 1}>
                  &times;
                </button>
              </div>
            </div>
            <div className="space-y-3">
              <div>
                <label className="input-label">{t('scenario_location')}</label>
                <input className="input" type="text" value={stage.location} onChange={(e) => updateStage(i, 'location', e.target.value)} required />
              </div>
              <div>
                <label className="input-label">{t('scenario_clue')}</label>
                <textarea className="input" value={stage.clue} onChange={(e) => updateStage(i, 'clue', e.target.value)} rows={2} />
              </div>
              {(mode === 'qr_quiz' || mode === 'qr_hunt') && (
                <div>
                  <label className="input-label">{t('scenario_unlock_code')}</label>
                  <input className="input" type="text" value={stage.unlockCode || ''} onChange={(e) => updateStage(i, 'unlockCode', e.target.value)} placeholder={t('scenario_unlock_code_placeholder')} />
                </div>
              )}
              {mode === 'math_puzzle' && (
                <div>
                  <label className="input-label">{t('scenario_location_number')}</label>
                  <input className="input" type="number" value={stage.locationNumber || ''} onChange={(e) => updateStage(i, 'locationNumber', parseInt(e.target.value) || 0)} required />
                </div>
              )}
              {modeNeedsQuestion(mode) && (
                <>
                  <div>
                    <label className="input-label">{t('scenario_question')}</label>
                    <input className="input" type="text" value={stage.question} onChange={(e) => updateStage(i, 'question', e.target.value)} required />
                  </div>
                  <div>
                    <label className="input-label">{t('scenario_correct_answer')}</label>
                    <input className="input" type="text" value={stage.correctAnswer} onChange={(e) => updateStage(i, 'correctAnswer', e.target.value)} required />
                  </div>
                  <div>
                    <label className="input-label">{t('scenario_fun_facts')}</label>
                    <div className="space-y-2">
                      {(stage.funFacts || []).map((fact, fi) => (
                        <div key={fi} className="flex gap-2">
                          <textarea
                            className="input flex-1"
                            value={fact}
                            onChange={(e) => updateFunFact(i, fi, e.target.value)}
                            onBlur={(e) => { if (!e.target.value.trim()) removeFunFact(i, fi) }}
                            rows={2}
                            placeholder={t('scenario_fun_fact_placeholder', { n: fi + 1 })}
                          />
                          <button type="button" className="btn-danger btn-sm self-start" onClick={() => removeFunFact(i, fi)}>
                            &times;
                          </button>
                        </div>
                      ))}
                      <button type="button" className="btn-ghost btn-sm" onClick={() => addFunFact(i)}>
                        {t('scenario_add_fun_fact')}
                      </button>
                    </div>
                  </div>
                </>
              )}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="input-label">{t('scenario_latitude')}</label>
                  <input className="input" type="number" step="any" value={stage.lat || ''} onChange={(e) => updateStage(i, 'lat', parseFloat(e.target.value) || 0)} />
                </div>
                <div>
                  <label className="input-label">{t('scenario_longitude')}</label>
                  <input className="input" type="number" step="any" value={stage.lng || ''} onChange={(e) => updateStage(i, 'lng', parseFloat(e.target.value) || 0)} />
                </div>
              </div>
            </div>
          </div>
        ))}

        <button type="button" className="btn-secondary w-full" onClick={addStage}>
          {t('scenario_add_stage')}
        </button>

        <div className="flex gap-4 mt-6">
          <button type="submit" disabled={saving} className="btn">
            {saving ? <Spinner /> : id ? t('scenario_update') : t('scenario_create')}
          </button>
          <button type="button" className="btn-secondary" onClick={() => navigate('/admin/scenarios')}>
            {t('scenario_cancel')}
          </button>
        </div>
      </form>
    </>
  )
}
