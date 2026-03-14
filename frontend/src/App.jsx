import { useState, useEffect } from 'react'
import Map from './components/Map'
import useGameSocket from './hooks/useGameSocket'
import { initSession, getSession } from './game/session'

function App() {
  const [healthStatus, setHealthStatus] = useState('checking...')
  const [session, setSession] = useState(getSession())
  const { connected, lastMessage, sendMessage } = useGameSocket(session?.sessionToken)

  useEffect(() => {
    fetch('/api/health')
      .then(res => res.json())
      .then(data => setHealthStatus(data.status === 'ok' ? 'connected' : 'error'))
      .catch(() => setHealthStatus('disconnected'))
  }, [])

  useEffect(() => {
    if (!session) {
      initSession()
        .then(setSession)
        .catch(err => console.error('Failed to init session:', err))
    }
  }, [session])

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
        {session && (
          <div className="text-xs text-gray-400 border-t border-gray-600 pt-1 mt-1">
            {session.name}
          </div>
        )}
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
