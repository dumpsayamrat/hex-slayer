// Game state reducer — central store for monsters, characters, combat logs

export const initialState = {
  monsters: [],    // from zone API + WS updates
  characters: [],  // from zone API + WS updates
  combatLogs: [],  // from WS combat_log events (last 50)
}

const MAX_LOGS = 50

export function gameReducer(state, action) {
  switch (action.type) {
    // Set initial zone data from GET /api/map/zones
    case 'ZONE_LOADED': {
      return {
        ...state,
        monsters: action.monsters || [],
        characters: action.characters || [],
        combatLogs: [],
      }
    }

    // WS: zone_snapshot — merge character fighting status + engaged monster HP
    case 'ZONE_SNAPSHOT': {
      const { characters: snapChars, monsters: snapMonsters } = action

      // Update characters with fighting_monster_id from snapshot
      const charById = {}
      for (const sc of snapChars || []) {
        charById[sc.id] = sc
      }
      const characters = state.characters.map(c => {
        const snap = charById[c.id]
        if (snap) {
          return { ...c, hp: snap.hp, max_hp: snap.max_hp, fighting_monster_id: snap.fighting_monster_id }
        }
        return c
      })
      // Add any chars from snapshot not in current state
      for (const sc of snapChars || []) {
        if (!characters.find(c => c.id === sc.id)) {
          characters.push(sc)
        }
      }

      // Update engaged monster HP
      const monsterHpMap = {}
      for (const sm of snapMonsters || []) {
        monsterHpMap[sm.id] = sm.current_hp
      }
      const monsters = state.monsters.map(m => {
        if (monsterHpMap[m.id] !== undefined) {
          return { ...m, current_hp: monsterHpMap[m.id] }
        }
        return m
      })

      return { ...state, characters, monsters }
    }

    // WS: combat_log
    case 'COMBAT_LOG': {
      const log = { attacker: action.attacker, defender: action.defender, damage: action.damage, is_crit: action.is_crit }
      const combatLogs = [log, ...state.combatLogs].slice(0, MAX_LOGS)
      return { ...state, combatLogs }
    }

    // WS: combat_engage
    case 'COMBAT_ENGAGE': {
      const characters = state.characters.map(c =>
        c.id === action.character_id ? { ...c, fighting_monster_id: action.monster_id } : c
      )
      return { ...state, characters }
    }

    // WS: char_move
    case 'CHAR_MOVE': {
      const characters = state.characters.map(c =>
        c.id === action.character_id ? { ...c, h3_index: action.h3_index } : c
      )
      return { ...state, characters }
    }

    // WS: monster_died
    case 'MONSTER_DIED': {
      const monsters = state.monsters.map(m =>
        m.id === action.monster_id ? { ...m, is_alive: false, current_hp: 0 } : m
      )
      // Clear engagement for any character fighting this monster
      const characters = state.characters.map(c =>
        c.fighting_monster_id === action.monster_id ? { ...c, fighting_monster_id: null } : c
      )
      const log = { attacker: action.killed_by, defender: '(dead)', damage: 0, is_crit: false, event: 'kill' }
      const combatLogs = [log, ...state.combatLogs].slice(0, MAX_LOGS)
      return { ...state, monsters, characters, combatLogs }
    }

    // WS: character_died
    case 'CHARACTER_DIED': {
      const characters = state.characters.filter(c => c.id !== action.character_id)
      const log = { attacker: action.killed_by, defender: '(char died)', damage: 0, is_crit: false, event: 'death' }
      const combatLogs = [log, ...state.combatLogs].slice(0, MAX_LOGS)
      return { ...state, characters, combatLogs }
    }

    // Add newly deployed character
    case 'CHAR_DEPLOYED': {
      return { ...state, characters: [...state.characters, action.character] }
    }

    default:
      return state
  }
}
