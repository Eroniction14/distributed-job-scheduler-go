import { useState } from 'react'

function CreateJobForm({ onJobCreated }) {
  const [formData, setFormData] = useState({
    name: '', command: '', schedule: '', status: 'active'
  })

  const handleSubmit = (e) => {
    e.preventDefault()
    fetch('/api/jobs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(formData)
    })
      .then(res => res.json())
      .then(() => {
        onJobCreated()
        setFormData({ name: '', command: '', schedule: '', status: 'active' })
      })
  }

  return (
    <div style={{ padding: '24px 32px' }}>
      <div style={{
        background: 'white',
        borderRadius: 12,
        padding: 24,
        boxShadow: '0 1px 4px rgba(0,0,0,0.08)'
      }}>
        <h2 style={{ marginBottom: 20, fontSize: 18 }}>Create Job</h2>
        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <input
              placeholder="Job Name"
              value={formData.name}
              onChange={(e) => setFormData({...formData, name: e.target.value})}
              style={inputStyle}
              required
            />
            <input
              placeholder="Command (e.g. echo hello)"
              value={formData.command}
              onChange={(e) => setFormData({...formData, command: e.target.value})}
              style={inputStyle}
              required
            />
            <input
              placeholder="Cron Schedule (e.g. * * * * *)"
              value={formData.schedule}
              onChange={(e) => setFormData({...formData, schedule: e.target.value})}
              style={inputStyle}
              required
            />
            <select
              value={formData.status}
              onChange={(e) => setFormData({...formData, status: e.target.value})}
              style={inputStyle}
            >
              <option value="active">Active</option>
              <option value="paused">Paused</option>
            </select>
          </div>
          <button type="submit" style={{
            background: '#e94560',
            color: 'white',
            border: 'none',
            borderRadius: 8,
            padding: '10px 24px',
            fontSize: 14,
            fontWeight: 600,
            alignSelf: 'flex-start'
          }}>
            Create Job
          </button>
        </form>
      </div>
    </div>
  )
}

const inputStyle = {
  border: '1px solid #e2e8f0',
  borderRadius: 8,
  padding: '10px 14px',
  fontSize: 14,
  outline: 'none',
  width: '100%'
}

export default CreateJobForm