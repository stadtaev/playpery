import { useState, useEffect } from 'react'
import { listScenarios, deleteScenario } from './adminApi'
import type { ScenarioSummary } from './adminTypes'

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
    return <p aria-busy="true">Loading scenarios...</p>
  }

  if (error) {
    return <p role="alert">{error}</p>
  }

  return (
    <>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
        <h2 style={{ margin: 0 }}>Scenarios</h2>
        <button onClick={() => navigate('/admin/scenarios/new')} style={{ width: 'auto' }}>
          New Scenario
        </button>
      </div>

      {scenarios.length === 0 ? (
        <p>No scenarios yet.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>City</th>
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
                <td>{s.stageCount}</td>
                <td>{new Date(s.createdAt).toLocaleDateString()}</td>
                <td>
                  <button
                    className="outline secondary"
                    onClick={() => handleDelete(s.id, s.name)}
                    style={{ width: 'auto', padding: '0.25rem 0.5rem', fontSize: 'small' }}
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
