function Header() {
  return (
    <header style={{
      background: '#1a1a2e',
      color: 'white',
      padding: '32px',
      borderBottom: '1px solid #e94560',
      textAlign: 'center'
    }}>
      <h1 style={{ margin: 0, fontSize: 36 }}>🧠 Job Scheduler</h1>
      <p style={{ margin: '8px 0 0', color: '#aaa', fontSize: 16 }}>
        Kafka · PostgreSQL · Go
      </p>
    </header>
  )
}

export default Header