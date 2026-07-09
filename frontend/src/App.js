import { useState, useEffect } from 'react'

import Header from './components/Header'
import CreateJobForm from './components/CreateJobForm'
import JobsTable from './components/JobsTable'
import LogsTable from './components/LogsTable'

function App() {
  // State for data
  const [jobs, setJobs] = useState([])
  const [logs, setLogs] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // State for the create job form
  const [formData, setFormData] = useState({
    name: '',
    command: '',
    schedule: '',
    status: 'active'
  })

  // Fetch jobs and logs when the page loads
  useEffect(() => {
  fetchJobs()
  fetchLogs()

  // Auto-refresh every 5 seconds
  const interval = setInterval(() => {
    fetchJobs()
    fetchLogs()
  }, 5000)

  // Cleanup: stop the interval when the component unmounts
  return () => clearInterval(interval)
}, [])

  const fetchJobs = () => {
    fetch('/api/jobs/all')
      .then(res => res.json())
      .then(data => setJobs(data || []))
      .catch(err => setError('Failed to fetch jobs'))
  }

  const fetchLogs = () => {
    fetch('/api/job_logs')
      .then(res => res.json())
      .then(data => setLogs(data || []))
      .catch(err => setError('Failed to fetch logs'))
  }

  return (
    <div>
      <Header />
      <CreateJobForm onJobCreated={fetchJobs} />
      <JobsTable jobs={jobs} onRefresh={fetchJobs} />
      <LogsTable logs={logs} onRefresh={fetchLogs} />
    </div>
  )
}

export default App