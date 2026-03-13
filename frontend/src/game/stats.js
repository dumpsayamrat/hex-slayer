// Generate random player stats — called once on first visit, saved to localStorage

function randBetween(min, max) {
  return Math.round((Math.random() * (max - min) + min) * 100) / 100
}

export function generatePlayerStats() {
  return {
    base_damage: Math.floor(randBetween(80, 120)),
    damage_amp: randBetween(1.0, 1.5),
    damage_reduction: randBetween(0.05, 0.2),
    crit_chance: randBetween(0.1, 0.25),
    crit_multiplier: randBetween(1.5, 2.0),
    max_hp: Math.floor(randBetween(250, 350)),
  }
}

export function getOrCreatePlayerStats() {
  const stored = localStorage.getItem('playerStats')
  if (stored) return JSON.parse(stored)

  const stats = generatePlayerStats()
  localStorage.setItem('playerStats', JSON.stringify(stats))
  return stats
}
