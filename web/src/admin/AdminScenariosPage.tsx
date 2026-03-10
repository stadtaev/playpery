import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { listScenarios, deleteScenario, exportScenario, importScenario } from './adminApi'
import type { ScenarioSummary } from './adminTypes'
import { LoadingPage } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function AdminScenariosPage() {
  const { t } = useTranslation('admin')
  const [scenarios, setScenarios] = useState<ScenarioSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    listScenarios()
      .then((s) => s.sort((a, b) => b.createdAt.localeCompare(a.createdAt)))
      .then(setScenarios)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  async function handleExport(id: string, name: string) {
    try {
      await exportScenario(id, name)
    } catch (e) {
      alert(e instanceof Error ? e.message : t('scenarios_export_failed'))
    }
  }

  async function handleImport(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    e.target.value = ''
    try {
      const created = await importScenario(file)
      setScenarios((prev) => [{ id: created.id, name: created.name, city: created.city, description: created.description, mode: created.mode, stageCount: created.stages.length, createdAt: created.createdAt }, ...prev])
    } catch (err) {
      alert(err instanceof Error ? err.message : t('scenarios_import_failed'))
    }
  }

  async function handleDelete(id: string, name: string) {
    if (!confirm(t('scenarios_delete_confirm', { name }))) return
    try {
      await deleteScenario(id)
      setScenarios((prev) => prev.filter((s) => s.id !== id))
    } catch (e) {
      alert(e instanceof Error ? e.message : t('scenarios_delete_failed'))
    }
  }

  if (loading) {
    return <LoadingPage message={t('scenarios_loading')} />
  }

  if (error) {
    return <ErrorMessage message={error} />
  }

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h2 className="m-0">{t('scenarios_title')}</h2>
        <div className="flex gap-2">
          <input
            ref={fileInputRef}
            type="file"
            accept=".md"
            className="hidden"
            onChange={handleImport}
          />
          <button onClick={() => fileInputRef.current?.click()} className="btn-secondary btn-sm">
            {t('scenarios_import')}
          </button>
          <button onClick={() => navigate('/admin/scenarios/new')} className="btn">
            {t('scenarios_new')}
          </button>
        </div>
      </div>

      {scenarios.length === 0 ? (
        <p className="text-secondary">{t('scenarios_empty')}</p>
      ) : (
        <table className="admin-table">
          <thead>
            <tr>
              <th>{t('scenarios_col_name')}</th>
              <th>{t('scenarios_col_city')}</th>
              <th>{t('scenarios_col_mode')}</th>
              <th>{t('scenarios_col_stages')}</th>
              <th>{t('scenarios_col_created')}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {scenarios.map((s) => (
              <tr key={s.id}>
                <td>
                  <a
                    href={`/admin/scenarios/${s.id}/edit`}
                    onClick={(e) => { e.preventDefault(); navigate(`/admin/scenarios/${s.id}/edit`) }}
                  >
                    {s.name}
                  </a>
                </td>
                <td>{s.city}</td>
                <td>{s.mode || 'classic'}</td>
                <td>{s.stageCount}</td>
                <td>{new Date(s.createdAt).toLocaleDateString()}</td>
                <td>
                  <div className="flex gap-2">
                    <button
                      className="btn-secondary btn-sm"
                      onClick={() => handleExport(s.id, s.name)}
                    >
                      {t('scenarios_export')}
                    </button>
                    <button
                      className="btn-danger btn-sm"
                      onClick={() => handleDelete(s.id, s.name)}
                    >
                      {t('scenarios_delete')}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </>
  )
}
