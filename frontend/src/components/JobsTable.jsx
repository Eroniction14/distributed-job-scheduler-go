function JobsTable({ jobs, onRefresh }) {
  const toggleStatus = (id, currentStatus) => {
    const newStatus = currentStatus === 'active' ? 'paused' : 'active'
    fetch(`/api/jobs/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status: newStatus })
    }).then(() => onRefresh())
  }

  const statusColor = (status) => {
    const colors = {
      Active: { bg: '#dcfce7', color: '#16a34a' },
      Paused: { bg: '#fef9c3', color: '#ca8a04' },
      Running: { bg: '#dbeafe', color: '#2563eb' },
      Done: { bg: '#f0fdf4', color: '#15803d' },
      Failed: { bg: '#fee2e2', color: '#dc2626' },
      Pending: { bg: '#f1f5f9', color: '#64748b' },
    }
    return colors[status] || { bg: '#f1f5f9', color: '#64748b' }
  }

  return (
    <div style={{ padding: '0 32px 24px' }}>
      <div style={{
        background: 'white',
        borderRadius: 12,
        boxShadow: '0 1px 4px rgba(0,0,0,0.08)',
        overflow: 'hidden'
      }}>
        <div style={{ padding: '20px 24px', borderBottom: '1px solid #f1f5f9' }}>
          <h2 style={{ fontSize: 18 }}>All Jobs</h2>
        </div>
        <table>
          <thead>
            <tr style={{ background: '#f8fafc' }}>
              {['ID', 'Name', 'Command', 'Schedule', 'Status', 'Last Run', 'Action'].map(h => (
                <th key={h} style={{ padding: '12px 16px', fontSize: 12, color: '#64748b', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em' }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {jobs.length === 0 ? (
              <tr>
                <td colSpan="7" style={{ textAlign: 'center', padding: 32, color: '#94a3b8' }}>No jobs yet</td>
              </tr>
            ) : (
              jobs.map(job => (
                <tr key={job.id} style={{ borderTop: '1px solid #f1f5f9' }}>
                  <td style={td}>{job.id}</td>
                  <td style={{ ...td, fontWeight: 500 }}>{job.name}</td>
                  <td style={{ ...td, fontFamily: 'monospace', fontSize: 13, color: '#475569' }}>{job.command}</td>
                  <td style={{ ...td, fontFamily: 'monospace', fontSize: 13 }}>{job.schedule}</td>
                  <td style={td}>
                    <span style={{
                      ...statusColor(job.status),
                      padding: '3px 10px',
                      borderRadius: 20,
                      fontSize: 12,
                      fontWeight: 500
                    }}>
                      {job.status}
                    </span>
                  </td>
                  <td style={{ ...td, color: '#94a3b8', fontSize: 13 }}>{job.last_run ? new Date(job.last_run).toLocaleString() : '-'}</td>
                  <td style={td}>
                    <button onClick={() => toggleStatus(job.id, job.status)} style={{
                      background: job.status === 'active' ? '#fef9c3' : '#dcfce7',
                      color: job.status === 'active' ? '#ca8a04' : '#16a34a',
                      border: 'none',
                      borderRadius: 6,
                      padding: '5px 12px',
                      fontSize: 12,
                      fontWeight: 500
                    }}>
                      {job.status === 'active' ? 'Pause' : 'Resume'}
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}

const td = { padding: '14px 16px', fontSize: 14 }

export default JobsTable