function LogsTable({ logs }) {
  const statusColor = (status) => ({
    success: { bg: '#dcfce7', color: '#16a34a' },
    failed: { bg: '#fee2e2', color: '#dc2626' },
  }[status] || { bg: '#f1f5f9', color: '#64748b' })

  return (
    <div style={{ padding: '0 32px 24px' }}>
      <div style={{
        background: 'white',
        borderRadius: 12,
        boxShadow: '0 1px 4px rgba(0,0,0,0.08)',
        overflow: 'hidden'
      }}>
        <div style={{ padding: '20px 24px', borderBottom: '1px solid #f1f5f9', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2 style={{ fontSize: 18 }}>Job Logs</h2>
          <span style={{ fontSize: 12, color: '#94a3b8' }}></span>
        </div>
        <table>
          <thead>
            <tr style={{ background: '#f8fafc' }}>
              {['Job ID', 'Time', 'Result', 'Status'].map(h => (
                <th key={h} style={{ padding: '12px 16px', fontSize: 12, color: '#64748b', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.05em' }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {logs.length === 0 ? (
              <tr>
                <td colSpan="4" style={{ textAlign: 'center', padding: 32, color: '#94a3b8' }}>No logs yet</td>
              </tr>
            ) : (
              logs.map(log => (
                <tr key={log.id} style={{ borderTop: '1px solid #f1f5f9' }}>
                  <td style={td}>{log.job_id}</td>
                  <td style={{ ...td, color: '#64748b', fontSize: 13 }}>{new Date(log.run_time).toLocaleString()}</td>
                  <td style={{ ...td, fontFamily: 'monospace', fontSize: 13, color: '#475569' }}>{log.result || '-'}</td>
                  <td style={td}>
                    <span style={{
                      ...statusColor(log.status),
                      padding: '3px 10px',
                      borderRadius: 20,
                      fontSize: 12,
                      fontWeight: 500
                    }}>
                      {log.status}
                    </span>
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

export default LogsTable