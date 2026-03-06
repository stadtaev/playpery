import { useState, useEffect } from 'react'
import { listScenarios, deleteScenario } from './adminApi'
import type { ScenarioSummary } from './adminTypes'
import { LoadingPage } from '../components/Spinner'
import { ErrorMessage } from '../components/ErrorMessage'

function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function AdminScenariosPage() {
  const [scenarios, setScenarios] = useState<ScenarioSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    listScenarios()
      .then((s) => s.sort((a, b) => b.createdAt.localeCompare(a.createdAt)))
      .then(setScenarios)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  async function handleDelete(id: string, name: string) {
    if (!confirm(`Delete scenario "${name}"?`)) return
    try {
      await deleteScenario(id)
      setScenarios((prev) => prev.filter((s) => s.id !== id))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'Delete failed')
    }
  }

  if (loading) {
    return <LoadingPage message="Loading scenarios..." />
  }

  if (error) {
    return <ErrorMessage message={error} />
  }

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h2 className="m-0">Scenarios</h2>
        <button onClick={() => navigate('/admin/scenarios/new')} className="btn">
          New Scenario
        </button>
      </div>

      {scenarios.length === 0 ? (
        <p className="text-secondary">No scenarios yet.</p>
      ) : (
        <table className="admin-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>City</th>
              <th>Mode</th>
              <th>Stages</th>
              <th>Created</th>
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
                  <button
                    className="btn-danger btn-sm"
                    onClick={() => handleDelete(s.id, s.name)}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </>
  )
}
