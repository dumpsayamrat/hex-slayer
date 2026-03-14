function CharacterPanel({ characters = [], monsters = [], onDeploy }) {
  const monsterById = {}
  for (const m of monsters) {
    monsterById[m.id] = m
  }

  return (
    <div className="absolute top-20 right-4 z-[1000] bg-black/80 text-white p-3 rounded-lg w-64 text-sm">
      <h3 className="font-bold mb-2">Characters</h3>
      {[0, 1].map((slot) => {
        const char = characters[slot]
        if (!char) {
          return (
            <div key={slot} className="border border-gray-600 rounded p-2 mb-2">
              <p className="text-gray-500">Empty Slot</p>
              <button
                onClick={() => onDeploy?.()}
                className="mt-1 px-2 py-1 bg-green-700 rounded text-xs hover:bg-green-600"
              >
                Deploy
              </button>
            </div>
          )
        }

        const dead = char.hp <= 0
        const hpPct = char.max_hp > 0 ? (char.hp / char.max_hp) * 100 : 0
        const hpColor = dead ? 'bg-gray-500' : hpPct > 50 ? 'bg-green-500' : hpPct > 25 ? 'bg-yellow-500' : 'bg-red-500'
        const fightingMonster = char.fighting_monster_id ? monsterById[char.fighting_monster_id] : null

        return (
          <div key={slot} className={`border rounded p-2 mb-2 ${dead ? 'border-red-800 opacity-60' : 'border-gray-600'}`}>
            <div className="flex justify-between items-center">
              <p className="font-semibold">{char.name}</p>
              {dead && <span className="text-red-400 text-xs">DEAD</span>}
            </div>

            {/* Character HP */}
            <div className="w-full bg-gray-700 rounded h-2 mt-1">
              <div className={`${hpColor} h-2 rounded transition-all`} style={{ width: `${hpPct}%` }} />
            </div>
            <p className="text-xs text-gray-400 mt-0.5">
              HP: {char.hp}/{char.max_hp}
            </p>

            {/* Fighting status */}
            {fightingMonster && !dead && (
              <div className="mt-1 border-t border-gray-700 pt-1">
                <p className="text-xs text-orange-400">Fighting: {fightingMonster.type}</p>
                <div className="w-full bg-gray-700 rounded h-1.5 mt-0.5">
                  <div
                    className="bg-red-500 h-1.5 rounded transition-all"
                    style={{ width: `${fightingMonster.max_hp > 0 ? (fightingMonster.current_hp / fightingMonster.max_hp) * 100 : 0}%` }}
                  />
                </div>
                <p className="text-xs text-gray-500">
                  {fightingMonster.current_hp}/{fightingMonster.max_hp}
                </p>
              </div>
            )}

            {!fightingMonster && !dead && (
              <p className="text-xs text-gray-500 mt-1">Wandering...</p>
            )}
          </div>
        )
      })}
    </div>
  )
}

export default CharacterPanel
