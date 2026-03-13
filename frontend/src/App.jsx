import { useState, useEffect } from 'react'
import Map from './components/Map'
import useGameSocket from './hooks/useGameSocket'

function App() {
  const [healthStatus, setHealthStatus] = useState('checking...')
  const { connected, lastMessage } = useGameSocket()

  useEffect(() => {
    fetch('/api/health')
      .then(res => res.json())
      .then(data => setHealthStatus(data.status === 'ok' ? 'connected' : 'error'))
      .catch(() => setHealthStatus('disconnected'))
  }, [])

  return (
    <div className="w-full h-full relative">
      <Map />

      {/* Connection status overlay */}
      <div className="absolute top-4 right-4 z-[1000] bg-black/80 text-white px-4 py-2 rounded-lg text-sm font-mono space-y-1">
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${healthStatus === 'connected' ? 'bg-green-400' : healthStatus === 'checking...' ? 'bg-yellow-400' : 'bg-red-400'}`} />
          <span>API: {healthStatus}</span>
        </div>
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${connected ? 'bg-green-400' : 'bg-red-400'}`} />
          <span>WS: {connected ? 'connected' : 'disconnected'}</span>
        </div>
      </div>

      {/* Title */}
      <div className="absolute top-4 left-4 z-[1000] bg-black/80 text-white px-4 py-2 rounded-lg">
        <h1 className="text-lg font-bold">HexSlayer</h1>
        <p className="text-xs text-gray-400">Idle Monster Hunting</p>
      </div>
    </div>
  )
}

export default App
