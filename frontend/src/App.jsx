import { useState, useEffect, useCallback, useReducer } from 'react'
import Map from './components/Map'
import CharacterPanel from './components/CharacterPanel'
import CombatLog from './components/CombatLog'
import useGameSocket from './hooks/useGameSocket'
import { initSession } from './game/session'
import { gameReducer, initialState } from './game/state'

const BANGKOK = { lat: 13.7563, lng: 100.5018 }

function App() {
  const [healthStatus, setHealthStatus] = useState('checking...')
  const [session, setSession] = useState(null)
  const [zoneId, setZoneId] = useState(null)
  const [picking, setPicking] = useState(false)
  const [game, dispatch] = useReducer(gameReducer, initialState)
  // Handle WS messages — dispatch directly to avoid dropped events
  const handleWsMessage = useCallback((msg) => {
    switch (msg.type) {
      case 'zone_snapshot':
        console.log(`%c[WS] zone_snapshot %c${msg.characters?.length || 0} chars, ${msg.monsters?.length || 0} monsters`, 'color:#4ade80;font-weight:bold', 'color:#9ca3af')
        dispatch({ type: 'ZONE_SNAPSHOT', characters: msg.characters, monsters: msg.monsters })
        break
      case 'combat_log':
        console.log(`%c[WS] combat_log %c${msg.attacker} → ${msg.defender} %c${msg.damage} dmg${msg.is_crit ? ' CRIT!' : ''} %c(char HP: ${msg.character_hp}, mon HP: ${msg.monster_hp})`, 'color:#f59e0b;font-weight:bold', 'color:#fbbf24', msg.is_crit ? 'color:#ef4444;font-weight:bold' : 'color:#fb923c', 'color:#9ca3af')
        dispatch({ ...msg, type: 'COMBAT_LOG' })
        break
      case 'combat_engage':
        console.log(`%c[WS] combat_engage %cchar:${msg.character_id.slice(0,8)} → mon:${msg.monster_id.slice(0,8)}`, 'color:#a78bfa;font-weight:bold', 'color:#c4b5fd')
        dispatch({ type: 'COMBAT_ENGAGE', character_id: msg.character_id, monster_id: msg.monster_id })
        break
      case 'char_move':
        console.log(`%c[WS] char_move %c${msg.character_id.slice(0,8)} → ${msg.h3_index}`, 'color:#60a5fa;font-weight:bold', 'color:#93c5fd')
        dispatch({ type: 'CHAR_MOVE', character_id: msg.character_id, h3_index: msg.h3_index })
        break
      case 'monster_died':
        console.log(`%c[WS] monster_died %c${msg.monster_id.slice(0,8)} killed by ${msg.killed_by}`, 'color:#ef4444;font-weight:bold', 'color:#fca5a5')
        dispatch({ type: 'MONSTER_DIED', monster_id: msg.monster_id, killed_by: msg.killed_by })
        break
      case 'character_died':
        console.log(`%c[WS] character_died %c${msg.character_id.slice(0,8)} killed by ${msg.killed_by}`, 'color:#dc2626;font-weight:bold', 'color:#fca5a5')
        dispatch({ type: 'CHARACTER_DIED', character_id: msg.character_id, killed_by: msg.killed_by })
        break
    }
  }, [])

  const { connected, sendMessage } = useGameSocket(session?.sessionToken, handleWsMessage)

  // Health check
  useEffect(() => {
    fetch('/api/health')
      .then(res => res.json())
      .then(data => setHealthStatus(data.status === 'ok' ? 'connected' : 'error'))
      .catch(() => setHealthStatus('disconnected'))
  }, [])

  // Init session (always validate — handles DB reset)
  useEffect(() => {
    initSession()
      .then(setSession)
      .catch(err => console.error('Failed to init session:', err))
  }, [])

  // Load zone from API
  const loadZone = useCallback((lat, lng) => {
    if (!session) return
    fetch(`/api/map/zones?lat=${lat}&lng=${lng}`, {
      headers: { Authorization: `Bearer ${session.sessionToken}` },
    })
      .then(res => res.json())
      .then(data => {
        console.log('[Zone] loaded:', data.h3_zone, 'monsters:', data.monsters?.length)
        setZoneId(data.h3_zone)
        dispatch({ type: 'ZONE_LOADED', monsters: data.monsters, characters: data.characters })
      })
      .catch(err => console.error('Failed to load zone:', err))
  }, [session])

  // Load default zone
  useEffect(() => {
    if (!session) return
    loadZone(BANGKOK.lat, BANGKOK.lng)
  }, [session, loadZone])

  // Subscribe to zone via WS when zone changes
  useEffect(() => {
    if (!zoneId || !connected) return
    sendMessage({ type: 'subscribe_zone', h3_zone: zoneId })
  }, [zoneId, connected, sendMessage])

  // Deploy character
  const handleDeploy = useCallback(() => {
    if (!session || !zoneId) return
    fetch('/api/character/deploy', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${session.sessionToken}`,
      },
      body: JSON.stringify({ h3_zone: zoneId }),
    })
      .then(res => {
        if (!res.ok) return res.json().then(d => { throw new Error(d.error) })
        return res.json()
      })
      .then(char => {
        console.log('[Deploy] character:', char.name)
        dispatch({ type: 'CHAR_DEPLOYED', character: { ...char, player_id: session.playerId } })
      })
      .catch(err => console.error('Deploy failed:', err.message))
  }, [session, zoneId])

  // Map click to pick zone
  const handleMapClick = useCallback((lat, lng) => {
    if (!picking) return
    setPicking(false)
    loadZone(lat, lng)
  }, [picking, loadZone])

  // My characters (alive ones belonging to current player)
  const myChars = game.characters.filter(c => c.player_id === session?.playerId)

  return (
    <div className="w-full h-full relative">
      <Map
        zoneId={zoneId}
        monsters={game.monsters}
        characters={game.characters}
        picking={picking}
        onMapClick={handleMapClick}
      />

      {/* Connection status */}
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

      {/* Title + Select Zone */}
      <div className="absolute top-4 left-4 z-[1000] bg-black/80 text-white px-4 py-2 rounded-lg">
        <h1 className="text-lg font-bold">HexSlayer</h1>
        <p className="text-xs text-gray-400">Idle Monster Hunting</p>
        <button
          onClick={() => setPicking(p => !p)}
          className={`mt-2 text-xs px-3 py-1 rounded font-mono ${
            picking ? 'bg-yellow-500 text-black' : 'bg-gray-700 text-white hover:bg-gray-600'
          }`}
        >
          {picking ? 'Click map to select zone...' : 'Select Zone'}
        </button>
      </div>

      {/* Character Panel */}
      <CharacterPanel
        characters={myChars}
        monsters={game.monsters}
        damagePopups={game.damagePopups}
        dispatch={dispatch}
        onDeploy={handleDeploy}
      />

      {/* Combat Log */}
      <CombatLog events={game.combatLogs} />
    </div>
  )
}

export default App
