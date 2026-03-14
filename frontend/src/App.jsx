import { useState, useEffect, useCallback } from 'react'
import Map from './components/Map'
import useGameSocket from './hooks/useGameSocket'
import { initSession, getSession } from './game/session'

const BANGKOK = { lat: 13.7563, lng: 100.5018 }

function App() {
  const [healthStatus, setHealthStatus] = useState('checking...')
  const [session, setSession] = useState(getSession())
  const [zone, setZone] = useState(null)
  const [picking, setPicking] = useState(false)
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

  const loadZone = useCallback((lat, lng) => {
    if (!session) return
    console.log('[Zone] loading zone for lat=%f lng=%f', lat, lng)
    fetch(`/api/map/zones?lat=${lat}&lng=${lng}`, {
      headers: { Authorization: `Bearer ${session.sessionToken}` },
    })
      .then(res => res.json())
      .then(data => {
        console.log('[Zone] response:', data.h3_zone, 'monsters:', data.monsters?.length)
        setZone(data)
      })
      .catch(err => console.error('Failed to load zone:', err))
  }, [session])

  // Load default zone on session ready
  useEffect(() => {
    if (!session) return
    loadZone(BANGKOK.lat, BANGKOK.lng)
  }, [session, loadZone])

  const handleMapClick = useCallback((lat, lng) => {
    if (!picking) return
    setPicking(false)
    loadZone(lat, lng)
  }, [picking, loadZone])

  return (
    <div className="w-full h-full relative">
      <Map zone={zone} picking={picking} onMapClick={handleMapClick} />

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

      {/* Title + Select Zone button */}
      <div className="absolute top-4 left-4 z-[1000] bg-black/80 text-white px-4 py-2 rounded-lg">
        <h1 className="text-lg font-bold">HexSlayer</h1>
        <p className="text-xs text-gray-400">Idle Monster Hunting</p>
        <button
          onClick={() => setPicking(p => !p)}
          className={`mt-2 text-xs px-3 py-1 rounded font-mono ${
            picking
              ? 'bg-yellow-500 text-black'
              : 'bg-gray-700 text-white hover:bg-gray-600'
          }`}
        >
          {picking ? 'Click map to select zone...' : 'Select Zone'}
        </button>
      </div>
    </div>
  )
}

export default App
