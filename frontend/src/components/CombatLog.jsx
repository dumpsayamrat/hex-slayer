// TODO: Scrolling combat log panel
// Receives combat_log events from WebSocket

function CombatLog({ events = [] }) {
  return (
    <div className="absolute bottom-4 left-4 z-[1000] bg-black/80 text-white p-3 rounded-lg w-80 max-h-48 overflow-y-auto text-xs font-mono">
      <h3 className="font-bold mb-2">Combat Log</h3>
      {events.length === 0 ? (
        <p className="text-gray-500">No combat events yet...</p>
      ) : (
        events.map((e, i) => (
          <div key={i} className="py-0.5">
            <span className="text-yellow-400">{e.attacker}</span>
            {' → '}
            <span className="text-red-400">{e.defender}</span>
            {' '}
            <span className={e.is_crit ? 'text-orange-400 font-bold' : ''}>
              {e.damage} dmg{e.is_crit ? ' CRIT' : ''}
            </span>
          </div>
        ))
      )}
    </div>
  )
}

export default CombatLog
