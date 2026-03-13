// TODO: Character slots panel (max 2 characters)
// Shows active characters with HP bars, deploy button for empty slots

function CharacterPanel({ characters = [], onDeploy }) {
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
                onClick={() => onDeploy?.(slot)}
                className="mt-1 px-2 py-1 bg-green-700 rounded text-xs hover:bg-green-600"
              >
                Deploy
              </button>
            </div>
          )
        }
        const hpPct = (char.hp / char.max_hp) * 100
        return (
          <div key={slot} className="border border-gray-600 rounded p-2 mb-2">
            <p className="font-semibold">{char.name}</p>
            <div className="w-full bg-gray-700 rounded h-2 mt-1">
              <div
                className="bg-green-500 h-2 rounded"
                style={{ width: `${hpPct}%` }}
              />
            </div>
            <p className="text-xs text-gray-400 mt-1">
              HP: {char.hp}/{char.max_hp}
            </p>
          </div>
        )
      })}
    </div>
  )
}

export default CharacterPanel
