import { useEffect, useRef } from 'react'

function CombatLog({ events = [] }) {
  const bottomRef = useRef(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [events.length])

  return (
    <div className="absolute bottom-4 left-4 z-[1000] bg-black/80 text-white p-3 rounded-lg w-80 max-h-48 overflow-y-auto text-xs font-mono">
      <h3 className="font-bold mb-2">Combat Log</h3>
      {events.length === 0 ? (
        <p className="text-gray-500">No combat events yet...</p>
      ) : (
        // Events are newest-first in state, render reversed so newest is at bottom
        [...events].reverse().map((e, i) => (
          <div key={i} className="py-0.5">
            {e.event === 'kill' ? (
              <span className="text-green-400">{e.attacker} killed a monster!</span>
            ) : e.event === 'death' ? (
              <span className="text-red-400">{e.killed_by || e.attacker} killed your character!</span>
            ) : (
              <>
                <span className="text-yellow-400">{e.attacker}</span>
                {' hit '}
                <span className="text-red-400">{e.defender}</span>
                {' for '}
                <span className={e.is_crit ? 'text-orange-400 font-bold' : ''}>
                  {e.damage}{e.is_crit ? ' CRIT!' : ''}
                </span>
              </>
            )}
          </div>
        ))
      )}
      <div ref={bottomRef} />
    </div>
  )
}

export default CombatLog
